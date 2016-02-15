// Code generated by protoc-gen-go.
// source: pwr/pwr.proto
// DO NOT EDIT!

/*
Package pwr is a generated protocol buffer package.

It is generated from these files:
	pwr/pwr.proto

It has these top-level messages:
	PatchHeader
	SyncHeader
	SyncOp
	SignatureHeader
	BlockHash
	CompressionSettings
*/
package pwr

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.ProtoPackageIsVersion1

type CompressionAlgorithm int32

const (
	CompressionAlgorithm_UNCOMPRESSED CompressionAlgorithm = 0
	CompressionAlgorithm_BROTLI       CompressionAlgorithm = 1
)

var CompressionAlgorithm_name = map[int32]string{
	0: "UNCOMPRESSED",
	1: "BROTLI",
}
var CompressionAlgorithm_value = map[string]int32{
	"UNCOMPRESSED": 0,
	"BROTLI":       1,
}

func (x CompressionAlgorithm) String() string {
	return proto.EnumName(CompressionAlgorithm_name, int32(x))
}
func (CompressionAlgorithm) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type SyncOp_Type int32

const (
	SyncOp_BLOCK          SyncOp_Type = 0
	SyncOp_BLOCK_RANGE    SyncOp_Type = 1
	SyncOp_DATA           SyncOp_Type = 2
	SyncOp_HEY_YOU_DID_IT SyncOp_Type = 2049
)

var SyncOp_Type_name = map[int32]string{
	0:    "BLOCK",
	1:    "BLOCK_RANGE",
	2:    "DATA",
	2049: "HEY_YOU_DID_IT",
}
var SyncOp_Type_value = map[string]int32{
	"BLOCK":          0,
	"BLOCK_RANGE":    1,
	"DATA":           2,
	"HEY_YOU_DID_IT": 2049,
}

func (x SyncOp_Type) String() string {
	return proto.EnumName(SyncOp_Type_name, int32(x))
}
func (SyncOp_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2, 0} }

type PatchHeader struct {
	Compression *CompressionSettings `protobuf:"bytes,1,opt,name=compression" json:"compression,omitempty"`
}

func (m *PatchHeader) Reset()                    { *m = PatchHeader{} }
func (m *PatchHeader) String() string            { return proto.CompactTextString(m) }
func (*PatchHeader) ProtoMessage()               {}
func (*PatchHeader) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *PatchHeader) GetCompression() *CompressionSettings {
	if m != nil {
		return m.Compression
	}
	return nil
}

type SyncHeader struct {
	FileIndex int64 `protobuf:"varint,16,opt,name=fileIndex" json:"fileIndex,omitempty"`
}

