// Code generated by protoc-gen-go. DO NOT EDIT.
// source: node.proto

package node

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

type ResultCode int32

const (
	ResultCode_SUCCESS            ResultCode = 0
	ResultCode_ERROR              ResultCode = 1
	ResultCode_SIGN_VALIDATE_FAIL ResultCode = 401
	ResultCode_SIGN_TIMEOUT       ResultCode = 402
	ResultCode_ENTRY_NOT_FOUND    ResultCode = 403
	ResultCode_ENTRY_INVALIDATE   ResultCode = 404
	ResultCode_CALL_YOUR_SELF     ResultCode = 405
	ResultCode_CALL_STACK_ERROR   ResultCode = 406
	ResultCode_NODE_NOT_EXIST     ResultCode = 407
	ResultCode_STACK_EMPTY        ResultCode = 408
)

var ResultCode_name = map[int32]string{
	0:   "SUCCESS",
	1:   "ERROR",
	401: "SIGN_VALIDATE_FAIL",
	402: "SIGN_TIMEOUT",
	403: "ENTRY_NOT_FOUND",
	404: "ENTRY_INVALIDATE",
	405: "CALL_YOUR_SELF",
	406: "CALL_STACK_ERROR",
	407: "NODE_NOT_EXIST",
	408: "STACK_EMPTY",
}

var ResultCode_value = map[string]int32{
	"SUCCESS":            0,
	"ERROR":              1,
	"SIGN_VALIDATE_FAIL": 401,
	"SIGN_TIMEOUT":       402,
	"ENTRY_NOT_FOUND":    403,
	"ENTRY_INVALIDATE":   404,
	"CALL_YOUR_SELF":     405,
	"CALL_STACK_ERROR":   406,
	"NODE_NOT_EXIST":     407,
	"STACK_EMPTY":        408,
}

func (x ResultCode) String() string {
	return proto.EnumName(ResultCode_name, int32(x))
}

func (ResultCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{0}
}

type EntryType int32

const (
	EntryType_INVALIDATE EntryType = 0
	EntryType_HTTP       EntryType = 1
	EntryType_RPC        EntryType = 2
)

var EntryType_name = map[int32]string{
	0: "INVALIDATE",
	1: "HTTP",
	2: "RPC",
}

var EntryType_value = map[string]int32{
	"INVALIDATE": 0,
	"HTTP":       1,
	"RPC":        2,
}

func (x EntryType) String() string {
	return proto.EnumName(EntryType_name, int32(x))
}

func (EntryType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{1}
}

type RequestMode int32

const (
	RequestMode_DEFAULT      RequestMode = 0
	RequestMode_LOCAL_FORCE  RequestMode = 2
	RequestMode_LOCAL_BETTER RequestMode = 3
)

var RequestMode_name = map[int32]string{
	0: "DEFAULT",
	2: "LOCAL_FORCE",
	3: "LOCAL_BETTER",
}

var RequestMode_value = map[string]int32{
	"DEFAULT":      0,
	"LOCAL_FORCE":  2,
	"LOCAL_BETTER": 3,
}

func (x RequestMode) String() string {
	return proto.EnumName(RequestMode_name, int32(x))
}

func (RequestMode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{2}
}

