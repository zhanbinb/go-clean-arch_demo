package main

import (
	"context"
	"fmt"
	"log"
	"time"

	articlev1 "github.com/zhanbinb/go-clean-arch_demo/api/gen/go/article/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient(
		"localhost:9091",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("connect gRPC server: %v", err)
	}
	defer func() { _ = conn.Close() }()

	client := articlev1.NewArticleServiceClient(conn)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	resp, err := client.GetArticleTitle(
		ctx,
		&articlev1.GetArticleTitleRequest{
			Id: 1,
		},
	)
	if err != nil {
		log.Fatalf("GetArticleTitle: %v", err)
	}

	fmt.Printf(
		"article id=%d title=%q\n",
		resp.GetId(),
		resp.GetTitle(),
	)
}
