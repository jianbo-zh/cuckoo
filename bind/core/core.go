package core

import (
	"fmt"
	"net"

	service "github.com/jianbo-zh/dchat/bind/grpc"
	"github.com/jianbo-zh/dchat/bind/grpc/proto"
	"github.com/jianbo-zh/go-anet"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func StartGrpcWithPort(nativedriver NativeNetDriver, hostport string) {
	logx, _ := zap.NewDevelopment()

	inet := &inet{
		net:    nativedriver,
		logger: logx,
	}

	anet.SetNetDriver(inet)

	interfaces, err := anet.Interfaces()
	if err != nil {
		fmt.Println("error: ", err)
	}
	fmt.Println("Interfaces", interfaces)

	interfaceAddr, err := anet.InterfaceAddrs()
	if err != nil {
		fmt.Println("error: ", err)
	}
	fmt.Println("InterfaceAddrs: ", interfaceAddr)

	go startGrpc("tcp", hostport)
}

func StartGrpc(nativedriver NativeNetDriver, socketpath string) {

	logx, _ := zap.NewDevelopment()

	inet := &inet{
		net:    nativedriver,
		logger: logx,
	}

	anet.SetNetDriver(inet)

	interfaces, err := anet.Interfaces()
	if err != nil {
		fmt.Println("error: ", err)
	}
	fmt.Println("Interfaces", interfaces)

	interfaceAddr, err := anet.InterfaceAddrs()
	if err != nil {
		fmt.Println("error: ", err)
	}
	fmt.Println("InterfaceAddrs: ", interfaceAddr)

	go startGrpc("unix", socketpath)
}

func startGrpc(scheme string, socketpath string) {

	listen, err := net.Listen(scheme, socketpath)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()

	proto.RegisterAccountSvcServer(s, &service.AccountSvc{})
	proto.RegisterContactSvcServer(s, &service.ContactSvc{})
	proto.RegisterGroupSvcServer(s, &service.GroupSvc{})
	proto.RegisterSessionSvcServer(s, &service.SessionSvc{})
	proto.RegisterSystemSvcServer(s, &service.SystemSvc{})

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