type SayHelloReq struct {
	NodeInfo             *NodeInfo `protobuf:"bytes,1,opt,name=NodeInfo,proto3" json:"NodeInfo,omitempty"`
	Sign                 string    `protobuf:"bytes,2,opt,name=Sign,proto3" json:"Sign,omitempty"`
	TimeStamp            int64     `protobuf:"varint,3,opt,name=TimeStamp,proto3" json:"TimeStamp,omitempty"`
	DataVersion          uint64    `protobuf:"varint,4,opt,name=DataVersion,proto3" json:"DataVersion,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *SayHelloReq) Reset()         { *m = SayHelloReq{} }
func (m *SayHelloReq) String() string { return proto.CompactTextString(m) }
func (*SayHelloReq) ProtoMessage()    {}
func (*SayHelloReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{0}
}

func (m *SayHelloReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SayHelloReq.Unmarshal(m, b)
}
func (m *SayHelloReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SayHelloReq.Marshal(b, m, deterministic)
}
func (m *SayHelloReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SayHelloReq.Merge(m, src)
}
func (m *SayHelloReq) XXX_Size() int {
	return xxx_messageInfo_SayHelloReq.Size(m)
}
func (m *SayHelloReq) XXX_DiscardUnknown() {
	xxx_messageInfo_SayHelloReq.DiscardUnknown(m)
}

var xxx_messageInfo_SayHelloReq proto.InternalMessageInfo

func (m *SayHelloReq) GetNodeInfo() *NodeInfo {
	if m != nil {
		return m.NodeInfo
	}
	return nil
}

func (m *SayHelloReq) GetSign() string {
	if m != nil {
		return m.Sign
	}
	return ""
}

func (m *SayHelloReq) GetTimeStamp() int64 {
	if m != nil {
		return m.TimeStamp
	}
	return 0
}

func (m *SayHelloReq) GetDataVersion() uint64 {
	if m != nil {
		return m.DataVersion
	}
	return 0
}

type SayHelloRsp struct {
	Result               ResultCode `protobuf:"varint,1,opt,name=Result,proto3,enum=node.ResultCode" json:"Result,omitempty"`
	Sign                 string     `protobuf:"bytes,2,opt,name=Sign,proto3" json:"Sign,omitempty"`
	TimeStamp            int64      `protobuf:"varint,3,opt,name=TimeStamp,proto3" json:"TimeStamp,omitempty"`
	NodeInfo             *NodeInfo  `protobuf:"bytes,4,opt,name=NodeInfo,proto3" json:"NodeInfo,omitempty"`
	DataVersion          uint64     `protobuf:"varint,5,opt,name=DataVersion,proto3" json:"DataVersion,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *SayHelloRsp) Reset()         { *m = SayHelloRsp{} }
func (m *SayHelloRsp) String() string { return proto.CompactTextString(m) }
func (*SayHelloRsp) ProtoMessage()    {}
func (*SayHelloRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{1}
}

func (m *SayHelloRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SayHelloRsp.Unmarshal(m, b)
}
func (m *SayHelloRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SayHelloRsp.Marshal(b, m, deterministic)
}
func (m *SayHelloRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SayHelloRsp.Merge(m, src)
}
func (m *SayHelloRsp) XXX_Size() int {
	return xxx_messageInfo_SayHelloRsp.Size(m)
}
func (m *SayHelloRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_SayHelloRsp.DiscardUnknown(m)
}

var xxx_messageInfo_SayHelloRsp proto.InternalMessageInfo

func (m *SayHelloRsp) GetResult() ResultCode {
	if m != nil {
		return m.Result
	}
	return ResultCode_SUCCESS
}

func (m *SayHelloRsp) GetSign() string {
	if m != nil {
		return m.Sign
	}
	return ""
}

func (m *SayHelloRsp) GetTimeStamp() int64 {
	if m != nil {
		return m.TimeStamp
	}
	return 0
}

func (m *SayHelloRsp) GetNodeInfo() *NodeInfo {
	if m != nil {
		return m.NodeInfo
	}
	return nil
}

func (m *SayHelloRsp) GetDataVersion() uint64 {
	if m != nil {
		return m.DataVersion
	}
	return 0
}

type NodeInfo struct {
	Name                 string   `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	IP                   string   `protobuf:"bytes,2,opt,name=IP,proto3" json:"IP,omitempty"`
	Port                 uint32   `protobuf:"varint,3,opt,name=Port,proto3" json:"Port,omitempty"`
	InstanceID           uint64   `protobuf:"varint,4,opt,name=InstanceID,proto3" json:"InstanceID,omitempty"`
	AgentPort            uint32   `protobuf:"varint,5,opt,name=AgentPort,proto3" json:"AgentPort,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NodeInfo) Reset()         { *m = NodeInfo{} }
func (m *NodeInfo) String() string { return proto.CompactTextString(m) }
func (*NodeInfo) ProtoMessage()    {}
func (*NodeInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{2}
}

func (m *NodeInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NodeInfo.Unmarshal(m, b)
}
func (m *NodeInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NodeInfo.Marshal(b, m, deterministic)
}
func (m *NodeInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NodeInfo.Merge(m, src)
}
func (m *NodeInfo) XXX_Size() int {
	return xxx_messageInfo_NodeInfo.Size(m)
}
func (m *NodeInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_NodeInfo.DiscardUnknown(m)
}

var xxx_messageInfo_NodeInfo proto.InternalMessageInfo

func (m *NodeInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *NodeInfo) GetIP() string {
	if m != nil {
		return m.IP
	}
	return ""
}

func (m *NodeInfo) GetPort() uint32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *NodeInfo) GetInstanceID() uint64 {
	if m != nil {
		return m.InstanceID
	}
	return 0
}

func (m *NodeInfo) GetAgentPort() uint32 {
	if m != nil {
		return m.AgentPort
	}
	return 0
}

type EntryInfo struct {
	URI                  string    `protobuf:"bytes,1,opt,name=URI,proto3" json:"URI,omitempty"`
	EntryType            EntryType `protobuf:"varint,2,opt,name=EntryType,proto3,enum=node.EntryType" json:"EntryType,omitempty"`
	IsNew                bool      `protobuf:"varint,3,opt,name=IsNew,proto3" json:"IsNew,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *EntryInfo) Reset()         { *m = EntryInfo{} }
func (m *EntryInfo) String() string { return proto.CompactTextString(m) }
func (*EntryInfo) ProtoMessage()    {}
func (*EntryInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{3}
}

func (m *EntryInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EntryInfo.Unmarshal(m, b)
}
func (m *EntryInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EntryInfo.Marshal(b, m, deterministic)
}
func (m *EntryInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EntryInfo.Merge(m, src)
}
func (m *EntryInfo) XXX_Size() int {
	return xxx_messageInfo_EntryInfo.Size(m)
}
func (m *EntryInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_EntryInfo.DiscardUnknown(m)
}

var xxx_messageInfo_EntryInfo proto.InternalMessageInfo

func (m *EntryInfo) GetURI() string {
	if m != nil {
		return m.URI
	}
	return ""
}

func (m *EntryInfo) GetEntryType() EntryType {
	if m != nil {
		return m.EntryType
	}
	return EntryType_INVALIDATE
}

func (m *EntryInfo) GetIsNew() bool {
	if m != nil {
		return m.IsNew
	}
	return false
}

type NewEntryNotify struct {
	EntryInfo            []*EntryInfo `protobuf:"bytes,1,rep,name=EntryInfo,proto3" json:"EntryInfo,omitempty"`
	DataVersion          uint64       `protobuf:"varint,2,opt,name=DataVersion,proto3" json:"DataVersion,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *NewEntryNotify) Reset()         { *m = NewEntryNotify{} }
func (m *NewEntryNotify) String() string { return proto.CompactTextString(m) }
func (*NewEntryNotify) ProtoMessage()    {}
func (*NewEntryNotify) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{4}
}

func (m *NewEntryNotify) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NewEntryNotify.Unmarshal(m, b)
}
func (m *NewEntryNotify) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NewEntryNotify.Marshal(b, m, deterministic)
}
func (m *NewEntryNotify) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NewEntryNotify.Merge(m, src)
}
func (m *NewEntryNotify) XXX_Size() int {
	return xxx_messageInfo_NewEntryNotify.Size(m)
}
func (m *NewEntryNotify) XXX_DiscardUnknown() {
	xxx_messageInfo_NewEntryNotify.DiscardUnknown(m)
}

