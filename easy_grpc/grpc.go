package easy_grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcHandler struct {
	Conn   *grpc.ClientConn
	Ctx    context.Context
	Cancel context.CancelFunc
	Log    *zap.Logger
}

func NewGRPCHandler(address string) *grpcHandler {
	log, _ := zap.NewDevelopment()
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("cannot established connetion to grpc server.", zap.String("address", address), zap.String("error", err.Error()))
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	handler := &grpcHandler{
		Conn:   conn,
		Ctx:    ctx,
		Cancel: cancel,
		Log:    log,
	}
	return handler
}

func (h *grpcHandler) Close() {
	defer h.Cancel()
	defer h.Conn.Close()
}
