package handlers

import (
	"context"
	"fmt"
	"os"
	"snoop/src/shared/protobuf"
)

func HandleResend(snoopd protobuf.SnoopdAPIClient, hash string) {
	resend, err := snoopd.Resend(context.Background(), &protobuf.ReqID{ID: hash})
	if err != nil {
		fmt.Println("Error: snoopd returned an error:", err)
		os.Exit(1)
	}

	fmt.Print(resend.Response)
}