var xxx_messageInfo_NewEntryNotify proto.InternalMessageInfo

func (m *NewEntryNotify) GetEntryInfo() []*EntryInfo {
	if m != nil {
		return m.EntryInfo
	}
	return nil
}

func (m *NewEntryNotify) GetDataVersion() uint64 {
	if m != nil {
		return m.DataVersion
	}
	return 0
}

type GetEntryReq struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetEntryReq) Reset()         { *m = GetEntryReq{} }
func (m *GetEntryReq) String() string { return proto.CompactTextString(m) }
func (*GetEntryReq) ProtoMessage()    {}
func (*GetEntryReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{5}
}

func (m *GetEntryReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetEntryReq.Unmarshal(m, b)
}
func (m *GetEntryReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetEntryReq.Marshal(b, m, deterministic)
}
func (m *GetEntryReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetEntryReq.Merge(m, src)
}
func (m *GetEntryReq) XXX_Size() int {
	return xxx_messageInfo_GetEntryReq.Size(m)
}
func (m *GetEntryReq) XXX_DiscardUnknown() {
	xxx_messageInfo_GetEntryReq.DiscardUnknown(m)
}

var xxx_messageInfo_GetEntryReq proto.InternalMessageInfo

type GetEntryResp struct {
	EntryInfo            []*EntryInfo `protobuf:"bytes,1,rep,name=EntryInfo,proto3" json:"EntryInfo,omitempty"`
	DataVersion          uint64       `protobuf:"varint,2,opt,name=DataVersion,proto3" json:"DataVersion,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *GetEntryResp) Reset()         { *m = GetEntryResp{} }
func (m *GetEntryResp) String() string { return proto.CompactTextString(m) }
func (*GetEntryResp) ProtoMessage()    {}
func (*GetEntryResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{6}
}

func (m *GetEntryResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetEntryResp.Unmarshal(m, b)
}
func (m *GetEntryResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetEntryResp.Marshal(b, m, deterministic)
}
func (m *GetEntryResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetEntryResp.Merge(m, src)
}
func (m *GetEntryResp) XXX_Size() int {
	return xxx_messageInfo_GetEntryResp.Size(m)
}
func (m *GetEntryResp) XXX_DiscardUnknown() {
	xxx_messageInfo_GetEntryResp.DiscardUnknown(m)
}

var xxx_messageInfo_GetEntryResp proto.InternalMessageInfo

func (m *GetEntryResp) GetEntryInfo() []*EntryInfo {
	if m != nil {
		return m.EntryInfo
	}
	return nil
}

func (m *GetEntryResp) GetDataVersion() uint64 {
	if m != nil {
		return m.DataVersion
	}
	return 0
}

type SayGoodbyeNotify struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SayGoodbyeNotify) Reset()         { *m = SayGoodbyeNotify{} }
func (m *SayGoodbyeNotify) String() string { return proto.CompactTextString(m) }
func (*SayGoodbyeNotify) ProtoMessage()    {}
func (*SayGoodbyeNotify) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{7}
}

func (m *SayGoodbyeNotify) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SayGoodbyeNotify.Unmarshal(m, b)
}
func (m *SayGoodbyeNotify) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SayGoodbyeNotify.Marshal(b, m, deterministic)
}
func (m *SayGoodbyeNotify) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SayGoodbyeNotify.Merge(m, src)
}
func (m *SayGoodbyeNotify) XXX_Size() int {
	return xxx_messageInfo_SayGoodbyeNotify.Size(m)
}
func (m *SayGoodbyeNotify) XXX_DiscardUnknown() {
	xxx_messageInfo_SayGoodbyeNotify.DiscardUnknown(m)
}

var xxx_messageInfo_SayGoodbyeNotify proto.InternalMessageInfo

type KeepAliveNotify struct {
	DataVersion          uint64      `protobuf:"varint,1,opt,name=DataVersion,proto3" json:"DataVersion,omitempty"`
	Neighbours           []*NodeInfo `protobuf:"bytes,2,rep,name=Neighbours,proto3" json:"Neighbours,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *KeepAliveNotify) Reset()         { *m = KeepAliveNotify{} }
func (m *KeepAliveNotify) String() string { return proto.CompactTextString(m) }
func (*KeepAliveNotify) ProtoMessage()    {}
func (*KeepAliveNotify) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{8}
}

