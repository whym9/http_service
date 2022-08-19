package upload

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	uploadpb "http_service/internal/proto"
)

type Client struct {
	client uploadpb.UploadServiceClient
}

func NewClient(conn grpc.ClientConnInterface) Client {
	return Client{
		client: uploadpb.NewUploadServiceClient(conn),
	}
}

func (c Client) Upload(con context.Context, file []byte) (string, error) {
	fmt.Println("Sending...")
	ctx, cancel := context.WithDeadline(con, time.Now().Add(10*time.Second))
	defer cancel()

	stream, err := c.client.Upload(ctx)
	if err != nil {

		return "", err
	}
	en := 1024
	be := 0
	for {

		if en > len(file) {
			if err := stream.Send(&uploadpb.UploadRequest{Chunk: file[be:]}); err != nil {

				return "", err
			}
			break
		}

		if err := stream.Send(&uploadpb.UploadRequest{Chunk: file[be:en]}); err != nil {

			return "", err
		}

		be = en
		en += 1024

	}

	res, err := stream.CloseAndRecv()
	if err != nil {

		return "", err
	}
	fmt.Println("stopped sending")

	return res.GetName(), nil
}
