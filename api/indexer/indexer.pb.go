// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.21.12
// source: indexer/indexer.proto

package indexer

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SubscribeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 从哪个区块开始订阅
	StartBlock uint64 `protobuf:"varint,1,opt,name=start_block,json=startBlock,proto3" json:"start_block,omitempty"`
}

func (x *SubscribeRequest) Reset() {
	*x = SubscribeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeRequest) ProtoMessage() {}

func (x *SubscribeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeRequest.ProtoReflect.Descriptor instead.
func (*SubscribeRequest) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{0}
}

func (x *SubscribeRequest) GetStartBlock() uint64 {
	if x != nil {
		return x.StartBlock
	}
	return 0
}

type SubscribeReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 区块号
	BlockNumber uint64 `protobuf:"varint,1,opt,name=block_number,json=blockNumber,proto3" json:"block_number,omitempty"`
	// 上一个有事件发生的区块. 客户端用于检验是否有区块缺失的情况
	PrevBlockNumber uint64 `protobuf:"varint,2,opt,name=prev_block_number,json=prevBlockNumber,proto3" json:"prev_block_number,omitempty"`
	// 这个区块上发生的事件
	Events []*Event `protobuf:"bytes,3,rep,name=events,proto3" json:"events,omitempty"`
}

func (x *SubscribeReply) Reset() {
	*x = SubscribeReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeReply) ProtoMessage() {}

func (x *SubscribeReply) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeReply.ProtoReflect.Descriptor instead.
func (*SubscribeReply) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{1}
}

func (x *SubscribeReply) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

func (x *SubscribeReply) GetPrevBlockNumber() uint64 {
	if x != nil {
		return x.PrevBlockNumber
	}
	return 0
}

func (x *SubscribeReply) GetEvents() []*Event {
	if x != nil {
		return x.Events
	}
	return nil
}

type SubscribeSystemStatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SubscribeSystemStatusRequest) Reset() {
	*x = SubscribeSystemStatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeSystemStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeSystemStatusRequest) ProtoMessage() {}

func (x *SubscribeSystemStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeSystemStatusRequest.ProtoReflect.Descriptor instead.
func (*SubscribeSystemStatusRequest) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{2}
}

type SubscribeSystemStatusReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 区块链最新高度
	LatestBlock uint64 `protobuf:"varint,1,opt,name=latest_block,json=latestBlock,proto3" json:"latest_block,omitempty"`
	// 当前系统索引到的高度
	IndexedBlock uint64 `protobuf:"varint,2,opt,name=indexed_block,json=indexedBlock,proto3" json:"indexed_block,omitempty"`
	// 当前系统同步高度
	SyncBlock uint64 `protobuf:"varint,3,opt,name=sync_block,json=syncBlock,proto3" json:"sync_block,omitempty"`
}

func (x *SubscribeSystemStatusReply) Reset() {
	*x = SubscribeSystemStatusReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeSystemStatusReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeSystemStatusReply) ProtoMessage() {}

func (x *SubscribeSystemStatusReply) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeSystemStatusReply.ProtoReflect.Descriptor instead.
func (*SubscribeSystemStatusReply) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{3}
}

func (x *SubscribeSystemStatusReply) GetLatestBlock() uint64 {
	if x != nil {
		return x.LatestBlock
	}
	return 0
}

func (x *SubscribeSystemStatusReply) GetIndexedBlock() uint64 {
	if x != nil {
		return x.IndexedBlock
	}
	return 0
}

func (x *SubscribeSystemStatusReply) GetSyncBlock() uint64 {
	if x != nil {
		return x.SyncBlock
	}
	return 0
}

type QueryEventsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 从哪个区块开始订阅
	StartBlock uint64 `protobuf:"varint,1,opt,name=start_block,json=startBlock,proto3" json:"start_block,omitempty"`
	Size       int64  `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
}

func (x *QueryEventsRequest) Reset() {
	*x = QueryEventsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryEventsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryEventsRequest) ProtoMessage() {}

func (x *QueryEventsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryEventsRequest.ProtoReflect.Descriptor instead.
func (*QueryEventsRequest) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{4}
}

func (x *QueryEventsRequest) GetStartBlock() uint64 {
	if x != nil {
		return x.StartBlock
	}
	return 0
}

func (x *QueryEventsRequest) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

type QueryEventsReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventByBlocks []*QueryEventsReply_EventsByBlock `protobuf:"bytes,1,rep,name=event_by_blocks,json=eventByBlocks,proto3" json:"event_by_blocks,omitempty"`
}

func (x *QueryEventsReply) Reset() {
	*x = QueryEventsReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryEventsReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryEventsReply) ProtoMessage() {}

func (x *QueryEventsReply) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryEventsReply.ProtoReflect.Descriptor instead.
func (*QueryEventsReply) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{5}
}

func (x *QueryEventsReply) GetEventByBlocks() []*QueryEventsReply_EventsByBlock {
	if x != nil {
		return x.EventByBlocks
	}
	return nil
}

type QuerySystemStatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *QuerySystemStatusRequest) Reset() {
	*x = QuerySystemStatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QuerySystemStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QuerySystemStatusRequest) ProtoMessage() {}

func (x *QuerySystemStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QuerySystemStatusRequest.ProtoReflect.Descriptor instead.
func (*QuerySystemStatusRequest) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{6}
}

type QuerySystemStatusReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 当前系统同步高度
	SyncBlock uint64 `protobuf:"varint,1,opt,name=sync_block,json=syncBlock,proto3" json:"sync_block,omitempty"`
}

func (x *QuerySystemStatusReply) Reset() {
	*x = QuerySystemStatusReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QuerySystemStatusReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QuerySystemStatusReply) ProtoMessage() {}

func (x *QuerySystemStatusReply) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QuerySystemStatusReply.ProtoReflect.Descriptor instead.
func (*QuerySystemStatusReply) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{7}
}

func (x *QuerySystemStatusReply) GetSyncBlock() uint64 {
	if x != nil {
		return x.SyncBlock
	}
	return 0
}

type CheckTransferRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hash          string `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	PositionIndex int64  `protobuf:"varint,2,opt,name=position_index,json=positionIndex,proto3" json:"position_index,omitempty"`
}

func (x *CheckTransferRequest) Reset() {
	*x = CheckTransferRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckTransferRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckTransferRequest) ProtoMessage() {}

func (x *CheckTransferRequest) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckTransferRequest.ProtoReflect.Descriptor instead.
func (*CheckTransferRequest) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{8}
}

func (x *CheckTransferRequest) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

func (x *CheckTransferRequest) GetPositionIndex() int64 {
	if x != nil {
		return x.PositionIndex
	}
	return 0
}

type CheckTransferReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data *CheckTransferReply_TransferRecord `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *CheckTransferReply) Reset() {
	*x = CheckTransferReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckTransferReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckTransferReply) ProtoMessage() {}

func (x *CheckTransferReply) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckTransferReply.ProtoReflect.Descriptor instead.
func (*CheckTransferReply) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{9}
}

func (x *CheckTransferReply) GetData() *CheckTransferReply_TransferRecord {
	if x != nil {
		return x.Data
	}
	return nil
}

type QueryEventsReply_EventsByBlock struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 区块号
	BlockNumber uint64 `protobuf:"varint,1,opt,name=block_number,json=blockNumber,proto3" json:"block_number,omitempty"`
	// 上一个有事件发生的区块. 客户端用于检验是否有区块缺失的情况
	PrevBlockNumber uint64 `protobuf:"varint,2,opt,name=prev_block_number,json=prevBlockNumber,proto3" json:"prev_block_number,omitempty"`
	// 这个区块上发生的事件
	Events []*Event `protobuf:"bytes,3,rep,name=events,proto3" json:"events,omitempty"`
}

func (x *QueryEventsReply_EventsByBlock) Reset() {
	*x = QueryEventsReply_EventsByBlock{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryEventsReply_EventsByBlock) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryEventsReply_EventsByBlock) ProtoMessage() {}

func (x *QueryEventsReply_EventsByBlock) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryEventsReply_EventsByBlock.ProtoReflect.Descriptor instead.
func (*QueryEventsReply_EventsByBlock) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{5, 0}
}

func (x *QueryEventsReply_EventsByBlock) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

func (x *QueryEventsReply_EventsByBlock) GetPrevBlockNumber() uint64 {
	if x != nil {
		return x.PrevBlockNumber
	}
	return 0
}

func (x *QueryEventsReply_EventsByBlock) GetEvents() []*Event {
	if x != nil {
		return x.Events
	}
	return nil
}

type CheckTransferReply_TransferRecord struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sender   string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Receiver string `protobuf:"bytes,2,opt,name=receiver,proto3" json:"receiver,omitempty"`
	Tick     string `protobuf:"bytes,3,opt,name=tick,proto3" json:"tick,omitempty"`
	Amount   string `protobuf:"bytes,4,opt,name=amount,proto3" json:"amount,omitempty"`
	Status   bool   `protobuf:"varint,5,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *CheckTransferReply_TransferRecord) Reset() {
	*x = CheckTransferReply_TransferRecord{}
	if protoimpl.UnsafeEnabled {
		mi := &file_indexer_indexer_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckTransferReply_TransferRecord) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckTransferReply_TransferRecord) ProtoMessage() {}

