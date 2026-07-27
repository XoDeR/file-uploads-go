.PHONY: run test run-s3

run:
	go run ./cmd/server

test:
	go test ./...

run-s3:
	@echo "Build with S3: go get aws deps then: go run -tags s3 ./cmd/server"
	go run -tags s3 ./cmd/server
