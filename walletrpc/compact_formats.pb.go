// Code generated by protoc-gen-go. DO NOT EDIT.
// source: compact_formats.proto

package walletrpc

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// CompactBlock is a packaging of ONLY the data from a block that's needed to:
//   1. Detect a payment to your shielded Sapling address
//   2. Detect a spend of your shielded Sapling notes
//   3. Update your witnesses to generate new Sapling spend proofs.
type CompactBlock struct {
	ProtoVersion         uint32       `protobuf:"varint,1,opt,name=protoVersion,proto3" json:"protoVersion,omitempty"`
	Height               uint64       `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
	Hash                 []byte       `protobuf:"bytes,3,opt,name=hash,proto3" json:"hash,omitempty"`
	PrevHash             []byte       `protobuf:"bytes,4,opt,name=prevHash,proto3" json:"prevHash,omitempty"`
	Time                 uint32       `protobuf:"varint,5,opt,name=time,proto3" json:"time,omitempty"`
	Header               []byte       `protobuf:"bytes,6,opt,name=header,proto3" json:"header,omitempty"`
	Vtx                  []*CompactTx `protobuf:"bytes,7,rep,name=vtx,proto3" json:"vtx,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *CompactBlock) Reset()         { *m = CompactBlock{} }
func (m *CompactBlock) String() string { return proto.CompactTextString(m) }
func (*CompactBlock) ProtoMessage()    {}
func (*CompactBlock) Descriptor() ([]byte, []int) {
	return fileDescriptor_dce29fee3ee34899, []int{0}
}

func (m *CompactBlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CompactBlock.Unmarshal(m, b)
}
func (m *CompactBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CompactBlock.Marshal(b, m, deterministic)
}
func (m *CompactBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CompactBlock.Merge(m, src)
}
func (m *CompactBlock) XXX_Size() int {
	return xxx_messageInfo_CompactBlock.Size(m)
}
func (m *CompactBlock) XXX_DiscardUnknown() {
	xxx_messageInfo_CompactBlock.DiscardUnknown(m)
}

var xxx_messageInfo_CompactBlock proto.InternalMessageInfo

func (m *CompactBlock) GetProtoVersion() uint32 {
	if m != nil {
		return m.ProtoVersion
	}
	return 0
}

func (m *CompactBlock) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *CompactBlock) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *CompactBlock) GetPrevHash() []byte {
	if m != nil {
		return m.PrevHash
	}
	return nil
}

func (m *CompactBlock) GetTime() uint32 {
	if m != nil {
		return m.Time
	}
	return 0
}

func (m *CompactBlock) GetHeader() []byte {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *CompactBlock) GetVtx() []*CompactTx {
	if m != nil {
		return m.Vtx
	}
	return nil
}

// Index and hash will allow the receiver to call out to chain
// explorers or other data structures to retrieve more information
// about this transaction.
type CompactTx struct {
	Index uint64 `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	Hash  []byte `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	// The transaction fee: present if server can provide. In the case of a
	// stateless server and a transaction with transparent inputs, this will be
	// unset because the calculation requires reference to prior transactions.
	// in a pure-Sapling context, the fee will be calculable as:
	//    valueBalance + (sum(vPubNew) - sum(vPubOld) - sum(tOut))
	Fee                  uint32           `protobuf:"varint,3,opt,name=fee,proto3" json:"fee,omitempty"`
	Spends               []*CompactSpend  `protobuf:"bytes,4,rep,name=spends,proto3" json:"spends,omitempty"`
	Outputs              []*CompactOutput `protobuf:"bytes,5,rep,name=outputs,proto3" json:"outputs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *CompactTx) Reset()         { *m = CompactTx{} }
func (m *CompactTx) String() string { return proto.CompactTextString(m) }
func (*CompactTx) ProtoMessage()    {}
func (*CompactTx) Descriptor() ([]byte, []int) {
	return fileDescriptor_dce29fee3ee34899, []int{1}
}

func (m *CompactTx) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CompactTx.Unmarshal(m, b)
}
func (m *CompactTx) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CompactTx.Marshal(b, m, deterministic)
}
func (m *CompactTx) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CompactTx.Merge(m, src)
}
func (m *CompactTx) XXX_Size() int {
	return xxx_messageInfo_CompactTx.Size(m)
}
func (m *CompactTx) XXX_DiscardUnknown() {
	xxx_messageInfo_CompactTx.DiscardUnknown(m)
}

var xxx_messageInfo_CompactTx proto.InternalMessageInfo

func (m *CompactTx) GetIndex() uint64 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *CompactTx) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *CompactTx) GetFee() uint32 {
	if m != nil {
		return m.Fee
	}
	return 0
}

func (m *CompactTx) GetSpends() []*CompactSpend {
	if m != nil {
		return m.Spends
	}
	return nil
}

func (m *CompactTx) GetOutputs() []*CompactOutput {
	if m != nil {
		return m.Outputs
	}
	return nil
}

type CompactSpend struct {
	Nf                   []byte   `protobuf:"bytes,1,opt,name=nf,proto3" json:"nf,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CompactSpend) Reset()         { *m = CompactSpend{} }
func (m *CompactSpend) String() string { return proto.CompactTextString(m) }
func (*CompactSpend) ProtoMessage()    {}
func (*CompactSpend) Descriptor() ([]byte, []int) {
	return fileDescriptor_dce29fee3ee34899, []int{2}
}

func (m *CompactSpend) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CompactSpend.Unmarshal(m, b)
}
func (m *CompactSpend) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CompactSpend.Marshal(b, m, deterministic)
}
func (m *CompactSpend) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CompactSpend.Merge(m, src)
}
func (m *CompactSpend) XXX_Size() int {
	return xxx_messageInfo_CompactSpend.Size(m)
}
func (m *CompactSpend) XXX_DiscardUnknown() {
	xxx_messageInfo_CompactSpend.DiscardUnknown(m)
}

