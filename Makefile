# build file
GOCMD=go
# Use -a flag to prevent code cache problems.
GOBUILD=$(GOCMD) build -mod=vendor -ldflags -s -v -i

pay: BIN_BINARY_NAME=das_pay_server
pay:
	GO111MODULE=on $(GOBUILD) -o $(BIN_BINARY_NAME) cmd/main.go
	mv $(BIN_BINARY_NAME) bin/
	@echo "Build $(BIN_BINARY_NAME) successfully. You can run bin/$(BIN_BINARY_NAME) now.If you can't see it soon,wait some seconds"


update:
	go env -w GOPRIVATE="github.com/DeAccountSystems"
	go mod tidy
	go mod vendor
