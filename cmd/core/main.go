package main

import (
	"flag"
	"fmt"
	"github.com/pinshare/config"
	"github.com/pinshare/core/service"
	"google.golang.org/grpc"
	"net"

	"os"
	"os/signal"
	"syscall"
)

// type addPinServer struct{}
//
// func (s *addPinServer) Execute(ctx context.Context, req *entity.AddRequest) (*entity.PinResponse, error) {
// 	fmt.Println(req.String())
// 	return &entity.PinResponse{
// 		Id:          0,
// 		UserId:      req.UserId,
// 		Title:       req.Title,
// 		Url:         req.Url,
// 		Timestamp:   req.Timestamp,
// 		Description: req.Description,
// 		Tags:        req.Tags,
// 	}, nil
// }
//
// type updatePinServer struct{}
//
// func (s *updatePinServer) Execute(ctx context.Context, req *entity.UpdateRequest) (*entity.PinResponse, error) {
// 	fmt.Println(req.String())
// 	return &entity.PinResponse{
// 		Id:          req.Id,
// 		UserId:      req.UserId,
// 		Title:       req.Title,
// 		Url:         req.Url,
// 		Timestamp:   req.Timestamp,
// 		Description: req.Description,
// 		Tags:        req.Tags,
// 	}, nil
// }
//
// type deletePinServer struct{}
//
// func (s *deletePinServer) Execute(ctx context.Context, req *entity.DeleteRequest) (*entity.PinDeleteResponse, error) {
// 	fmt.Println(req.String())
// 	return &entity.PinDeleteResponse{}, nil
// }

type cliOptions struct {
	Host       string
	Port       int
	ConfigPath string
}

var cli = cliOptions{}

func init() {
	flag.IntVar(&cli.Port, "p", 5001, "Listen Port Number")
	flag.StringVar(&cli.Host, "h", "0.0.0.0", "Listen Host IP")
	flag.StringVar(&cli.ConfigPath, "c", "/etc/pinshare.conf", "Config path")
	flag.Parse()
}

func main() {
	c, err := config.Init(cli.ConfigPath)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	listen := fmt.Sprintf("%s:%d", cli.Host, cli.Port)
	socket, err := net.Listen("tcp", listen)
	if err != nil {
		fmt.Println("Error", err)
		return
	}
	fmt.Println("Listen:", listen, "...")

	server := grpc.NewServer()
	for _, s := range service.GetServices() {
		fmt.Printf("Add service: %s\n", s.Name())
		s.Register(server, c)
	}
	go server.Serve(socket)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	server.GracefulStop()

	// entity.RegisterAddPinServer(server, &addPinServer{})
	// entity.RegisterUpdatePinServer(server, &updatePinServer{})
	// entity.RegisterDeletePinServer(server, &deletePinServer{})
	// server.Serve(socket)
}