func (m *SyncHeader) Reset()                    { *m = SyncHeader{} }
func (m *SyncHeader) String() string            { return proto.CompactTextString(m) }
func (*SyncHeader) ProtoMessage()               {}
func (*SyncHeader) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type SyncOp struct {
	Type       SyncOp_Type `protobuf:"varint,1,opt,name=type,enum=io.itch.wharf.pwr.SyncOp_Type" json:"type,omitempty"`
	FileIndex  int64       `protobuf:"varint,2,opt,name=fileIndex" json:"fileIndex,omitempty"`
	BlockIndex int64       `protobuf:"varint,3,opt,name=blockIndex" json:"blockIndex,omitempty"`
	BlockSpan  int64       `protobuf:"varint,4,opt,name=blockSpan" json:"blockSpan,omitempty"`
	Data       []byte      `protobuf:"bytes,5,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *SyncOp) Reset()                    { *m = SyncOp{} }
func (m *SyncOp) String() string            { return proto.CompactTextString(m) }
func (*SyncOp) ProtoMessage()               {}
func (*SyncOp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type SignatureHeader struct {
	Compression *CompressionSettings `protobuf:"bytes,1,opt,name=compression" json:"compression,omitempty"`
}

func (m *SignatureHeader) Reset()                    { *m = SignatureHeader{} }
func (m *SignatureHeader) String() string            { return proto.CompactTextString(m) }
func (*SignatureHeader) ProtoMessage()               {}
func (*SignatureHeader) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *SignatureHeader) GetCompression() *CompressionSettings {
	if m != nil {
		return m.Compression
	}
	return nil
}

type BlockHash struct {
	WeakHash   uint32 `protobuf:"varint,1,opt,name=weakHash" json:"weakHash,omitempty"`
	StrongHash []byte `protobuf:"bytes,2,opt,name=strongHash,proto3" json:"strongHash,omitempty"`
}

func (m *BlockHash) Reset()                    { *m = BlockHash{} }
func (m *BlockHash) String() string            { return proto.CompactTextString(m) }
func (*BlockHash) ProtoMessage()               {}
func (*BlockHash) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type CompressionSettings struct {
	Algorithm CompressionAlgorithm `protobuf:"varint,1,opt,name=algorithm,enum=io.itch.wharf.pwr.CompressionAlgorithm" json:"algorithm,omitempty"`
	Quality   int32                `protobuf:"varint,2,opt,name=quality" json:"quality,omitempty"`
}

func (m *CompressionSettings) Reset()                    { *m = CompressionSettings{} }
func (m *CompressionSettings) String() string            { return proto.CompactTextString(m) }
func (*CompressionSettings) ProtoMessage()               {}
func (*CompressionSettings) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func init() {
	proto.RegisterType((*PatchHeader)(nil), "io.itch.wharf.pwr.PatchHeader")
	proto.RegisterType((*SyncHeader)(nil), "io.itch.wharf.pwr.SyncHeader")
	proto.RegisterType((*SyncOp)(nil), "io.itch.wharf.pwr.SyncOp")
	proto.RegisterType((*SignatureHeader)(nil), "io.itch.wharf.pwr.SignatureHeader")
	proto.RegisterType((*BlockHash)(nil), "io.itch.wharf.pwr.BlockHash")
	proto.RegisterType((*CompressionSettings)(nil), "io.itch.wharf.pwr.CompressionSettings")
	proto.RegisterEnum("io.itch.wharf.pwr.CompressionAlgorithm", CompressionAlgorithm_name, CompressionAlgorithm_value)
	proto.RegisterEnum("io.itch.wharf.pwr.SyncOp_Type", SyncOp_Type_name, SyncOp_Type_value)
}

var fileDescriptor0 = []byte{
	// 420 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x52, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0xc6, 0x89, 0x13, 0x9a, 0x49, 0xda, 0x2e, 0x5b, 0x0e, 0x16, 0x42, 0x15, 0xf2, 0x01, 0x50,
	0x0f, 0x46, 0x0a, 0xbc, 0x80, 0x93, 0x58, 0xb5, 0x45, 0xa9, 0xd1, 0xda, 0x95, 0x28, 0x1c, 0xac,
	0xad, 0xb3, 0xb5, 0x57, 0xb8, 0x5e, 0xb3, 0xde, 0x12, 0x72, 0xe4, 0x89, 0x79, 0x05, 0xec, 0x4d,
	0x9b, 0x04, 0x88, 0x38, 0xf5, 0x60, 0x69, 0x66, 0xbe, 0x3f, 0xcf, 0xd8, 0xb0, 0x5f, 0x2d, 0xe4,
	0x9b, 0xe6, 0x71, 0x2a, 0x29, 0x94, 0xc0, 0x4f, 0xb8, 0x70, 0xb8, 0x4a, 0x73, 0x67, 0x91, 0x53,
	0x79, 0xed, 0x34, 0x80, 0xfd, 0x09, 0x46, 0x84, 0xa5, 0xbc, 0x62, 0x3e, 0xa3, 0x73, 0x26, 0xb1,
	0x0f, 0xc3, 0x54, 0xdc, 0x54, 0x92, 0xd5, 0x35, 0x17, 0xa5, 0x65, 0xbc, 0x30, 0x5e, 0x0f, 0xc7,
	0x2f, 0x9d, 0x7f, 0x84, 0xce, 0x74, 0xc3, 0x8a, 0x98, 0x52, 0xbc, 0xcc, 0x6a, 0xb2, 0x2d, 0xb5,
	0x4f, 0x00, 0xa2, 0x65, 0x99, 0xde, 0xf9, 0x3e, 0x87, 0xc1, 0x35, 0x2f, 0x58, 0x50, 0xce, 0xd9,
	0x0f, 0x0b, 0x35, 0xae, 0x5d, 0xb2, 0x19, 0xd8, 0xbf, 0x0c, 0xe8, 0xb7, 0xe4, 0xb0, 0xc2, 0x63,
	0x30, 0xd5, 0xb2, 0x62, 0x3a, 0xf9, 0x60, 0x7c, 0xbc, 0x23, 0x79, 0x45, 0x74, 0xe2, 0x86, 0x45,
	0x34, 0xf7, 0x4f, 0xf3, 0xce, 0x5f, 0xe6, 0xf8, 0x18, 0xe0, 0xaa, 0x10, 0xe9, 0xd7, 0x15, 0xdc,
	0xd5, 0xf0, 0xd6, 0xa4, 0x55, 0xeb, 0x2e, 0xaa, 0x68, 0x69, 0x99, 0x2b, 0xf5, 0x7a, 0x80, 0x31,
	0x98, 0x73, 0xaa, 0xa8, 0xd5, 0x6b, 0x80, 0x11, 0xd1, 0xb5, 0xed, 0x82, 0xd9, 0xa6, 0xe3, 0x01,
	0xf4, 0x26, 0x67, 0xe1, 0xf4, 0x3d, 0x7a, 0x84, 0x0f, 0x61, 0xa8, 0xcb, 0x84, 0xb8, 0xe7, 0xa7,
	0x1e, 0x32, 0xf0, 0x1e, 0x98, 0x33, 0x37, 0x76, 0x51, 0x07, 0x1f, 0xc1, 0x81, 0xef, 0x5d, 0x26,
	0x97, 0xe1, 0x45, 0x32, 0x0b, 0x66, 0x49, 0x10, 0xa3, 0x9f, 0xc8, 0xfe, 0x02, 0x87, 0x11, 0xcf,
	0x4a, 0xaa, 0x6e, 0xe5, 0xc3, 0x9f, 0xfe, 0x14, 0x06, 0x93, 0x76, 0x01, 0x9f, 0xd6, 0x39, 0x7e,
	0x06, 0x7b, 0x0b, 0x46, 0x75, 0xad, 0x3d, 0xf7, 0xc9, 0xba, 0x6f, 0x4f, 0x53, 0x2b, 0x29, 0xca,
	0x4c, 0xa3, 0x1d, 0xbd, 0xe2, 0xd6, 0xc4, 0xfe, 0x0e, 0x47, 0x3b, 0xc2, 0xb0, 0x07, 0x03, 0x5a,
	0x64, 0x42, 0x72, 0x95, 0xdf, 0xdc, 0x7d, 0xa8, 0x57, 0xff, 0x7f, 0x4f, 0xf7, 0x9e, 0x4e, 0x36,
	0x4a, 0x6c, 0xc1, 0xe3, 0x6f, 0xb7, 0xb4, 0xe0, 0x6a, 0xa9, 0xa3, 0x7b, 0xe4, 0xbe, 0x3d, 0x79,
	0x07, 0x4f, 0x77, 0x89, 0x31, 0x82, 0xd1, 0xc5, 0xf9, 0x34, 0xfc, 0xf0, 0x91, 0x78, 0x51, 0xe4,
	0xcd, 0x9a, 0xbb, 0x03, 0xf4, 0x27, 0x24, 0x8c, 0xcf, 0x02, 0x64, 0x4c, 0x7a, 0x9f, 0xbb, 0x4d,
	0xec, 0x55, 0x5f, 0xff, 0xec, 0x6f, 0x7f, 0x07, 0x00, 0x00, 0xff, 0xff, 0xf2, 0x20, 0xcc, 0xff,
	0xfd, 0x02, 0x00, 0x00,
}
