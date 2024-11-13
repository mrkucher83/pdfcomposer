generate:
	mkdir -p pdfcompose/pb

	protoc \
		--proto_path=api/ \
		--go_out=pdfcompose/pb \
		--go-grpc_out=pdfcompose/pb \
		api/*.proto
