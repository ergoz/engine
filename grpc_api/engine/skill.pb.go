// Code generated by protoc-gen-go. DO NOT EDIT.
// source: skill.proto

package engine

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Skill struct {
	Id          int64  `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	DomainId    int64  `protobuf:"varint,2,opt,name=domain_id,json=domainId" json:"domain_id,omitempty"`
	Name        string `protobuf:"bytes,3,opt,name=name" json:"name,omitempty"`
	Description string `protobuf:"bytes,4,opt,name=description" json:"description,omitempty"`
}

func (m *Skill) Reset()                    { *m = Skill{} }
func (m *Skill) String() string            { return proto.CompactTextString(m) }
func (*Skill) ProtoMessage()               {}
func (*Skill) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func (m *Skill) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Skill) GetDomainId() int64 {
	if m != nil {
		return m.DomainId
	}
	return 0
}

func (m *Skill) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Skill) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

type ListSkill struct {
	Items []*Skill `protobuf:"bytes,1,rep,name=items" json:"items,omitempty"`
}

func (m *ListSkill) Reset()                    { *m = ListSkill{} }
func (m *ListSkill) String() string            { return proto.CompactTextString(m) }
func (*ListSkill) ProtoMessage()               {}
func (*ListSkill) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{1} }

func (m *ListSkill) GetItems() []*Skill {
	if m != nil {
		return m.Items
	}
	return nil
}

func init() {
	proto.RegisterType((*Skill)(nil), "engine.Skill")
	proto.RegisterType((*ListSkill)(nil), "engine.ListSkill")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for SkillApi service

type SkillApiClient interface {
	Create(ctx context.Context, in *Skill, opts ...grpc.CallOption) (*Skill, error)
	List(ctx context.Context, in *ListReqeust, opts ...grpc.CallOption) (*ListSkill, error)
	Get(ctx context.Context, in *ItemRequest, opts ...grpc.CallOption) (*Skill, error)
	Update(ctx context.Context, in *Skill, opts ...grpc.CallOption) (*Skill, error)
	Remove(ctx context.Context, in *ItemRequest, opts ...grpc.CallOption) (*Skill, error)
}

type skillApiClient struct {
	cc *grpc.ClientConn
}

func NewSkillApiClient(cc *grpc.ClientConn) SkillApiClient {
	return &skillApiClient{cc}
}

func (c *skillApiClient) Create(ctx context.Context, in *Skill, opts ...grpc.CallOption) (*Skill, error) {
	out := new(Skill)
	err := grpc.Invoke(ctx, "/engine.SkillApi/Create", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *skillApiClient) List(ctx context.Context, in *ListReqeust, opts ...grpc.CallOption) (*ListSkill, error) {
	out := new(ListSkill)
	err := grpc.Invoke(ctx, "/engine.SkillApi/List", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *skillApiClient) Get(ctx context.Context, in *ItemRequest, opts ...grpc.CallOption) (*Skill, error) {
	out := new(Skill)
	err := grpc.Invoke(ctx, "/engine.SkillApi/Get", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *skillApiClient) Update(ctx context.Context, in *Skill, opts ...grpc.CallOption) (*Skill, error) {
	out := new(Skill)
	err := grpc.Invoke(ctx, "/engine.SkillApi/Update", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *skillApiClient) Remove(ctx context.Context, in *ItemRequest, opts ...grpc.CallOption) (*Skill, error) {
	out := new(Skill)
	err := grpc.Invoke(ctx, "/engine.SkillApi/Remove", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for SkillApi service

type SkillApiServer interface {
	Create(context.Context, *Skill) (*Skill, error)
	List(context.Context, *ListReqeust) (*ListSkill, error)
	Get(context.Context, *ItemRequest) (*Skill, error)
	Update(context.Context, *Skill) (*Skill, error)
	Remove(context.Context, *ItemRequest) (*Skill, error)
}

func RegisterSkillApiServer(s *grpc.Server, srv SkillApiServer) {
	s.RegisterService(&_SkillApi_serviceDesc, srv)
}

func _SkillApi_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Skill)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SkillApiServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine.SkillApi/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SkillApiServer).Create(ctx, req.(*Skill))
	}
	return interceptor(ctx, in, info, handler)
}

func _SkillApi_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListReqeust)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SkillApiServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine.SkillApi/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SkillApiServer).List(ctx, req.(*ListReqeust))
	}
	return interceptor(ctx, in, info, handler)
}

func _SkillApi_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SkillApiServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine.SkillApi/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SkillApiServer).Get(ctx, req.(*ItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SkillApi_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Skill)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SkillApiServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine.SkillApi/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SkillApiServer).Update(ctx, req.(*Skill))
	}
	return interceptor(ctx, in, info, handler)
}

func _SkillApi_Remove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SkillApiServer).Remove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/engine.SkillApi/Remove",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SkillApiServer).Remove(ctx, req.(*ItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SkillApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "engine.SkillApi",
	HandlerType: (*SkillApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _SkillApi_Create_Handler,
		},
		{
			MethodName: "List",
			Handler:    _SkillApi_List_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _SkillApi_Get_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _SkillApi_Update_Handler,
		},
		{
			MethodName: "Remove",
			Handler:    _SkillApi_Remove_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "skill.proto",
}

func init() { proto.RegisterFile("skill.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 257 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x91, 0xcd, 0x4a, 0xc3, 0x40,
	0x10, 0xc7, 0xcd, 0x47, 0x43, 0x33, 0x41, 0xc1, 0xf1, 0x12, 0xe2, 0x25, 0xc4, 0x4b, 0x40, 0x08,
	0xa5, 0x3e, 0x81, 0x78, 0x90, 0x82, 0xa7, 0x15, 0xcf, 0x12, 0xbb, 0x83, 0x8c, 0x76, 0x77, 0xd3,
	0xec, 0xd6, 0xb7, 0xf6, 0x1d, 0x24, 0xbb, 0x14, 0x52, 0xbc, 0xf4, 0xb6, 0xf3, 0xfb, 0xcd, 0xc7,
	0x1f, 0x16, 0x0a, 0xfb, 0xcd, 0xbb, 0x5d, 0x37, 0x8c, 0xc6, 0x19, 0xcc, 0x48, 0x7f, 0xb2, 0xa6,
	0xaa, 0xd8, 0x1a, 0x6d, 0x5d, 0x80, 0xcd, 0x17, 0x2c, 0x5e, 0xa7, 0x1e, 0xbc, 0x82, 0x98, 0x65,
	0x19, 0xd5, 0x51, 0x9b, 0x88, 0x98, 0x25, 0xde, 0x42, 0x2e, 0x8d, 0xea, 0x59, 0xbf, 0xb3, 0x2c,
	0x63, 0x8f, 0x97, 0x01, 0x6c, 0x24, 0x22, 0xa4, 0xba, 0x57, 0x54, 0x26, 0x75, 0xd4, 0xe6, 0xc2,
	0xbf, 0xb1, 0x86, 0x42, 0x92, 0xdd, 0x8e, 0x3c, 0x38, 0x36, 0xba, 0x4c, 0xbd, 0x9a, 0xa3, 0x66,
	0x05, 0xf9, 0x0b, 0x5b, 0x17, 0xee, 0xdd, 0xc1, 0x82, 0x1d, 0x29, 0x5b, 0x46, 0x75, 0xd2, 0x16,
	0xeb, 0xcb, 0x2e, 0xa4, 0xeb, 0xbc, 0x15, 0xc1, 0xad, 0x7f, 0x23, 0x58, 0x7a, 0xf0, 0x38, 0x30,
	0xb6, 0x90, 0x3d, 0x8d, 0xd4, 0x3b, 0xc2, 0xd3, 0xe6, 0xea, 0xb4, 0x6c, 0x2e, 0x70, 0x05, 0xe9,
	0x74, 0x08, 0x6f, 0x8e, 0x62, 0xaa, 0x04, 0xed, 0xe9, 0x60, 0x5d, 0x75, 0x3d, 0x87, 0xc7, 0x89,
	0x7b, 0x48, 0x9e, 0x69, 0x36, 0xb0, 0x71, 0xa4, 0x04, 0xed, 0x0f, 0x64, 0xdd, 0xff, 0xf5, 0x2d,
	0x64, 0x6f, 0x83, 0x3c, 0x27, 0x48, 0x07, 0x99, 0x20, 0x65, 0x7e, 0xe8, 0xbc, 0xcd, 0x1f, 0x99,
	0xff, 0x94, 0x87, 0xbf, 0x00, 0x00, 0x00, 0xff, 0xff, 0xbb, 0x9a, 0x1d, 0x7c, 0xb8, 0x01, 0x00,
	0x00,
}