func (m *KeepAliveNotify) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KeepAliveNotify.Unmarshal(m, b)
}
func (m *KeepAliveNotify) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KeepAliveNotify.Marshal(b, m, deterministic)
}
func (m *KeepAliveNotify) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KeepAliveNotify.Merge(m, src)
}
func (m *KeepAliveNotify) XXX_Size() int {
	return xxx_messageInfo_KeepAliveNotify.Size(m)
}
func (m *KeepAliveNotify) XXX_DiscardUnknown() {
	xxx_messageInfo_KeepAliveNotify.DiscardUnknown(m)
}

var xxx_messageInfo_KeepAliveNotify proto.InternalMessageInfo

func (m *KeepAliveNotify) GetDataVersion() uint64 {
	if m != nil {
		return m.DataVersion
	}
	return 0
}

func (m *KeepAliveNotify) GetNeighbours() []*NodeInfo {
	if m != nil {
		return m.Neighbours
	}
	return nil
}

type RpcCallReq struct {
	Data                 []byte      `protobuf:"bytes,1,opt,name=Data,proto3" json:"Data,omitempty"`
	URI                  string      `protobuf:"bytes,2,opt,name=URI,proto3" json:"URI,omitempty"`
	Caller               []string    `protobuf:"bytes,3,rep,name=Caller,proto3" json:"Caller,omitempty"`
	RequestMode          RequestMode `protobuf:"varint,4,opt,name=RequestMode,proto3,enum=node.RequestMode" json:"RequestMode,omitempty"`
	Key                  string      `protobuf:"bytes,5,opt,name=Key,proto3" json:"Key,omitempty"`
	EntryType            EntryType   `protobuf:"varint,6,opt,name=EntryType,proto3,enum=node.EntryType" json:"EntryType,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *RpcCallReq) Reset()         { *m = RpcCallReq{} }
func (m *RpcCallReq) String() string { return proto.CompactTextString(m) }
func (*RpcCallReq) ProtoMessage()    {}
func (*RpcCallReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{9}
}

func (m *RpcCallReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RpcCallReq.Unmarshal(m, b)
}
func (m *RpcCallReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RpcCallReq.Marshal(b, m, deterministic)
}
func (m *RpcCallReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RpcCallReq.Merge(m, src)
}
func (m *RpcCallReq) XXX_Size() int {
	return xxx_messageInfo_RpcCallReq.Size(m)
}
func (m *RpcCallReq) XXX_DiscardUnknown() {
	xxx_messageInfo_RpcCallReq.DiscardUnknown(m)
}

var xxx_messageInfo_RpcCallReq proto.InternalMessageInfo

func (m *RpcCallReq) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *RpcCallReq) GetURI() string {
	if m != nil {
		return m.URI
	}
	return ""
}

func (m *RpcCallReq) GetCaller() []string {
	if m != nil {
		return m.Caller
	}
	return nil
}

func (m *RpcCallReq) GetRequestMode() RequestMode {
	if m != nil {
		return m.RequestMode
	}
	return RequestMode_DEFAULT
}

func (m *RpcCallReq) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *RpcCallReq) GetEntryType() EntryType {
	if m != nil {
		return m.EntryType
	}
	return EntryType_INVALIDATE
}

type RpcCallResp struct {
	Result               ResultCode `protobuf:"varint,1,opt,name=Result,proto3,enum=node.ResultCode" json:"Result,omitempty"`
	Data                 []byte     `protobuf:"bytes,2,opt,name=Data,proto3" json:"Data,omitempty"`
	URI                  string     `protobuf:"bytes,3,opt,name=URI,proto3" json:"URI,omitempty"`
	Caller               []string   `protobuf:"bytes,4,rep,name=Caller,proto3" json:"Caller,omitempty"`
	EntryType            EntryType  `protobuf:"varint,5,opt,name=EntryType,proto3,enum=node.EntryType" json:"EntryType,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *RpcCallResp) Reset()         { *m = RpcCallResp{} }
func (m *RpcCallResp) String() string { return proto.CompactTextString(m) }
func (*RpcCallResp) ProtoMessage()    {}
func (*RpcCallResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_0c843d59d2d938e7, []int{10}
}

func (m *RpcCallResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RpcCallResp.Unmarshal(m, b)
}
func (m *RpcCallResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RpcCallResp.Marshal(b, m, deterministic)
}
func (m *RpcCallResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RpcCallResp.Merge(m, src)
}
func (m *RpcCallResp) XXX_Size() int {
	return xxx_messageInfo_RpcCallResp.Size(m)
}
func (m *RpcCallResp) XXX_DiscardUnknown() {
	xxx_messageInfo_RpcCallResp.DiscardUnknown(m)
}

var xxx_messageInfo_RpcCallResp proto.InternalMessageInfo

func (m *RpcCallResp) GetResult() ResultCode {
	if m != nil {
		return m.Result
	}
	return ResultCode_SUCCESS
}

func (m *RpcCallResp) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *RpcCallResp) GetURI() string {
	if m != nil {
		return m.URI
	}
	return ""
}

