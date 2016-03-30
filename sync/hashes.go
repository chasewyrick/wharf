package sync

import (
	"bufio"
	"bytes"
	"io"
)

func (ctx *Context) splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// still have more data than blockSize ? return a block-full
	if len(data) >= ctx.blockSize {
		return ctx.blockSize, data[:ctx.blockSize], nil
	}

	if atEOF {
		// at eof, but still have data: return all of it (must be <= ctx.blockSize)
		if len(data) > 0 {
			return len(data), data, nil
		}

		// at eof, no data left, signal EOF ourselves.
		return 0, nil, io.EOF
	}

	// wait for more data
	return 0, nil, nil
}

// CreateSignature calculate the signature of target.
func (ctx *Context) CreateSignature(fileIndex int64, fileReader io.Reader, writeHash SignatureWriter) error {
	s := bufio.NewScanner(fileReader)
	s.Split(ctx.splitFunc)

	blockIndex := int64(0)

	for s.Scan() {
		block := s.Bytes()

		weakHash, _, _ := βhash(block)
		strongHash := ctx.uniqueHash(block)

		blockHash := BlockHash{
			FileIndex:  fileIndex,
			BlockIndex: blockIndex,
			WeakHash:   weakHash,
			StrongHash: strongHash,
		}

		if len(block) < ctx.blockSize {
			blockHash.ShortSize = int32(len(block))
		}

		err := writeHash(blockHash)
		if err != nil {
			return err
		}
		blockIndex++
	}

	return nil
}

// Use a more unique way to identify a set of bytes.
func (ctx *Context) uniqueHash(v []byte) []byte {
	ctx.uniqueHasher.Reset()
	_, err := ctx.uniqueHasher.Write(v)
	if err != nil {
		panic(err)
	}
	return ctx.uniqueHasher.Sum(nil)
}

// Searches for a given strong hash among all strong hashes in this bucket.
func findUniqueHash(hh []BlockHash, hashValue []byte, shortSize int32, preferredFileIndex int64) *BlockHash {
	if len(hashValue) == 0 {
		return nil
	}

	// try to find block in preferred file first
	// this helps detect files that aren't touched by patches
	if preferredFileIndex != -1 {
		for _, block := range hh {
			if block.FileIndex == preferredFileIndex {
				if block.ShortSize == shortSize && bytes.Equal(block.StrongHash, hashValue) {
					return &block
				}
			}
		}
	}

	for _, block := range hh {
		// full blocks have 0 shortSize
		if block.ShortSize == shortSize && bytes.Equal(block.StrongHash, hashValue) {
			return &block
		}
	}
	return nil
}

// βhash implements the rolling hash when signing an entire block at a time
func βhash(block []byte) (β uint32, β1 uint32, β2 uint32) {
	var a, b uint32
	for i, val := range block {
		a += uint32(val)
		b += (uint32(len(block)-1) - uint32(i) + 1) * uint32(val)
	}
	β = (a % _M) + (_M * (b % _M))
	β1 = a % _M
	β2 = b % _M
	return
}
