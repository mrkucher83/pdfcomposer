package main

import (
	"bytes"
	"context"
	"github.com/mrkucher83/pdfcomposer/pdfcompose/pb"
	"github.com/mrkucher83/pdfcomposer/pkg/composer"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
)

type Server struct {
	pb.UnimplementedImagePDFServiceServer
}

func (s *Server) UploadImages(ctx context.Context, req *pb.ImageUploadRequest) (*pb.PDFResponse, error) {
	var rcs []io.ReadCloser
	for _, img := range req.Images {
		rc := io.NopCloser(bytes.NewReader(img.Content))
		rcs = append(rcs, rc)
	}

	pdf, err := composer.ComposeFromFiles(rcs)
	if err != nil {
		return nil, err
	}

	return &pb.PDFResponse{PdfContent: pdf.Bytes()}, nil
}

func main() {
	lsn, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterImagePDFServiceServer(grpcServer, new(Server))

	log.Printf("starting server on %s", lsn.Addr().String())
	if err := grpcServer.Serve(lsn); err != nil {
		log.Fatal(err)
	}
}
