package core

import (
	"fmt"
	"net"

	service "github.com/jianbo-zh/dchat/bind/grpc"
	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/dchat/cuckoo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type ServiceWrapper struct{}

func (s *ServiceWrapper) GetCuckoo() (*cuckoo.Cuckoo, error) {
	if node == nil {
		return nil, fmt.Errorf("node is nil, maybe not started")
	}

	return node.GetCuckoo()
}

func StartGrpcWithPort(hostport string) {
	go startGrpc("tcp", hostport)
}

func StartGrpc(socketpath string) {
	go startGrpc("unix", socketpath)
}

func startGrpc(scheme string, socketpath string) {

	listen, err := net.Listen(scheme, socketpath)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()

	serviceWrapper := &ServiceWrapper{}

	proto.RegisterConfigSvcServer(s, service.NewConfigSvc(serviceWrapper))
	proto.RegisterAccountSvcServer(s, service.NewAccountSvc(serviceWrapper))
	proto.RegisterContactSvcServer(s, service.NewContactSvc(serviceWrapper))
	proto.RegisterGroupSvcServer(s, service.NewGroupSvc(serviceWrapper))
	proto.RegisterSessionSvcServer(s, service.NewSessionSvc(serviceWrapper))
	proto.RegisterSystemSvcServer(s, service.NewSystemSvc(serviceWrapper))
	proto.RegisterSubscribeSvcServer(s, service.NewSubscribeSvc(serviceWrapper))

	reflection.Register(s)

	defer func() {
		s.Stop()
		listen.Close()
	}()

	err = s.Serve(listen)
	if err != nil {
		fmt.Printf("failed to serve: %v", err)
		panic(err)
	}
}
