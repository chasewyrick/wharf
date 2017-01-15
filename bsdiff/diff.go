package bsdiff

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"runtime"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/golang/protobuf/proto"
	"github.com/itchio/wharf/state"
	"github.com/jgallagher/gosaca"
)

// MaxFileSize is the largest size bsdiff will diff (for both old and new file): 2GB - 1 bytes
// a different codepath could be used for larger files, at the cost of unreasonable memory usage
// (even in 2016). If a big corporate user is willing to sponsor that part of the code, get in touch!
// Fair warning though: it won't be covered, our CI workers don't have that much RAM :)
const MaxFileSize = int64(math.MaxInt32 - 1)

// MaxMessageSize is the maximum amount of bytes that will be stored
// in a protobuf message generated by bsdiff. This enable friendlier streaming apply
// at a small storage cost
// TODO: actually use
const MaxMessageSize int64 = 16 * 1024 * 1024

type DiffStats struct {
	TimeSpentSorting  time.Duration
	TimeSpentScanning time.Duration
	BiggestAdd        int64
}

// DiffContext holds settings for the diff process, along with some
// internal storage: re-using a diff context is good to avoid GC thrashing
// (but never do it concurrently!)
type DiffContext struct {
	// SuffixSortConcurrency specifies the number of workers to use for suffix sorting.
	// Exceeding the number of cores will only slow it down. A 0 value (default) uses
	// sequential suffix sorting, which uses less RAM and has less overhead (might be faster
	// in some scenarios). A negative value means (number of cores - value).
	SuffixSortConcurrency int

	// number of partitions into which to separate the input data, sort concurrently
	// and scan in concurrently
	Partitions int

	// MeasureMem enables printing memory usage statistics at various points in the
	// diffing process.
	MeasureMem bool

	// MeasureParallelOverhead prints some stats on the overhead of parallel suffix sorting
	MeasureParallelOverhead bool

	Stats *DiffStats

	db bytes.Buffer
	ws *gosaca.WorkSpace

	obuf bytes.Buffer
	nbuf bytes.Buffer
}

// WriteMessageFunc should write a given protobuf message and relay any errors
// No reference to the given message can be kept, as its content may be modified
// after WriteMessageFunc returns. See the `wire` package for an example implementation.
type WriteMessageFunc func(msg proto.Message) (err error)

