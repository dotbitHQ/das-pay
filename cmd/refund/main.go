package main

import (
	"context"
	"das-pay/chain/chain_sign"
	"das-pay/config"
	"das-pay/dao"
	"das-pay/parser"
	"das-pay/timer"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/sign"
	"github.com/DeAccountSystems/das-lib/txbuilder"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/scorpiotzh/mylog"
	"github.com/scorpiotzh/toolib"
	"github.com/urfave/cli/v2"
	"os"
	"sync"
	"time"
)

var (
	log               = mylog.NewLogger("refund", mylog.LevelDebug)
	exit              = make(chan struct{})
	ctxServer, cancel = context.WithCancel(context.Background())
	wgServer          = sync.WaitGroup{}
)

func main() {
	log.Debugf("server start：")
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Load configuration from `FILE`",
			},
		},
		Action: runServer,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runServer(ctx *cli.Context) error {
	// config file
	configFilePath := ctx.String("config")
	if err := config.InitCfg(configFilePath); err != nil {
		return err
	}

	// config file watcher
	watcher, err := config.AddCfgFileWatcher(configFilePath)
	if err != nil {
		return err
	}
	// ============= service start =============

	// db
	dbDao, err := dao.NewGormDB(config.Cfg.DB.Mysql, config.Cfg.DB.ParserMysql)
	if err != nil {
		return fmt.Errorf("dao.NewGormDB err: %s", err.Error())
	}

	// parser kit
	kp, err := parser.NewKitParser(ctxServer, cancel, &wgServer, dbDao)
	if err != nil {
		return fmt.Errorf("NewKitParser err: %s", err.Error())
	}

	// das init
	dasCore, txBuilderBase, err := initDas(kp.ParserCkb.ChainCkb.Client)
	if err != nil {
		return fmt.Errorf("initDas err: %s", err.Error())
	}
	log.Info("das core ok")

	// timer
	dt := timer.DasTimer{
		ChainEth:      nil,
		ChainBsc:      nil,
		ChainPolygon:  nil,
		ChainTron:     nil,
		ChainCkb:      nil,
		SignClient:    nil,
		TxBuilderBase: txBuilderBase,
		DasCore:       dasCore,
		DbDao:         dbDao,
		Ctx:           ctxServer,
		Wg:            &wgServer,
	}
	dt.InitChain(kp)
	if config.Cfg.Server.RemoteSignApiUrl != "" {
		signClient, err := chain_sign.NewRemoteSignClient(ctxServer, config.Cfg.Server.RemoteSignApiUrl)
		if err != nil {
			return fmt.Errorf("chain_common.NewRemoteSignClient err: %s", err.Error())
		}
		dt.SignClient = signClient
	}
	if err := dt.DoOrderRefund(); err != nil {
		return fmt.Errorf("dt.DoOrderRefund() err: %s", err.Error())
	}

	// ============= service end =============
	toolib.ExitMonitoring(func(sig os.Signal) {
		log.Warn("ExitMonitoring:", sig.String())
		if watcher != nil {
			log.Warn("close watcher ... ")
			_ = watcher.Close()
		}
		dt.CloseCron()
		cancel()
		wgServer.Wait()
		log.Warn("success exit server. bye bye!")
		time.Sleep(time.Second)
		exit <- struct{}{}
	})

	<-exit
	return nil
}

func initDas(client rpc.Client) (*core.DasCore, *txbuilder.DasTxBuilderBase, error) {
	ops := []core.DasCoreOption{
		core.WithClient(client),
		core.WithDasContractArgs(config.Cfg.DasLib.DasContractArgs),
		core.WithDasContractCodeHash(config.Cfg.DasLib.DasContractCodeHash),
		core.WithDasNetType(config.Cfg.Server.Net),
		core.WithTHQCodeHash(config.Cfg.DasLib.THQCodeHash),
	}
	dasCore := core.NewDasCore(ctxServer, &wgServer, ops...)
	dasCore.InitDasContract(config.Cfg.DasLib.MapDasContract)
	if err := dasCore.InitDasConfigCell(); err != nil {
		return nil, nil, fmt.Errorf("InitDasConfigCell err: %s", err.Error())
	}
	dasCore.RunAsyncDasContract(time.Minute * 3)   // contract outpoint
	dasCore.RunAsyncDasConfigCell(time.Minute * 5) // config cell outpoint

	payServerAddressArgs := ""
	if config.Cfg.Chain.Ckb.Address != "" {
		parseAddress, err := address.Parse(config.Cfg.Chain.Ckb.Address)
		if err != nil {
			return nil, nil, fmt.Errorf("address.Parse err: %s", err.Error())
		} else {
			payServerAddressArgs = common.Bytes2Hex(parseAddress.Script.Args)
		}
	}
	var handleSign sign.HandleSignCkbMessage
	if config.Cfg.Chain.Ckb.Private != "" {
		handleSign = sign.LocalSign(config.Cfg.Chain.Ckb.Private)
	} else if config.Cfg.Server.RemoteSignApiUrl != "" && payServerAddressArgs != "" {
		remoteSignClient, err := sign.NewClient(ctxServer, config.Cfg.Server.RemoteSignApiUrl)
		if err != nil {
			return nil, nil, fmt.Errorf("sign.NewClient err: %s", err.Error())
		}
		handleSign = sign.RemoteSign(remoteSignClient, config.Cfg.Server.Net, payServerAddressArgs)
	}
	txBuilderBase := txbuilder.NewDasTxBuilderBase(ctxServer, dasCore, handleSign, payServerAddressArgs)
	return dasCore, txBuilderBase, nil
}
