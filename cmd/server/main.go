package main

import (
	"bytes"
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

//func (s *Server) UploadImages(ctx context.Context, req *pb.ImageUploadRequest) (*pb.PDFResponse, error) {
//	var rcs []io.ReadCloser
//	for _, img := range req.Images {
//		rc := io.NopCloser(bytes.NewReader(img.Content))
//		rcs = append(rcs, rc)
//	}
//
//	pdf, err := composer.ComposeFromFiles(rcs)
//	if err != nil {
//		return nil, err
//	}
//
//	return &pb.PDFResponse{PdfContent: pdf.Bytes()}, nil
//}

func (s *Server) UploadImages(stream pb.ImagePDFService_UploadImagesServer) (err error) {
	var fileBuff []byte
	var rcs []io.ReadCloser
	for {
		res, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				goto END
			}

			log.Printf("failed unexpectadely while reading chunkd form stream: %v", err)
			return err
		}
		fileBuff = append(fileBuff, res.Content...)
	}

END:
	rcs = append(rcs, io.NopCloser(bytes.NewReader(fileBuff)))
	pdf, err := composer.ComposeFromFiles(rcs)
	if err != nil {
		return err
	}
	err = stream.SendAndClose(&pb.PDFResponse{PdfContent: pdf.Bytes()})
	if err != nil {
		log.Printf("failed to send pdf file to client: %v", err)
		return err
	}

	return nil
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
