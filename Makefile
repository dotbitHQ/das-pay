# build file
GO_BUILD=go build -ldflags -s -v

pay: BIN_BINARY_NAME=das_pay_server
pay:
	GO111MODULE=on $(GO_BUILD) -o $(BIN_BINARY_NAME) cmd/main.go
	@echo "Build $(BIN_BINARY_NAME) successfully. You can run ./$(BIN_BINARY_NAME) now.If you can't see it soon,wait some seconds"

refund: BIN_BINARY_NAME=das_refund_server
refund:
	GO111MODULE=on $(GO_BUILD) -o $(BIN_BINARY_NAME) cmd/refund/main.go
	@echo "Build $(BIN_BINARY_NAME) successfully. You can run ./$(BIN_BINARY_NAME) now.If you can't see it soon,wait some seconds"

update:
	go mod tidy

docker:
	docker build --network host -t dotbitteam/das-pay:latest .

docker-publish:
	docker image push dotbitteam/das-pay:latest
