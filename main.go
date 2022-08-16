package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/google/gopacket/pcap"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	"http_service/upload"

	"fmt"
	"io"

	"log"
)

var client upload.Client

func main() {
	addr := *flag.String("address", "localhost:8080", "http server address")
	saddr := *flag.String("sender_address", ":5005", "drpc serder address")
	conn, err := grpc.Dial(saddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	client = upload.NewClient(conn)
	HandleHTTP(addr)
}

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

func HandleHTTP(addr string) {

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
	f, err := os.Open(fileHeader.Filename)
	defer f.Close()

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

	handle, err := pcap.OpenOfflineFile(f)
	defer handle.Close()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Couldn't open pcap"))
		return
	}
	packets := []upload.Packet{}

	for {
		data, ci, err := handle.ReadPacketData()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		cap := upload.Capture{
			ci.Timestamp,
			ci.CaptureLength,
			ci.Length,
			ci.InterfaceIndex,
			ci.AncillaryData,
		}

		packets = append(packets, upload.Packet{cap, data})

	}

	name, err := client.Upload(context.Background(), packets)
	if err != nil {
		log.Fatalln(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(name))

	return

}
