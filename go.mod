module das-pay

go 1.16

require (
	github.com/dotbitHQ/das-lib v1.0.1-0.20220811023155-80cb374d8e28
	github.com/elazarl/goproxy v0.0.0-20220529153421-8ea89ba92021 // indirect
	github.com/ethereum/go-ethereum v1.10.17
	github.com/fbsobreira/gotron-sdk v0.0.0-20211102183839-58a64f4da5f4
	github.com/fsnotify/fsnotify v1.5.4
	github.com/golang/protobuf v1.5.2
	github.com/nervosnetwork/ckb-sdk-go v0.101.3
	github.com/parnurzeal/gorequest v0.2.16
	github.com/robfig/cron/v3 v3.0.1
	github.com/scorpiotzh/mylog v1.0.10
	github.com/scorpiotzh/toolib v1.1.3
	github.com/shopspring/decimal v1.3.1
	github.com/urfave/cli/v2 v2.5.0
	google.golang.org/grpc v1.46.0
	gorm.io/gorm v1.22.1
	moul.io/http2curl v1.0.0 // indirect
)

replace (
	github.com/btcsuite/btcd v0.22.0-beta => github.com/btcsuite/btcd v0.23.1
	github.com/btcsuite/btcd v0.22.0-beta.0.20220111032746-97732e52810c => github.com/btcsuite/btcd v0.23.1
)
