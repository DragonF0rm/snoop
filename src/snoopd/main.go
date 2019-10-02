package main

import (
	"google.golang.org/grpc"
	"net"
	"snoopd/api"
	"snoopd/cfg"
	"snoopd/log"
	"snoopd/protobuf"
	"snoopd/proxy"
	"strconv"
)

func main() {
	serverName := cfg.GetString("snoopd.name")
	serverVersion := cfg.GetString("snoopd.version")
	var mode string
	if cfg.GetBool("snoopd.debug_mode") {
		mode = "in debug mode"
	} else {
		mode = "in production mode"
	}
	log.Info("Starting", serverName, serverVersion, mode)

	go proxy.ListenAndServe()
	go proxy.ListenAndServeTLS()

	grpcPort := cfg.GetInt("snoopd.grpc_port")
	apiService := api.NewGrpcApiService(cfg.GetString("snoopd.log.access_logger_file"))
	grpcServer := grpc.NewServer()
	protobuf.RegisterSnoopdAPIServer(grpcServer, &apiService)

	listener, err := net.Listen("tcp", ":" + strconv.Itoa(grpcPort))
	if err != nil {
		log.Fatal("Unable to start listening for GRPC, err:", err)
	}
	log.Info("Listening for GRPC on port", grpcPort)

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("Grpc server exited with error:", err)
	}
}
