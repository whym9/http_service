package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

func main() {

	fileName := *flag.String("name", "lo.pcapng", "path to the file")
	addr := *flag.String("address", "http://localhost:8080/", "address of the GRPC server")
	flag.Parse()
	err := call(addr, "POST", fileName)
	if err != nil {
		log.Fatal(err)
		return
	}

}

func call(urlPath, method, filename string) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("uploadFile", filename)
	if err != nil {

		return err
	}
	file, err := os.Open(filename)
	if err != nil {

		return err
	}
	_, err = io.Copy(fw, file)
	if err != nil {

		return err
	}
	writer.Close()
	req, err := http.NewRequest(method, urlPath, bytes.NewReader(body.Bytes()))
	if err != nil {
		fmt.Println(".request")
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rsp, _ := client.Do(req)
	ans := make([]byte, 1024)
	rsp.Body.Read(ans)
	fmt.Println(string(ans))

	if rsp.StatusCode != http.StatusOK {
		log.Printf("Request failed with response code: %d", rsp.StatusCode)

		return nil
	}
	return nil
}
