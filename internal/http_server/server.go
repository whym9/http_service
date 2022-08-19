package httpserver

import (
	"context"
	"fmt"
	"http_service/internal/upload"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var client upload.Client

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

func HandleHTTP(addr, saddr string) {
	conn, err := grpc.Dial(saddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	recordMetrics()

	client = upload.NewClient(conn)
	fmt.Println("HTTP Server has started")
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", uploadFile)
	http.ListenAndServe(addr, nil)

}

func uploadFile(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong request method"))
		return
	}

	if err := r.ParseMultipartForm(100 * 1024 * 1024); err != nil {
		fmt.Printf("could not parse multipart form: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("CANT_PARSE_FORM"))
		return
	}

	file, fileHeader, err := r.FormFile("uploadFile")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Couldn't convert"))
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("INVALID_FILE"))
		return
	}
	defer file.Close()

	fileSize := fileHeader.Size

	fileContent, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("INVALID_FILE"))
		return
	}

	fileType := http.DetectContentType(fileContent)
	if fileType != "application/octet-stream" {
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte("Wrong file type!"))
		if err != nil {
			log.Fatal(err)
			return

		}
		return
	}

	fmt.Printf("FileType: %s, File: %s\n", fileType, fileHeader.Filename)
	fmt.Printf("File size (bytes): %v\n", fileSize)

	name, err := client.Upload(context.Background(), fileContent)
	if err != nil {
		log.Fatalln(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(name))

	return

}