var xxx_messageInfo_CompactSpend proto.InternalMessageInfo

func (m *CompactSpend) GetNf() []byte {
	if m != nil {
		return m.Nf
	}
	return nil
}

type CompactOutput struct {
	Cmu                  []byte   `protobuf:"bytes,1,opt,name=cmu,proto3" json:"cmu,omitempty"`
	Epk                  []byte   `protobuf:"bytes,2,opt,name=epk,proto3" json:"epk,omitempty"`
	Ciphertext           []byte   `protobuf:"bytes,3,opt,name=ciphertext,proto3" json:"ciphertext,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CompactOutput) Reset()         { *m = CompactOutput{} }
func (m *CompactOutput) String() string { return proto.CompactTextString(m) }
func (*CompactOutput) ProtoMessage()    {}
func (*CompactOutput) Descriptor() ([]byte, []int) {
	return fileDescriptor_dce29fee3ee34899, []int{3}
}

func (m *CompactOutput) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CompactOutput.Unmarshal(m, b)
}
func (m *CompactOutput) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CompactOutput.Marshal(b, m, deterministic)
}
func (m *CompactOutput) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CompactOutput.Merge(m, src)
}
func (m *CompactOutput) XXX_Size() int {
	return xxx_messageInfo_CompactOutput.Size(m)
}
func (m *CompactOutput) XXX_DiscardUnknown() {
	xxx_messageInfo_CompactOutput.DiscardUnknown(m)
}

var xxx_messageInfo_CompactOutput proto.InternalMessageInfo

func (m *CompactOutput) GetCmu() []byte {
	if m != nil {
		return m.Cmu
	}
	return nil
}

func (m *CompactOutput) GetEpk() []byte {
	if m != nil {
		return m.Epk
	}
	return nil
}

func (m *CompactOutput) GetCiphertext() []byte {
	if m != nil {
		return m.Ciphertext
	}
	return nil
}

func init() {
	proto.RegisterType((*CompactBlock)(nil), "cash.z.wallet.sdk.rpc.CompactBlock")
	proto.RegisterType((*CompactTx)(nil), "cash.z.wallet.sdk.rpc.CompactTx")
	proto.RegisterType((*CompactSpend)(nil), "cash.z.wallet.sdk.rpc.CompactSpend")
	proto.RegisterType((*CompactOutput)(nil), "cash.z.wallet.sdk.rpc.CompactOutput")
}

func init() { proto.RegisterFile("compact_formats.proto", fileDescriptor_dce29fee3ee34899) }

var fileDescriptor_dce29fee3ee34899 = []byte{
	// 353 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x52, 0x3f, 0x4f, 0xeb, 0x30,
	0x1c, 0x7c, 0xf9, 0xd3, 0xf4, 0xf5, 0xf7, 0xd2, 0xea, 0xc9, 0x7a, 0x7d, 0xb2, 0x18, 0xaa, 0x28,
	0x30, 0x64, 0xca, 0x50, 0x46, 0x24, 0x86, 0xb2, 0xb0, 0x21, 0xb9, 0x88, 0x81, 0x05, 0x05, 0xc7,
	0x21, 0x51, 0x9b, 0xc4, 0xb2, 0xdd, 0x12, 0xf1, 0xd1, 0x98, 0xf9, 0x14, 0x7c, 0x1a, 0x64, 0xc7,
	0x44, 0xad, 0x84, 0xba, 0xdd, 0x9d, 0xee, 0x9c, 0x3b, 0xfd, 0x02, 0x73, 0xda, 0xd6, 0x3c, 0xa3,
	0xea, 0xa9, 0x68, 0x45, 0x9d, 0x29, 0x99, 0x72, 0xd1, 0xaa, 0x16, 0xcd, 0x69, 0x26, 0xcb, 0xf4,
	0x2d, 0x7d, 0xcd, 0xb6, 0x5b, 0xa6, 0x52, 0x99, 0x6f, 0x52, 0xc1, 0x69, 0xfc, 0xe9, 0x40, 0x78,
	0xd3, 0x07, 0x56, 0xdb, 0x96, 0x6e, 0x50, 0x0c, 0xa1, 0x09, 0x3c, 0x30, 0x21, 0xab, 0xb6, 0xc1,
	0x4e, 0xe4, 0x24, 0x53, 0x72, 0xa4, 0xa1, 0xff, 0x10, 0x94, 0xac, 0x7a, 0x29, 0x15, 0x76, 0x23,
	0x27, 0xf1, 0x89, 0x65, 0x08, 0x81, 0x5f, 0x66, 0xb2, 0xc4, 0x5e, 0xe4, 0x24, 0x21, 0x31, 0x18,
	0x9d, 0xc1, 0x6f, 0x2e, 0xd8, 0xfe, 0x56, 0xeb, 0xbe, 0xd1, 0x07, 0xae, 0xfd, 0xaa, 0xaa, 0x19,
	0x1e, 0x99, 0x6f, 0x18, 0xdc, 0xbf, 0x9d, 0xe5, 0x4c, 0xe0, 0xc0, 0xb8, 0x2d, 0x43, 0x4b, 0xf0,
	0xf6, 0xaa, 0xc3, 0xe3, 0xc8, 0x4b, 0xfe, 0x2c, 0xa3, 0xf4, 0xc7, 0x35, 0xa9, 0x5d, 0x72, 0xdf,
	0x11, 0x6d, 0x8e, 0x3f, 0x1c, 0x98, 0x0c, 0x12, 0xfa, 0x07, 0xa3, 0xaa, 0xc9, 0x59, 0x67, 0x26,
	0xf9, 0xa4, 0x27, 0x43, 0x67, 0xf7, 0xa0, 0xf3, 0x5f, 0xf0, 0x0a, 0xc6, 0xcc, 0x8c, 0x29, 0xd1,
	0x10, 0x5d, 0x41, 0x20, 0x39, 0x6b, 0x72, 0x89, 0x7d, 0x53, 0xe0, 0xfc, 0x74, 0x81, 0xb5, 0xf6,
	0x12, 0x1b, 0x41, 0xd7, 0x30, 0x6e, 0x77, 0x8a, 0xef, 0x94, 0xc4, 0x23, 0x93, 0xbe, 0x38, 0x9d,
	0xbe, 0x33, 0x66, 0xf2, 0x1d, 0x8a, 0x17, 0xc3, 0x89, 0xcc, 0xbb, 0x68, 0x06, 0x6e, 0x53, 0x98,
	0x15, 0x21, 0x71, 0x9b, 0x22, 0x5e, 0xc3, 0xf4, 0x28, 0xa9, 0xfb, 0xd3, 0x7a, 0x67, 0x1d, 0x1a,
	0x6a, 0x85, 0xf1, 0x8d, 0x1d, 0xa9, 0x21, 0x5a, 0x00, 0xd0, 0x8a, 0x97, 0x4c, 0x28, 0xd6, 0x29,
	0x7b, 0xb1, 0x03, 0x65, 0x35, 0x7b, 0x9c, 0xf4, 0xed, 0x04, 0xa7, 0xef, 0xee, 0xaf, 0xe7, 0xc0,
	0xfc, 0x01, 0x97, 0x5f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x3d, 0x39, 0x8c, 0xce, 0x5f, 0x02, 0x00,
	0x00,
}
