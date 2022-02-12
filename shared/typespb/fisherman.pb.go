// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.3
// source: fisherman.proto

package typespb

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

type Fisherman struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address         []byte   `protobuf:"bytes,1,opt,name=Address,proto3" json:"Address,omitempty"`
	PublicKey       []byte   `protobuf:"bytes,2,opt,name=PublicKey,proto3" json:"PublicKey,omitempty"`
	Paused          bool     `protobuf:"varint,3,opt,name=Paused,proto3" json:"Paused,omitempty"`
	Status          int32    `protobuf:"varint,4,opt,name=Status,proto3" json:"Status,omitempty"`
	Chains          []string `protobuf:"bytes,5,rep,name=Chains,proto3" json:"Chains,omitempty"`
	ServiceURL      string   `protobuf:"bytes,6,opt,name=ServiceURL,proto3" json:"ServiceURL,omitempty"`
	StakedTokens    string   `protobuf:"bytes,7,opt,name=StakedTokens,proto3" json:"StakedTokens,omitempty"`
	PausedHeight    uint64   `protobuf:"varint,8,opt,name=PausedHeight,proto3" json:"PausedHeight,omitempty"`
	UnstakingHeight int64    `protobuf:"varint,9,opt,name=UnstakingHeight,proto3" json:"UnstakingHeight,omitempty"`
	Output          []byte   `protobuf:"bytes,10,opt,name=Output,proto3" json:"Output,omitempty"`
}

func (x *Fisherman) Reset() {
	*x = Fisherman{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fisherman_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Fisherman) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Fisherman) ProtoMessage() {}

func (x *Fisherman) ProtoReflect() protoreflect.Message {
	mi := &file_fisherman_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Fisherman.ProtoReflect.Descriptor instead.
func (*Fisherman) Descriptor() ([]byte, []int) {
	return file_fisherman_proto_rawDescGZIP(), []int{0}
}

func (x *Fisherman) GetAddress() []byte {
	if x != nil {
		return x.Address
	}
	return nil
}

func (x *Fisherman) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *Fisherman) GetPaused() bool {
	if x != nil {
		return x.Paused
	}
	return false
}

func (x *Fisherman) GetStatus() int32 {
	if x != nil {
		return x.Status
	}
	return 0
}

func (x *Fisherman) GetChains() []string {
	if x != nil {
		return x.Chains
	}
	return nil
}

func (x *Fisherman) GetServiceURL() string {
	if x != nil {
		return x.ServiceURL
	}
	return ""
}

func (x *Fisherman) GetStakedTokens() string {
	if x != nil {
		return x.StakedTokens
	}
	return ""
}

func (x *Fisherman) GetPausedHeight() uint64 {
	if x != nil {
		return x.PausedHeight
	}
	return 0
}

func (x *Fisherman) GetUnstakingHeight() int64 {
	if x != nil {
		return x.UnstakingHeight
	}
	return 0
}

func (x *Fisherman) GetOutput() []byte {
	if x != nil {
		return x.Output
	}
	return nil
}

var File_fisherman_proto protoreflect.FileDescriptor

var file_fisherman_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x66, 0x69, 0x73, 0x68, 0x65, 0x72, 0x6d, 0x61, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x07, 0x74, 0x79, 0x70, 0x65, 0x73, 0x70, 0x62, 0x22, 0xb5, 0x02, 0x0a, 0x09, 0x46,
	0x69, 0x73, 0x68, 0x65, 0x72, 0x6d, 0x61, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79,
	0x12, 0x16, 0x0a, 0x06, 0x50, 0x61, 0x75, 0x73, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x06, 0x50, 0x61, 0x75, 0x73, 0x65, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x16, 0x0a, 0x06, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x06, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x55, 0x52, 0x4c, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x55, 0x52, 0x4c, 0x12, 0x22, 0x0a, 0x0c, 0x53, 0x74, 0x61, 0x6b,
	0x65, 0x64, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c,
	0x53, 0x74, 0x61, 0x6b, 0x65, 0x64, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x73, 0x12, 0x22, 0x0a, 0x0c,
	0x50, 0x61, 0x75, 0x73, 0x65, 0x64, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x0c, 0x50, 0x61, 0x75, 0x73, 0x65, 0x64, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74,
	0x12, 0x28, 0x0a, 0x0f, 0x55, 0x6e, 0x73, 0x74, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x48, 0x65, 0x69,
	0x67, 0x68, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x55, 0x6e, 0x73, 0x74, 0x61,
	0x6b, 0x69, 0x6e, 0x67, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x4f, 0x75,
	0x74, 0x70, 0x75, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x4f, 0x75, 0x74, 0x70,
	0x75, 0x74, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x70, 0x62, 0x2f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_fisherman_proto_rawDescOnce sync.Once
	file_fisherman_proto_rawDescData = file_fisherman_proto_rawDesc
)

func file_fisherman_proto_rawDescGZIP() []byte {
	file_fisherman_proto_rawDescOnce.Do(func() {
		file_fisherman_proto_rawDescData = protoimpl.X.CompressGZIP(file_fisherman_proto_rawDescData)
	})
	return file_fisherman_proto_rawDescData
}

var file_fisherman_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_fisherman_proto_goTypes = []interface{}{
	(*Fisherman)(nil), // 0: typespb.Fisherman
}
var file_fisherman_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_fisherman_proto_init() }
func file_fisherman_proto_init() {
	if File_fisherman_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_fisherman_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Fisherman); i {
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
			RawDescriptor: file_fisherman_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_fisherman_proto_goTypes,
		DependencyIndexes: file_fisherman_proto_depIdxs,
		MessageInfos:      file_fisherman_proto_msgTypes,
	}.Build()
	File_fisherman_proto = out.File
	file_fisherman_proto_rawDesc = nil
	file_fisherman_proto_goTypes = nil
	file_fisherman_proto_depIdxs = nil
}