func (m *RpcCallResp) GetCaller() []string {
	if m != nil {
		return m.Caller
	}
	return nil
}

func (m *RpcCallResp) GetEntryType() EntryType {
	if m != nil {
		return m.EntryType
	}
	return EntryType_INVALIDATE
}

func init() {
	proto.RegisterEnum("node.ResultCode", ResultCode_name, ResultCode_value)
	proto.RegisterEnum("node.EntryType", EntryType_name, EntryType_value)
	proto.RegisterEnum("node.RequestMode", RequestMode_name, RequestMode_value)
	proto.RegisterType((*SayHelloReq)(nil), "node.SayHelloReq")
	proto.RegisterType((*SayHelloRsp)(nil), "node.SayHelloRsp")
	proto.RegisterType((*NodeInfo)(nil), "node.NodeInfo")
	proto.RegisterType((*EntryInfo)(nil), "node.EntryInfo")
	proto.RegisterType((*NewEntryNotify)(nil), "node.NewEntryNotify")
	proto.RegisterType((*GetEntryReq)(nil), "node.GetEntryReq")
	proto.RegisterType((*GetEntryResp)(nil), "node.GetEntryResp")
	proto.RegisterType((*SayGoodbyeNotify)(nil), "node.SayGoodbyeNotify")
	proto.RegisterType((*KeepAliveNotify)(nil), "node.KeepAliveNotify")
	proto.RegisterType((*RpcCallReq)(nil), "node.RpcCallReq")
	proto.RegisterType((*RpcCallResp)(nil), "node.RpcCallResp")
}

