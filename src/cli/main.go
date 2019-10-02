package main

import (
	"google.golang.org/grpc"
	"snoop/src/shared/cfg"
)

func main() {
	grpcConn, err := grpc.Dial(
		cfg.GetString,
		grpc.WithInsecure(),
	)
	if err != nil {
		logger.Fatal("Can't connect to auth microservice via grpc")
	}
	userManager = services.NewUserMSClient(grpcConn)
}
