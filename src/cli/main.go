package main

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"os"
	"snoop/src/cli/handlers"
	"snoop/src/shared/cfg"
	"snoop/src/shared/protobuf"
)

const ( // API
	history = "history"
	resend = "resend"
)

var snoopd protobuf.SnoopdAPIClient

func main() {
	snoopdAddr := cfg.GetString("cli.snoopd_ip") + ":" + cfg.GetString("snoopd.grpc_port")
	grpcConn, err := grpc.Dial(snoopdAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Unable to connect to snoopd on addr", snoopdAddr + ",", "err:", err)
	}
	defer grpcConn.Close()

	snoopd = protobuf.NewSnoopdAPIClient(grpcConn)
	if len(os.Args) < 2 {
		fmt.Println("Error: to use snoop one must specify a command")
		os.Exit(1)
	}
	args := os.Args[1:]

	switch args[0] {
	case history:
		handlers.HandleHistory(snoopd)
	case resend:
		if len(args) <= 1 {
			fmt.Println("Error: request hash must be passed for resend function")
			os.Exit(1)
		}
		handlers.HandleResend(snoopd, args[1])
	default:
		fmt.Println("Error: invalid command")
		os.Exit(1)
	}
}