func init() { proto.RegisterFile("node.proto", fileDescriptor_0c843d59d2d938e7) }

var fileDescriptor_0c843d59d2d938e7 = []byte{
	// 732 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x95, 0xcf, 0x6e, 0xda, 0x4e,
	0x10, 0xc7, 0xe3, 0x3f, 0x90, 0x30, 0x26, 0x66, 0xb3, 0xbf, 0xfc, 0x5a, 0x0e, 0x55, 0x85, 0x7c,
	0x42, 0x48, 0xe5, 0x40, 0xce, 0x3d, 0xb8, 0xc6, 0x24, 0x56, 0x1c, 0x1b, 0xad, 0x97, 0xa8, 0x39,
	0x59, 0x26, 0x6c, 0x52, 0x24, 0x62, 0x13, 0xec, 0x34, 0xe2, 0xd6, 0x17, 0xe8, 0xa1, 0xff, 0xfb,
	0x00, 0x7d, 0x87, 0x5e, 0xfb, 0x08, 0x7d, 0xa4, 0xca, 0x6b, 0x03, 0x06, 0x2a, 0x45, 0xad, 0x7a,
	0x9b, 0xfd, 0xee, 0xcc, 0xf0, 0xf9, 0x8e, 0x67, 0x05, 0x40, 0x18, 0x8d, 0x58, 0x7b, 0x3a, 0x8b,
	0x92, 0x08, 0xcb, 0x69, 0xac, 0xbd, 0x15, 0x40, 0xf1, 0x82, 0xf9, 0x09, 0x9b, 0x4c, 0x22, 0xc2,
	0x6e, 0x71, 0x0b, 0xf6, 0x9c, 0x68, 0xc4, 0xac, 0xf0, 0x2a, 0xaa, 0x0b, 0x0d, 0xa1, 0xa9, 0x74,
	0xd4, 0x36, 0x2f, 0x5a, 0xa8, 0x64, 0x79, 0x8f, 0x31, 0xc8, 0xde, 0xf8, 0x3a, 0xac, 0x8b, 0x0d,
	0xa1, 0x59, 0x21, 0x3c, 0xc6, 0x4f, 0xa0, 0x42, 0xc7, 0x37, 0xcc, 0x4b, 0x82, 0x9b, 0x69, 0x5d,
	0x6a, 0x08, 0x4d, 0x89, 0xac, 0x04, 0xdc, 0x00, 0xa5, 0x1b, 0x24, 0xc1, 0x39, 0x9b, 0xc5, 0xe3,
	0x28, 0xac, 0xcb, 0x0d, 0xa1, 0x29, 0x93, 0xa2, 0xa4, 0x7d, 0x2f, 0xf2, 0xc4, 0x53, 0xdc, 0x84,
	0x32, 0x61, 0xf1, 0xdd, 0x24, 0xe1, 0x34, 0x6a, 0x07, 0x65, 0x34, 0x99, 0x66, 0x44, 0x23, 0x46,
	0xf2, 0xfb, 0xbf, 0xa0, 0x29, 0x7a, 0x95, 0x1f, 0xf0, 0xba, 0x41, 0x5e, 0xda, 0x26, 0x7f, 0x23,
	0xc0, 0xda, 0x68, 0x9c, 0xe0, 0x86, 0x71, 0xe8, 0x0a, 0xe1, 0x31, 0x56, 0x41, 0xb4, 0xfa, 0x39,
	0x9e, 0x68, 0xf5, 0xd3, 0x9c, 0x7e, 0x34, 0x4b, 0x38, 0xd7, 0x3e, 0xe1, 0x31, 0x7e, 0x0a, 0x60,
	0x85, 0x71, 0x12, 0x84, 0x97, 0xcc, 0xea, 0xe6, 0xf3, 0x29, 0x28, 0xa9, 0x21, 0xfd, 0x9a, 0x85,
	0x09, 0x2f, 0x2c, 0xf1, 0xc2, 0x95, 0xa0, 0x0d, 0xa1, 0x62, 0x86, 0xc9, 0x6c, 0xce, 0x11, 0x10,
	0x48, 0x03, 0x62, 0xe5, 0x04, 0x69, 0x88, 0x9f, 0xe5, 0xd7, 0x74, 0x3e, 0x65, 0x9c, 0x43, 0xed,
	0xd4, 0x32, 0xc3, 0x4b, 0x99, 0xac, 0x32, 0xf0, 0x21, 0x94, 0xac, 0xd8, 0x61, 0xf7, 0x1c, 0x70,
	0x8f, 0x64, 0x07, 0x2d, 0x00, 0xd5, 0x61, 0xf7, 0x3c, 0xcb, 0x89, 0x92, 0xf1, 0xd5, 0x7c, 0xd9,
	0x36, 0xdf, 0x19, 0xa9, 0xa9, 0xac, 0xb5, 0xe5, 0x83, 0x2c, 0x70, 0x6d, 0x4c, 0x52, 0xdc, 0x9e,
	0xe4, 0x3e, 0x28, 0xc7, 0x2c, 0xe1, 0x15, 0x84, 0xdd, 0x6a, 0x3e, 0x54, 0x57, 0xc7, 0x78, 0xfa,
	0xef, 0x7f, 0x0f, 0x03, 0xf2, 0x82, 0xf9, 0x71, 0x14, 0x8d, 0x86, 0x73, 0x96, 0x99, 0xd2, 0x2e,
	0xa1, 0x76, 0xca, 0xd8, 0x54, 0x9f, 0x8c, 0x5f, 0xe7, 0xd2, 0x66, 0x23, 0x61, 0xab, 0x11, 0x6e,
	0x03, 0x38, 0x6c, 0x7c, 0xfd, 0x6a, 0x18, 0xdd, 0xcd, 0xe2, 0xba, 0xc8, 0xd1, 0x36, 0x57, 0xaa,
	0x90, 0xa1, 0xfd, 0x10, 0x00, 0xc8, 0xf4, 0xd2, 0x08, 0x26, 0x93, 0xf4, 0xed, 0x61, 0x90, 0xd3,
	0x6e, 0xbc, 0x73, 0x95, 0xf0, 0x78, 0xf1, 0x15, 0xc5, 0xd5, 0x57, 0x7c, 0x04, 0xe5, 0xb4, 0x80,
	0xcd, 0xea, 0x52, 0x43, 0x6a, 0x56, 0x48, 0x7e, 0xc2, 0x47, 0xa0, 0x10, 0x76, 0x7b, 0xc7, 0xe2,
	0xe4, 0x2c, 0x1a, 0x31, 0xbe, 0x3b, 0x6a, 0xe7, 0x60, 0xf1, 0x5c, 0x96, 0x17, 0xa4, 0x98, 0x95,
	0xb6, 0x3f, 0x65, 0x73, 0xbe, 0x49, 0x15, 0x92, 0x86, 0xeb, 0x4b, 0x52, 0x7e, 0x68, 0x49, 0xb4,
	0x6f, 0x02, 0x28, 0x4b, 0x0b, 0x7f, 0xfa, 0x5e, 0xb9, 0x5b, 0x71, 0xdb, 0xad, 0xf4, 0x3b, 0xb7,
	0xf2, 0x9a, 0xdb, 0x35, 0xcc, 0xd2, 0x43, 0x98, 0xad, 0x9f, 0xe9, 0xa4, 0x97, 0x0c, 0x58, 0x81,
	0x5d, 0x6f, 0x60, 0x18, 0xa6, 0xe7, 0xa1, 0x1d, 0x5c, 0x81, 0x92, 0x49, 0x88, 0x4b, 0x90, 0x80,
	0x1f, 0x03, 0xf6, 0xac, 0x63, 0xc7, 0x3f, 0xd7, 0x6d, 0xab, 0xab, 0x53, 0xd3, 0xef, 0xe9, 0x96,
	0x8d, 0xde, 0x49, 0xf8, 0x00, 0xaa, 0xfc, 0x82, 0x5a, 0x67, 0xa6, 0x3b, 0xa0, 0xe8, 0xbd, 0x84,
	0x0f, 0xa1, 0x66, 0x3a, 0x94, 0x5c, 0xf8, 0x8e, 0x4b, 0xfd, 0x9e, 0x3b, 0x70, 0xba, 0xe8, 0x83,
	0x84, 0xff, 0x07, 0x94, 0xa9, 0x96, 0xb3, 0x68, 0x82, 0x3e, 0x4a, 0xf8, 0x3f, 0x50, 0x0d, 0xdd,
	0xb6, 0xfd, 0x0b, 0x77, 0x40, 0x7c, 0xcf, 0xb4, 0x7b, 0xe8, 0x13, 0xcf, 0xe5, 0xa2, 0x47, 0x75,
	0xe3, 0xd4, 0xcf, 0x18, 0x3e, 0xf3, 0x5c, 0xc7, 0xed, 0x9a, 0xbc, 0xaf, 0xf9, 0xd2, 0xf2, 0x28,
	0xfa, 0x22, 0x61, 0x04, 0x4a, 0x9e, 0x76, 0xd6, 0xa7, 0x17, 0xe8, 0xab, 0xd4, 0x6a, 0x17, 0x26,
	0x80, 0x55, 0x80, 0xc2, 0x0f, 0xee, 0xe0, 0x3d, 0x90, 0x4f, 0x28, 0xed, 0x23, 0x01, 0xef, 0x82,
	0x44, 0xfa, 0x06, 0x12, 0x5b, 0xcf, 0xd7, 0xf6, 0x23, 0x1d, 0x41, 0xd7, 0xec, 0xe9, 0x03, 0x9b,
	0xa2, 0x1d, 0x5c, 0x03, 0xc5, 0x76, 0x0d, 0xdd, 0xf6, 0x7b, 0x2e, 0x31, 0x4c, 0x24, 0x62, 0x04,
	0xd5, 0x4c, 0x78, 0x61, 0x52, 0x6a, 0x12, 0x24, 0x0d, 0xcb, 0xfc, 0x5f, 0xe3, 0xe8, 0x57, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x3e, 0xc3, 0x62, 0x41, 0x43, 0x06, 0x00, 0x00,
}
