package main

import (
	"bytes"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
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

func (s *Server) UploadImages(stream pb.ImagePDFService_UploadImagesServer) error {
	var fileBuff []byte
	var rcs []io.ReadCloser
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("failed unexpectadely while reading chunkd from stream: %v", err)
			return err
		}

		fileBuff = append(fileBuff, chunk.Content...)
		if chunk.IsLastChunk {
			rcs = append(rcs, io.NopCloser(bytes.NewReader(fileBuff)))
			fileBuff = []byte{}
		}
	}

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

	grpcServer := grpc.NewServer(grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(logFileSizeInterceptor)))
	pb.RegisterImagePDFServiceServer(grpcServer, new(Server))

	log.Printf("starting server on %s", lsn.Addr().String())
	if err := grpcServer.Serve(lsn); err != nil {
		log.Fatal(err)
	}
}

func logFileSizeInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	stream := &wrappedServerStream{ServerStream: ss}

	err := handler(srv, stream)
	if err != nil {
		return err
	}
	log.Printf("Total file size received: %d bytes", stream.totalSize)
	return nil
}

type wrappedServerStream struct {
	grpc.ServerStream
	totalSize int64
}

func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)
	if err == nil {
		if chunk, ok := m.(*pb.Chunk); ok {
			w.totalSize += int64(len(chunk.Content))
		}
	}
	return err
}
