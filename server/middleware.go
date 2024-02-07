package server

import (
	"fmt"

	"google.golang.org/grpc"
)

func StreamInterceptor() func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		fmt.Printf("stream start: %s\n", info.FullMethod)
		defer fmt.Printf("stream end: %s\n", info.FullMethod)
		return handler(srv, &streamHandler{ss})
	}
}

type streamHandler struct {
	grpc.ServerStream
}

var _ grpc.ServerStream = (*streamHandler)(nil)

func (h *streamHandler) SendMsg(m any) error {
	fmt.Printf("before send: %v\n", m)
	defer fmt.Printf("after send: %v\n", m)
	return h.ServerStream.SendMsg(m)
}

func (h *streamHandler) RecvMsg(m any) error {
	fmt.Printf("before recv: %v\n", m)
	defer fmt.Printf("after recv: %v\n", m)
	return h.ServerStream.RecvMsg(m)
}
