// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v3.21.12
// source: mission_service.proto

package missionservice

import (
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

type TriggerMissionEvent int32

const (
	TriggerMissionEvent_MISSION_EVENT_UNKNOWN           TriggerMissionEvent = 0
	TriggerMissionEvent_MISSION_EVENT_CREATE_ASSIGNMENT TriggerMissionEvent = 1
	TriggerMissionEvent_MISSION_EVENT_DONE_ASSIGNMENT   TriggerMissionEvent = 2
	TriggerMissionEvent_MISSION_EVENT_UNDONE_ASSIGNMENT TriggerMissionEvent = 3
	TriggerMissionEvent_MISSION_EVENT_DELETE_ASSIGNMENT TriggerMissionEvent = 4
)

// Enum value maps for TriggerMissionEvent.
var (
	TriggerMissionEvent_name = map[int32]string{
		0: "MISSION_EVENT_UNKNOWN",
		1: "MISSION_EVENT_CREATE_ASSIGNMENT",
		2: "MISSION_EVENT_DONE_ASSIGNMENT",
		3: "MISSION_EVENT_UNDONE_ASSIGNMENT",
		4: "MISSION_EVENT_DELETE_ASSIGNMENT",
	}
	TriggerMissionEvent_value = map[string]int32{
		"MISSION_EVENT_UNKNOWN":           0,
		"MISSION_EVENT_CREATE_ASSIGNMENT": 1,
		"MISSION_EVENT_DONE_ASSIGNMENT":   2,
		"MISSION_EVENT_UNDONE_ASSIGNMENT": 3,
		"MISSION_EVENT_DELETE_ASSIGNMENT": 4,
	}
)

func (x TriggerMissionEvent) Enum() *TriggerMissionEvent {
	p := new(TriggerMissionEvent)
	*p = x
	return p
}

func (x TriggerMissionEvent) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TriggerMissionEvent) Descriptor() protoreflect.EnumDescriptor {
	return file_mission_service_proto_enumTypes[0].Descriptor()
}

func (TriggerMissionEvent) Type() protoreflect.EnumType {
	return &file_mission_service_proto_enumTypes[0]
}

func (x TriggerMissionEvent) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TriggerMissionEvent.Descriptor instead.
func (TriggerMissionEvent) EnumDescriptor() ([]byte, []int) {
	return file_mission_service_proto_rawDescGZIP(), []int{0}
}

type TriggerMissionEventRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId string              `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Event  TriggerMissionEvent `protobuf:"varint,2,opt,name=event,proto3,enum=TriggerMissionEvent" json:"event,omitempty"`
}

func (x *TriggerMissionEventRequest) Reset() {
	*x = TriggerMissionEventRequest{}
	mi := &file_mission_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TriggerMissionEventRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TriggerMissionEventRequest) ProtoMessage() {}

func (x *TriggerMissionEventRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mission_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TriggerMissionEventRequest.ProtoReflect.Descriptor instead.
func (*TriggerMissionEventRequest) Descriptor() ([]byte, []int) {
	return file_mission_service_proto_rawDescGZIP(), []int{0}
}

func (x *TriggerMissionEventRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *TriggerMissionEventRequest) GetEvent() TriggerMissionEvent {
	if x != nil {
		return x.Event
	}
	return TriggerMissionEvent_MISSION_EVENT_UNKNOWN
}

type TriggerMissionEventResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TriggerMissionEventResponse) Reset() {
	*x = TriggerMissionEventResponse{}
	mi := &file_mission_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TriggerMissionEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TriggerMissionEventResponse) ProtoMessage() {}

func (x *TriggerMissionEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mission_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TriggerMissionEventResponse.ProtoReflect.Descriptor instead.
func (*TriggerMissionEventResponse) Descriptor() ([]byte, []int) {
	return file_mission_service_proto_rawDescGZIP(), []int{1}
}

var File_mission_service_proto protoreflect.FileDescriptor

var file_mission_service_proto_rawDesc = []byte{
	0x0a, 0x15, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x61, 0x0a, 0x1a, 0x54, 0x72, 0x69, 0x67, 0x67,
	0x65, 0x72, 0x4d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x2a,
	0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e,
	0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x4d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x1d, 0x0a, 0x1b, 0x54, 0x72,
	0x69, 0x67, 0x67, 0x65, 0x72, 0x4d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2a, 0xc2, 0x01, 0x0a, 0x13, 0x54, 0x72,
	0x69, 0x67, 0x67, 0x65, 0x72, 0x4d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x12, 0x19, 0x0a, 0x15, 0x4d, 0x49, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x45, 0x56, 0x45,
	0x4e, 0x54, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x00, 0x12, 0x23, 0x0a, 0x1f,
	0x4d, 0x49, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x43, 0x52,
	0x45, 0x41, 0x54, 0x45, 0x5f, 0x41, 0x53, 0x53, 0x49, 0x47, 0x4e, 0x4d, 0x45, 0x4e, 0x54, 0x10,
	0x01, 0x12, 0x21, 0x0a, 0x1d, 0x4d, 0x49, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x45, 0x56, 0x45,
	0x4e, 0x54, 0x5f, 0x44, 0x4f, 0x4e, 0x45, 0x5f, 0x41, 0x53, 0x53, 0x49, 0x47, 0x4e, 0x4d, 0x45,
	0x4e, 0x54, 0x10, 0x02, 0x12, 0x23, 0x0a, 0x1f, 0x4d, 0x49, 0x53, 0x53, 0x49, 0x4f, 0x4e, 0x5f,
	0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x55, 0x4e, 0x44, 0x4f, 0x4e, 0x45, 0x5f, 0x41, 0x53, 0x53,
	0x49, 0x47, 0x4e, 0x4d, 0x45, 0x4e, 0x54, 0x10, 0x03, 0x12, 0x23, 0x0a, 0x1f, 0x4d, 0x49, 0x53,
	0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x44, 0x45, 0x4c, 0x45, 0x54,
	0x45, 0x5f, 0x41, 0x53, 0x53, 0x49, 0x47, 0x4e, 0x4d, 0x45, 0x4e, 0x54, 0x10, 0x04, 0x32, 0x62,
	0x0a, 0x0e, 0x4d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x50, 0x0a, 0x13, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x4d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x1b, 0x2e, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65,
	0x72, 0x4d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x54, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x4d, 0x69,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x4d, 0x61, 0x73, 0x4c, 0x61, 0x7a, 0x75, 0x2f, 0x64, 0x65, 0x76, 0x2d, 0x6f, 0x70, 0x73,
	0x2d, 0x70, 0x6f, 0x72, 0x74, 0x6f, 0x2f, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mission_service_proto_rawDescOnce sync.Once
	file_mission_service_proto_rawDescData = file_mission_service_proto_rawDesc
)

func file_mission_service_proto_rawDescGZIP() []byte {
	file_mission_service_proto_rawDescOnce.Do(func() {
		file_mission_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_mission_service_proto_rawDescData)
	})
	return file_mission_service_proto_rawDescData
}

var file_mission_service_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_mission_service_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_mission_service_proto_goTypes = []any{
	(TriggerMissionEvent)(0),            // 0: TriggerMissionEvent
	(*TriggerMissionEventRequest)(nil),  // 1: TriggerMissionEventRequest
	(*TriggerMissionEventResponse)(nil), // 2: TriggerMissionEventResponse
}
var file_mission_service_proto_depIdxs = []int32{
	0, // 0: TriggerMissionEventRequest.event:type_name -> TriggerMissionEvent
	1, // 1: MissionService.TriggerMissionEvent:input_type -> TriggerMissionEventRequest
	2, // 2: MissionService.TriggerMissionEvent:output_type -> TriggerMissionEventResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_mission_service_proto_init() }
func file_mission_service_proto_init() {
	if File_mission_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mission_service_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_mission_service_proto_goTypes,
		DependencyIndexes: file_mission_service_proto_depIdxs,
		EnumInfos:         file_mission_service_proto_enumTypes,
		MessageInfos:      file_mission_service_proto_msgTypes,
	}.Build()
	File_mission_service_proto = out.File
	file_mission_service_proto_rawDesc = nil
	file_mission_service_proto_goTypes = nil
	file_mission_service_proto_depIdxs = nil
}
