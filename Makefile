# build file
GOCMD=go
# Use -a flag to prevent code cache problems.
GOBUILD=$(GOCMD) build -mod=vendor -ldflags -s -v -i

pay: BIN_BINARY_NAME=das_pay_server
pay:
	GO111MODULE=on $(GOBUILD) -o $(BIN_BINARY_NAME) cmd/main.go
	@echo "Build $(BIN_BINARY_NAME) successfully. You can run ./$(BIN_BINARY_NAME) now.If you can't see it soon,wait some seconds"

refund: BIN_BINARY_NAME=das_refund_server
refund:
	GO111MODULE=on $(GOBUILD) -o $(BIN_BINARY_NAME) cmd/refund/main.go
	@echo "Build $(BIN_BINARY_NAME) successfully. You can run ./$(BIN_BINARY_NAME) now.If you can't see it soon,wait some seconds"

update:
	go mod tidy
	go mod vendor