func (x *CheckTransferReply_TransferRecord) ProtoReflect() protoreflect.Message {
	mi := &file_indexer_indexer_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckTransferReply_TransferRecord.ProtoReflect.Descriptor instead.
func (*CheckTransferReply_TransferRecord) Descriptor() ([]byte, []int) {
	return file_indexer_indexer_proto_rawDescGZIP(), []int{9, 0}
}

func (x *CheckTransferReply_TransferRecord) GetSender() string {
	if x != nil {
		return x.Sender
	}
	return ""
}

func (x *CheckTransferReply_TransferRecord) GetReceiver() string {
	if x != nil {
		return x.Receiver
	}
	return ""
}

func (x *CheckTransferReply_TransferRecord) GetTick() string {
	if x != nil {
		return x.Tick
	}
	return ""
}

func (x *CheckTransferReply_TransferRecord) GetAmount() string {
	if x != nil {
		return x.Amount
	}
	return ""
}

func (x *CheckTransferReply_TransferRecord) GetStatus() bool {
	if x != nil {
		return x.Status
	}
	return false
}

var File_indexer_indexer_proto protoreflect.FileDescriptor

var file_indexer_indexer_proto_rawDesc = []byte{
	0x0a, 0x15, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x64,
	0x65, 0x78, 0x65, 0x72, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x13, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2f, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x33, 0x0a, 0x10, 0x53, 0x75, 0x62, 0x73, 0x63,
	0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x8b, 0x01, 0x0a,
	0x0e, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12,
	0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62,
	0x65, 0x72, 0x12, 0x2a, 0x0a, 0x11, 0x70, 0x72, 0x65, 0x76, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0f, 0x70,
	0x72, 0x65, 0x76, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x2a,
	0x0a, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2e, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x1e, 0x0a, 0x1c, 0x53, 0x75,
	0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x83, 0x01, 0x0a, 0x1a, 0x53,
	0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x21, 0x0a, 0x0c, 0x6c, 0x61, 0x74,
	0x65, 0x73, 0x74, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x0b, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x23, 0x0a, 0x0d,
	0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x64, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x0c, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x64, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x73, 0x79, 0x6e, 0x63, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x22, 0x49, 0x0a, 0x12, 0x51, 0x75, 0x65, 0x72, 0x79, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x22, 0xf4, 0x01, 0x0a, 0x10,
	0x51, 0x75, 0x65, 0x72, 0x79, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x79,
	0x12, 0x53, 0x0a, 0x0f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x62, 0x79, 0x5f, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x42,
	0x79, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x0d, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x42, 0x79, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x1a, 0x8a, 0x01, 0x0a, 0x0d, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73,
	0x42, 0x79, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x21, 0x0a, 0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x2a, 0x0a, 0x11, 0x70, 0x72,
	0x65, 0x76, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0f, 0x70, 0x72, 0x65, 0x76, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x2a, 0x0a, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x64,
	0x65, 0x78, 0x65, 0x72, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x22, 0x1a, 0x0a, 0x18, 0x51, 0x75, 0x65, 0x72, 0x79, 0x53, 0x79, 0x73, 0x74, 0x65,
	0x6d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x37,
	0x0a, 0x16, 0x51, 0x75, 0x65, 0x72, 0x79, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x79, 0x6e, 0x63,
	0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x73, 0x79,
	0x6e, 0x63, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x51, 0x0a, 0x14, 0x43, 0x68, 0x65, 0x63, 0x6b,
	0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68,
	0x61, 0x73, 0x68, 0x12, 0x25, 0x0a, 0x0e, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x70, 0x6f, 0x73,
	0x69, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x22, 0xe3, 0x01, 0x0a, 0x12, 0x43,
	0x68, 0x65, 0x63, 0x6b, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x12, 0x42, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x2e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2e, 0x43, 0x68,
	0x65, 0x63, 0x6b, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x52, 0x65, 0x70, 0x6c, 0x79,
	0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x88, 0x01, 0x0a, 0x0e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66,
	0x65, 0x72, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e, 0x64,
	0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72,
	0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04,
	0x74, 0x69, 0x63, 0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x69, 0x63, 0x6b,
	0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x32, 0xaf, 0x04, 0x0a, 0x07, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x12, 0x4e, 0x0a, 0x0e,
	0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x1d,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2e, 0x53, 0x75, 0x62,
	0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2e, 0x53, 0x75, 0x62, 0x73,
	0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x30, 0x01, 0x12, 0x6d, 0x0a, 0x15,
	0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x29, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x64, 0x65,
	0x78, 0x65, 0x72, 0x2e, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x53, 0x79, 0x73,
	0x74, 0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x27, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2e, 0x53,
	0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x30, 0x01, 0x12, 0x6b, 0x0a, 0x0b, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x1f, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x1c, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x16, 0x12, 0x14, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x32, 0x2f, 0x69, 0x6e, 0x64, 0x65,
	0x78, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x7d, 0x0a, 0x11, 0x51, 0x75, 0x65, 0x72,
	0x79, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x25, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2e, 0x51, 0x75, 0x65, 0x72,
	0x79, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78,
	0x65, 0x72, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x1c, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x16, 0x12, 0x14, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x32, 0x2f, 0x69, 0x6e, 0x64, 0x65, 0x78,
	0x2f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x79, 0x0a, 0x0d, 0x43, 0x68, 0x65, 0x63, 0x6b,
	0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x12, 0x21, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69,
	0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x54, 0x72, 0x61, 0x6e,
	0x73, 0x66, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x54,
	0x72, 0x61, 0x6e, 0x73, 0x66, 0x65, 0x72, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x24, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x1e, 0x12, 0x1c, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x32, 0x2f, 0x69, 0x6e,
	0x64, 0x65, 0x78, 0x2f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x66,
	0x65, 0x72, 0x42, 0x44, 0x0a, 0x0b, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65,
	0x72, 0x50, 0x01, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x49, 0x45, 0x72, 0x63, 0x4f, 0x72, 0x67, 0x2f, 0x49, 0x45, 0x52, 0x43, 0x5f, 0x49, 0x6e, 0x64,
	0x65, 0x78, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72,
	0x3b, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_indexer_indexer_proto_rawDescOnce sync.Once
	file_indexer_indexer_proto_rawDescData = file_indexer_indexer_proto_rawDesc
)

func file_indexer_indexer_proto_rawDescGZIP() []byte {
	file_indexer_indexer_proto_rawDescOnce.Do(func() {
		file_indexer_indexer_proto_rawDescData = protoimpl.X.CompressGZIP(file_indexer_indexer_proto_rawDescData)
	})
	return file_indexer_indexer_proto_rawDescData
}

var file_indexer_indexer_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_indexer_indexer_proto_goTypes = []interface{}{
	(*SubscribeRequest)(nil),                  // 0: api.indexer.SubscribeRequest
	(*SubscribeReply)(nil),                    // 1: api.indexer.SubscribeReply
	(*SubscribeSystemStatusRequest)(nil),      // 2: api.indexer.SubscribeSystemStatusRequest
	(*SubscribeSystemStatusReply)(nil),        // 3: api.indexer.SubscribeSystemStatusReply
	(*QueryEventsRequest)(nil),                // 4: api.indexer.QueryEventsRequest
	(*QueryEventsReply)(nil),                  // 5: api.indexer.QueryEventsReply
	(*QuerySystemStatusRequest)(nil),          // 6: api.indexer.QuerySystemStatusRequest
	(*QuerySystemStatusReply)(nil),            // 7: api.indexer.QuerySystemStatusReply
	(*CheckTransferRequest)(nil),              // 8: api.indexer.CheckTransferRequest
	(*CheckTransferReply)(nil),                // 9: api.indexer.CheckTransferReply
	(*QueryEventsReply_EventsByBlock)(nil),    // 10: api.indexer.QueryEventsReply.EventsByBlock
	(*CheckTransferReply_TransferRecord)(nil), // 11: api.indexer.CheckTransferReply.TransferRecord
	(*Event)(nil),                             // 12: api.indexer.Event
}
var file_indexer_indexer_proto_depIdxs = []int32{
	12, // 0: api.indexer.SubscribeReply.events:type_name -> api.indexer.Event
	10, // 1: api.indexer.QueryEventsReply.event_by_blocks:type_name -> api.indexer.QueryEventsReply.EventsByBlock
	11, // 2: api.indexer.CheckTransferReply.data:type_name -> api.indexer.CheckTransferReply.TransferRecord
	12, // 3: api.indexer.QueryEventsReply.EventsByBlock.events:type_name -> api.indexer.Event
	0,  // 4: api.indexer.Indexer.SubscribeEvent:input_type -> api.indexer.SubscribeRequest
	2,  // 5: api.indexer.Indexer.SubscribeSystemStatus:input_type -> api.indexer.SubscribeSystemStatusRequest
	4,  // 6: api.indexer.Indexer.QueryEvents:input_type -> api.indexer.QueryEventsRequest
	6,  // 7: api.indexer.Indexer.QuerySystemStatus:input_type -> api.indexer.QuerySystemStatusRequest
	8,  // 8: api.indexer.Indexer.CheckTransfer:input_type -> api.indexer.CheckTransferRequest
	1,  // 9: api.indexer.Indexer.SubscribeEvent:output_type -> api.indexer.SubscribeReply
	3,  // 10: api.indexer.Indexer.SubscribeSystemStatus:output_type -> api.indexer.SubscribeSystemStatusReply
	5,  // 11: api.indexer.Indexer.QueryEvents:output_type -> api.indexer.QueryEventsReply
	7,  // 12: api.indexer.Indexer.QuerySystemStatus:output_type -> api.indexer.QuerySystemStatusReply
	9,  // 13: api.indexer.Indexer.CheckTransfer:output_type -> api.indexer.CheckTransferReply
	9,  // [9:14] is the sub-list for method output_type
	4,  // [4:9] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_indexer_indexer_proto_init() }
func file_indexer_indexer_proto_init() {
	if File_indexer_indexer_proto != nil {
		return
	}
	file_indexer_event_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_indexer_indexer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_indexer_indexer_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_indexer_indexer_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeSystemStatusRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_indexer_indexer_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeSystemStatusReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_indexer_indexer_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryEventsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_indexer_indexer_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryEventsReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_indexer_indexer_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QuerySystemStatusRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_indexer_indexer_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QuerySystemStatusReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_indexer_indexer_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckTransferRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_indexer_indexer_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckTransferReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_indexer_indexer_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryEventsReply_EventsByBlock); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_indexer_indexer_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckTransferReply_TransferRecord); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_indexer_indexer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_indexer_indexer_proto_goTypes,
		DependencyIndexes: file_indexer_indexer_proto_depIdxs,
		MessageInfos:      file_indexer_indexer_proto_msgTypes,
	}.Build()
	File_indexer_indexer_proto = out.File
	file_indexer_indexer_proto_rawDesc = nil
	file_indexer_indexer_proto_goTypes = nil
	file_indexer_indexer_proto_depIdxs = nil
}