// Do computes the difference between old and new, according to the bsdiff
// algorithm, and writes the result to patch.
func (ctx *DiffContext) Do(old, new io.Reader, writeMessage WriteMessageFunc, consumer *state.Consumer) error {
	var memstats *runtime.MemStats
	var err error

	if ctx.MeasureMem {
		memstats = &runtime.MemStats{}
		runtime.ReadMemStats(memstats)
		consumer.Debugf("Allocated bytes at start of bsdiff: %s (%s total)", humanize.IBytes(uint64(memstats.Alloc)), humanize.IBytes(uint64(memstats.TotalAlloc)))
	}

	ctx.obuf.Reset()
	_, err = io.Copy(&ctx.obuf, old)
	if err != nil {
		return err
	}

	obuf := ctx.obuf.Bytes()
	obuflen := ctx.obuf.Len()

	ctx.nbuf.Reset()
	_, err = io.Copy(&ctx.nbuf, new)
	if err != nil {
		return err
	}

	nbuf := ctx.nbuf.Bytes()
	nbuflen := ctx.nbuf.Len()

	if ctx.MeasureMem {
		runtime.ReadMemStats(memstats)
		consumer.Debugf("Allocated bytes after ReadAll: %s (%s total)", humanize.IBytes(uint64(memstats.Alloc)), humanize.IBytes(uint64(memstats.TotalAlloc)))
	}

	if ctx.Partitions > 0 {
		return ctx.doPartitioned(obuf, obuflen, nbuf, nbuflen, memstats, writeMessage, consumer)
	}

	var lenf int
	startTime := time.Now()

	I := make([]int, obuflen+1)
	if ctx.ws == nil {
		ctx.ws = &gosaca.WorkSpace{}
	}

	if obuflen > 0 {
		ctx.ws.ComputeSuffixArray(obuf, I[:obuflen])
	}

	if ctx.Stats != nil {
		ctx.Stats.TimeSpentSorting += time.Since(startTime)
	}

	if ctx.MeasureMem {
		runtime.ReadMemStats(memstats)
		consumer.Debugf("Allocated bytes after qsufsort: %s (%s total)", humanize.IBytes(uint64(memstats.Alloc)), humanize.IBytes(uint64(memstats.TotalAlloc)))
	}

	bsdc := &Control{}

	consumer.ProgressLabel(fmt.Sprintf("Scanning %s...", humanize.IBytes(uint64(nbuflen))))

	var lastProgressUpdate int
	var updateEvery = 64 * 1024 * 1046 // 64MB

	startTime = time.Now()

	// Compute the differences, writing ctrl as we go
	var scan, pos, length int
	var lastscan, lastpos, lastoffset int
	for scan < nbuflen {
		var oldscore int
		scan += length

		if scan-lastProgressUpdate > updateEvery {
			lastProgressUpdate = scan
			progress := float64(scan) / float64(nbuflen)
			consumer.Progress(progress)
		}

		for scsc := scan; scan < nbuflen; scan++ {
			pos, length = search(I, obuf, nbuf[scan:], 0, obuflen)

			for ; scsc < scan+length; scsc++ {
				if scsc+lastoffset < obuflen &&
					obuf[scsc+lastoffset] == nbuf[scsc] {
					oldscore++
				}
			}

			if (length == oldscore && length != 0) || length > oldscore+8 {
				break
			}

			if scan+lastoffset < obuflen && obuf[scan+lastoffset] == nbuf[scan] {
				oldscore--
			}
		}

		if length != oldscore || scan == nbuflen {
			var s, Sf int
			lenf = 0
			for i := int(0); lastscan+i < scan && lastpos+i < obuflen; {
				if obuf[lastpos+i] == nbuf[lastscan+i] {
					s++
				}
				i++
				if s*2-i > Sf*2-lenf {
					Sf = s
					lenf = i
				}
			}

			lenb := 0
			if scan < nbuflen {
				var s, Sb int
				for i := int(1); (scan >= lastscan+i) && (pos >= i); i++ {
					if obuf[pos-i] == nbuf[scan-i] {
						s++
					}
					if s*2-i > Sb*2-lenb {
						Sb = s
						lenb = i
					}
				}
			}

			if lastscan+lenf > scan-lenb {
				overlap := (lastscan + lenf) - (scan - lenb)
				s := int(0)
				Ss := int(0)
				lens := int(0)
				for i := int(0); i < overlap; i++ {
					if nbuf[lastscan+lenf-overlap+i] == obuf[lastpos+lenf-overlap+i] {
						s++
					}
					if nbuf[scan-lenb+i] == obuf[pos-lenb+i] {
						s--
					}
					if s > Ss {
						Ss = s
						lens = i + 1
					}
				}

				lenf += lens - overlap
				lenb -= lens
			}

			ctx.db.Reset()
			ctx.db.Grow(int(lenf))

			for i := int(0); i < lenf; i++ {
				ctx.db.WriteByte(nbuf[lastscan+i] - obuf[lastpos+i])
			}

			bsdc.Add = ctx.db.Bytes()
			bsdc.Copy = nbuf[(lastscan + lenf):(scan - lenb)]
			bsdc.Seek = int64((pos - lenb) - (lastpos + lenf))

			err := writeMessage(bsdc)
			if err != nil {
				return err
			}

			if ctx.Stats != nil && ctx.Stats.BiggestAdd < int64(lenf) {
				ctx.Stats.BiggestAdd = int64(lenf)
			}

			lastscan = scan - lenb
			lastpos = pos - lenb
			lastoffset = pos - scan
		}
	}

	if ctx.Stats != nil {
		ctx.Stats.TimeSpentScanning += time.Since(startTime)
	}

	if ctx.MeasureMem {
		runtime.ReadMemStats(memstats)
		consumer.Debugf("Allocated bytes after scan: %s (%s total)", humanize.IBytes(uint64(memstats.Alloc)), humanize.IBytes(uint64(memstats.TotalAlloc)))
	}

	bsdc.Reset()
	bsdc.Eof = true
	err = writeMessage(bsdc)
	if err != nil {
		return err
	}

	return nil
}

