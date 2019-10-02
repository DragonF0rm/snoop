package handlers

import (
	"context"
	"fmt"
	"os"
	"snoop/src/shared/protobuf"
)

func HandleHistory(snoopd protobuf.SnoopdAPIClient) {
	history, err := snoopd.GetHistory(context.Background(), &protobuf.Nothing{})
	if err != nil {
		fmt.Println("Error: snoopd returned an error:", err)
		os.Exit(1)
	}

	for _, rt := range history.RoundTrips {
		fmt.Println(rt)
	}
}
