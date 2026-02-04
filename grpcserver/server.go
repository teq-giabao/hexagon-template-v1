package grpcserver

import (
	"context"
	"io"
	"net"

	"hexagon/grpcserver/hexagonpb"

	"google.golang.org/grpc"
)

type Server struct {
	Addr       string
	grpcServer *grpc.Server
	listener   net.Listener
}

func New(addr string) *Server {
	return &Server{Addr: addr}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	s.listener = lis
	s.grpcServer = grpc.NewServer()
	hexagonpb.RegisterDemoServiceServer(s.grpcServer, &demoService{})
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

type demoService struct {
	hexagonpb.UnimplementedDemoServiceServer
}

func (demoService) Unary(_ context.Context, req *hexagonpb.EchoRequest) (*hexagonpb.EchoResponse, error) {
	return &hexagonpb.EchoResponse{
		Message: req.GetMessage(),
		Count:   1,
	}, nil
}

func (demoService) ServerStream(req *hexagonpb.StreamRequest, stream hexagonpb.DemoService_ServerStreamServer) error {
	count := req.GetCount()
	if count <= 0 {
		count = 1
	}

	for i := int32(1); i <= count; i++ {
		err := stream.Send(&hexagonpb.StreamResponse{
			Message: req.GetMessage(),
			Index:   i,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (demoService) ClientStream(stream hexagonpb.DemoService_ClientStreamServer) error {
	var (
		count   int32
		lastMsg string
	)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&hexagonpb.EchoResponse{
				Message: lastMsg,
				Count:   count,
			})
		}
		if err != nil {
			return err
		}

		count++
		lastMsg = req.GetMessage()
	}
}

func (demoService) BidiStream(stream hexagonpb.DemoService_BidiStreamServer) error {
	var idx int32
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		idx++
		err = stream.Send(&hexagonpb.StreamResponse{
			Message: req.GetMessage(),
			Index:   idx,
		})
		if err != nil {
			return err
		}
	}
}
