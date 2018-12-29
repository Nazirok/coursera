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
	"io"
	"strings"
	"sync"
	"time"
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
		sync.RWMutex
		logID      int
		statID     int
		logStreams map[int]chan *Event
		stats      map[int]*Stat
	}

	BizService struct {
	}
)

func NewMicroService(acl map[string][]string) *MicroService {
	return &MicroService{
		acl: acl,
		admin: &AdminService{
			logStreams: make(map[int]chan *Event),
			stats:      make(map[int]*Stat),
			// stats: &Stat{
			// 	ByMethod:   make(map[string]uint64),
			// 	ByConsumer: make(map[string]uint64),
			// },
		},
		biz: &BizService{},
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
	md, _ := metadata.FromIncomingContext(ctx)
	consumer := md.Get("consumer")
	e, err := m.authorize(consumer, info.FullMethod)
	if err != nil {
		return nil, err
	}
	m.admin.addStat(consumer[0], info.FullMethod)
	m.admin.writeLog(e)
	return handler(ctx, req)
}

func (m *MicroService) streamAuthInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	md, _ := metadata.FromIncomingContext(ss.Context())
	consumer := md.Get("consumer")
	e, err := m.authorize(consumer, info.FullMethod)
	if err != nil {
		return err
	}
	m.admin.addStat(consumer[0], info.FullMethod)
	m.admin.writeLog(e)
	return handler(srv, ss)
}

func (m *MicroService) authorize(consumer []string, method string) (*Event, error) {
	if len(consumer) == 0 {
		return nil, status.Error(codes.Unauthenticated, "consumer not found")
	}
	list, ok := m.acl[consumer[0]]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unknown consumer")
	}
	if !allowedMethod(method, list) {
		return nil, status.Error(codes.Unauthenticated, "method not allowed")
	}
	return &Event{
		Timestamp: 0,
		Host:      "127.0.0.1:8089",
		Consumer:  consumer[0],
		Method:    method,
	}, nil
}

func (a *AdminService) addLogChan() {
	a.logID++
	a.logStreams[a.logID] = make(chan *Event, 100)
}

func (a *AdminService) getLastLogStream() chan *Event {
	return a.logStreams[a.logID]
}

func (a *AdminService) writeLog(e *Event) {
	for _, v := range a.logStreams {
		select {
		case v <- e:
		default:
			return
		}
	}
}

func (a AdminService) addStat(consumer, method string) {
	for _, v := range a.stats {
		a.RLock()
		v.ByConsumer[consumer]++
		v.ByMethod[method]++
		a.RUnlock()
	}

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

func (a *AdminService) Logging(n *Nothing, out Admin_LoggingServer) error {
	a.addLogChan()
	for e := range a.getLastLogStream() {
		err := out.Send(e)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AdminService) Statistics(interval *StatInterval, out Admin_StatisticsServer) error {
	fmt.Println(out.Context())
	ticker := time.NewTicker(time.Duration(interval.IntervalSeconds) * time.Second)
	select {
	case <- ticker.C:
		err := out.Send(a.stats)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	case <- out.Context().Done():
		ticker.Stop()
		return nil
	}
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
		if parts[0] == methodParts[0] && (parts[1] == methodParts[1] || parts[1] == "*") {
			return true
		}
	}
	return false
}
