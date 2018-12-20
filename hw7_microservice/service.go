package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"encoding/json"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type (
	AdminService struct {
	}

	BizService struct {
	}
)

func NewServiceManager() *AdminService {
	return &AdminService{}
}

func NewBizService() *BizService {
	return &BizService{}
}

func StartMyMicroservice(ctx context.Context, conn string, acl string) error {
	access, err := parseACL(acl)
	if err != nil {
		return err
	}
	fmt.Println(access)
	lis, err := net.Listen("tcp", conn)
	if err != nil {
		fmt.Errorf("can`t listen port %s", err)
	}

	server := grpc.NewServer()
	RegisterAdminServer(server, NewServiceManager())
	RegisterBizServer(server, NewBizService())
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