func (ctx *DiffContext) doPartitioned(obuf []byte, obuflen int, nbuf []byte, nbuflen int, memstats *runtime.MemStats, writeMessage WriteMessageFunc, consumer *state.Consumer) error {
	var err error

	var lenf int
	startTime := time.Now()

	psa := NewPSA(ctx.Partitions, obuf)

	if ctx.Stats != nil {
		ctx.Stats.TimeSpentSorting += time.Since(startTime)
	}

	if ctx.MeasureMem {
		runtime.ReadMemStats(memstats)
		consumer.Debugf("Allocated bytes after qsufsort: %s (%s total)", humanize.IBytes(uint64(memstats.Alloc)), humanize.IBytes(uint64(memstats.TotalAlloc)))
	}

	bsdc := &Control{}

	consumer.ProgressLabel(fmt.Sprintf("Scanning %s...", humanize.IBytes(uint64(nbuflen))))

	var lastProgressUpdate int
	var updateEvery = 64 * 1024 * 1046 // 64MB

	startTime = time.Now()

	// Compute the differences, writing ctrl as we go
	var scan, pos, length int
	var lastscan, lastpos, lastoffset int

	for scan < nbuflen {
		// fmt.Printf("\n")
		// fmt.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
		// fmt.Printf(">>>> main loop: lastscan %d, scan %d, lastpos %d, pos %d, lastoffset %d, length %d\n", lastscan, scan, lastpos, pos, lastoffset, length)
		// fmt.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n")
		// fmt.Printf("\n")

		var oldscore int
		scan += length

		if scan-lastProgressUpdate > updateEvery {
			lastProgressUpdate = scan
			progress := float64(scan) / float64(nbuflen)
			consumer.Progress(progress)
		}

		for scsc := scan; scan < nbuflen; scan++ {
			pos, length = psa.search(nbuf[scan:])

			// fmt.Printf("\n@ scanning from %d, found %d common bytes at %d\n", scan, length, pos)

			// scan+length is the end of the common sequence we just found
			// if length > 0 {
			// fmt.Printf("@ found common seq of length %d, scoring %d bytes now\n", length, scan+length-scsc)
			// }

			for ; scsc < scan+length; scsc++ {
				if scsc+lastoffset < obuflen &&
					obuf[scsc+lastoffset] == nbuf[scsc] {
					oldscore++
				}
			}

			// fmt.Printf("@ current candidate seq has %d/%d matching bytes\n", oldscore, scan-lastscan+length)

			if (length == oldscore && length != 0) || length > oldscore+8 {
				// if length == oldscore && length != 0 {
				// 	fmt.Printf("@ perfect non-empty common sequence of %d bytes!\n", length)
				// }
				// if length > oldscore+8 {
				// 	fmt.Printf("@ new common sequence is 8 bytes better than previous one\n")
				// }
				break
			}

			if scan+lastoffset < obuflen && obuf[scan+lastoffset] == nbuf[scan] {
				oldscore--
			}
		}

		// if length != oldscore {
		// 	fmt.Printf("@ score %d != length %d, need to compute extensions\n", oldscore, length)
		// } else if scan == nbuflen {
		// 	fmt.Printf("@ scanned nbuf entirely, it doesn't get any better\n")
		// } else {
		// 	fmt.Printf("@ score is length & not scanned entire nbuf, scanning some more\n")
		// }

		if length != oldscore || scan == nbuflen {
			// fmt.Printf("@ considered sequence is [%d...%d]\n", lastscan, scan)
			// fmt.Printf("@ in obuf:               [%d...%d]\n", lastpos, pos)

			var s, Sf int
			lenf = 0
			for i := int(0); lastscan+i < scan && lastpos+i < obuflen; {
				if obuf[lastpos+i] == nbuf[lastscan+i] {
					// fmt.Printf("@ matched 1 forward (at %d)\n", lastscan+i)
					s++
				}
				i++
				if s*2-i > Sf*2-lenf {
					// fmt.Printf("@ now extending %d forward\n", s)
					Sf = s
					lenf = i
				}
			}

			lenb := 0
			if scan < nbuflen {
				var s, Sb int
				for i := int(1); (scan >= lastscan+i) && (pos >= i); i++ {
					if obuf[pos-i] == nbuf[scan-i] {
						// fmt.Printf("@ matched 1 backward (at %d)\n", scan-i)
						s++
					}
					if s*2-i > Sb*2-lenb {
						// fmt.Printf("@ now extending %d backward\n", s)
						Sb = s
						lenb = i
					}
				}
			}
			// fmt.Printf("@ original: [%d...%d]\n", lastscan, scan)
			// fmt.Printf("@ forw-ext: [%d...%d]\n", lastscan, lastscan+lenf)
			// fmt.Printf("@ back-ext: [%d...%d]\n", scan-lenb, scan)

			if lastscan+lenf > scan-lenb {
				overlap := (lastscan + lenf) - (scan - lenb)
				// fmt.Printf("@ forward and backwards overlap by %d bytes\n", overlap)
				s := int(0)
				Ss := int(0)
				lens := int(0)
				for i := int(0); i < overlap; i++ {
					if nbuf[lastscan+lenf-overlap+i] == obuf[lastpos+lenf-overlap+i] {
						s++
					}
					if nbuf[scan-lenb+i] == obuf[pos-lenb+i] {
						s--
					}
					if s > Ss {
						Ss = s
						lens = i + 1
					}
				}

				lenf += lens - overlap
				lenb -= lens
			}

			ctx.db.Reset()
			ctx.db.Grow(int(lenf))

			for i := int(0); i < lenf; i++ {
				ctx.db.WriteByte(nbuf[lastscan+i] - obuf[lastpos+i])
			}

			// fmt.Printf("@ storing %d bytes of add: %#v\n", lenf, ctx.db.Bytes())
			// fmt.Printf("@ storing %d bytes of copy\n", (scan-lenb)-(lastscan+lenf))
			// fmt.Printf("@ seek: %d\n", (pos-lenb)-(lastpos+lenf))

			bsdc.Add = ctx.db.Bytes()
			bsdc.Copy = nbuf[(lastscan + lenf):(scan - lenb)]
			bsdc.Seek = int64((pos - lenb) - (lastpos + lenf))

			err := writeMessage(bsdc)
			if err != nil {
				return err
			}

			if ctx.Stats != nil && ctx.Stats.BiggestAdd < int64(lenf) {
				ctx.Stats.BiggestAdd = int64(lenf)
			}

			lastscan = scan - lenb
			lastpos = pos - lenb
			lastoffset = pos - scan
		}
	}

	if ctx.Stats != nil {
		ctx.Stats.TimeSpentScanning += time.Since(startTime)
	}

	if ctx.MeasureMem {
		runtime.ReadMemStats(memstats)
		consumer.Debugf("Allocated bytes after scan: %s (%s total)", humanize.IBytes(uint64(memstats.Alloc)), humanize.IBytes(uint64(memstats.TotalAlloc)))
	}

	bsdc.Reset()
	bsdc.Eof = true
	err = writeMessage(bsdc)
	if err != nil {
		return err
	}

	return nil
}
