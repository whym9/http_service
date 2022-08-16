package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/grpc"

	uploadpb "http_service/proto"
)

type Capture struct {
	TimeStamp      time.Time     `json: "time"`
	CaptureLength  int           `json: "caplength"`
	Length         int           `json: "length"`
	InterfaceIndex int           `json :  "index"`
	AccalaryData   []interface{} `json: "accalary"`
}

type Packet struct {
	Ci   Capture
	Data []byte
}

type Client struct {
	client uploadpb.UploadServiceClient
}

func NewClient(conn grpc.ClientConnInterface) Client {
	return Client{
		client: uploadpb.NewUploadServiceClient(conn),
	}
}

func (c Client) Upload(con context.Context, packets []Packet) (string, error) {
	ctx, cancel := context.WithDeadline(con, time.Now().Add(10*time.Second))
	defer cancel()

	stream, err := c.client.Upload(ctx)
	if err != nil {

		return "", err
	}

	for _, pack := range packets {

		b, err := json.Marshal(&pack.Ci)

		if err != nil {

			return "", err
		}

		if err := stream.Send(&uploadpb.UploadRequest{Chunk: b}); err != nil {

			return "", err
		}

		if err := stream.Send(&uploadpb.UploadRequest{Chunk: pack.Data}); err != nil {

			return "", err
		}

	}

	res, err := stream.CloseAndRecv()
	if err != nil {

		return "", err
	}
	fmt.Println("stopped sending")
	fmt.Println(res.GetName())

	return res.GetName(), nil
}
