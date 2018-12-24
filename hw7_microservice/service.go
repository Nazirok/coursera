package main

import (
	"encoding/json"
	"fmt"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"

)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type (
	MicroService struct {
		acl   map[string][]string
		admin *AdminService
		biz   *BizService
	}

	AdminService struct {
	}

	BizService struct {
	}

	AClStore struct {
		service string
		method  string
	}
)

func NewMicroService(acl map[string][]string) *MicroService {
	return &MicroService{
		acl:   acl,
		admin: &AdminService{},
		biz:   &BizService{},
	}
}

func (m *MicroService) GetAdminService() *AdminService {
	return m.admin
}

func (m *MicroService) GetBizService() *BizService {
	return m.biz
}

func (m *MicroService) authInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	err := m.authorize(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (m *MicroService) streamAuthInterceptor (
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	err := m.authorize(ss.Context(), info.FullMethod)
	if err != nil {
		return err
	}
	return handler(srv, ss)
}

func (m *MicroService) authorize(ctx context.Context, method string) error {
	md, _ := metadata.FromIncomingContext(ctx)
	consumer := md.Get("consumer")
	if len(consumer) == 0 {
		return status.Error(codes.Unauthenticated, "consumer not found")
	}
	list, ok := m.acl[consumer[0]]
	if !ok {
		return status.Error(codes.Unauthenticated, "unknown consumer")
	}
	if !allowedMethod(method, list) {
		return status.Error(codes.Unauthenticated, "method not allowed")
	}
	return nil
}

func StartMyMicroservice(ctx context.Context, conn string, acl string) error {
	access, err := parseACL(acl)
	if err != nil {
		return err
	}
	micro := NewMicroService(access)
	lis, err := net.Listen("tcp", conn)
	if err != nil {
		return fmt.Errorf("can`t listen port %s", err)
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(micro.authInterceptor),
		grpc.StreamInterceptor(micro.streamAuthInterceptor),
	)
	RegisterAdminServer(server, micro.GetAdminService())
	RegisterBizServer(server, micro.GetBizService())
	go func() {
		server.Serve(lis)
	}()
	go func() {
		<-ctx.Done()
		server.Stop()
	}()
	return nil
}

func (s *AdminService) Logging(n *Nothing, out Admin_LoggingServer) error {
	out.SendMsg(status.Error(codes.Unauthenticated, "method not allowed"))
	return nil
}

func (s *AdminService) Statistics(interval *StatInterval, out Admin_StatisticsServer) error {
	return nil
}

func (b *BizService) Check(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (b *BizService) Add(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (b *BizService) Test(context.Context, *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func parseACL(data string) (map[string][]string, error) {
	out := make(map[string][]string)
	err := json.Unmarshal([]byte(data), &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func allowedMethod(method string, list []string) bool {
	for _, m := range list {
		parts := strings.Split(m, "/")[1:]
		methodParts := strings.Split(method, "/")[1:]
		if parts[0] == methodParts[0] && (parts[1] == methodParts[1] || parts[1] == "*")  {
			return true
		}
	}
	return false
}
