module das-pay

go 1.16

require (
	github.com/btcsuite/btcd v0.23.0
	github.com/dotbitHQ/das-lib v1.0.1-0.20230321033933-9c3f96663bbe
	github.com/ethereum/go-ethereum v1.10.17
	github.com/fbsobreira/gotron-sdk v0.0.0-20211102183839-58a64f4da5f4
	github.com/fsnotify/fsnotify v1.5.4
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/nervosnetwork/ckb-sdk-go v0.101.3
	github.com/parnurzeal/gorequest v0.2.16
	github.com/robfig/cron/v3 v3.0.1
	github.com/scorpiotzh/mylog v1.0.10
	github.com/scorpiotzh/toolib v1.1.3
	github.com/shopspring/decimal v1.3.1
	github.com/stretchr/testify v1.7.1 // indirect
	github.com/urfave/cli/v2 v2.5.0
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/text v0.3.8-0.20211105212822-18b340fc7af2 // indirect
	google.golang.org/grpc v1.46.0
	gorm.io/gorm v1.22.1
)

replace (
	github.com/btcsuite/btcd v0.22.0-beta => github.com/btcsuite/btcd v0.23.1
	github.com/btcsuite/btcd v0.22.0-beta.0.20220111032746-97732e52810c => github.com/btcsuite/btcd v0.23.1
)
