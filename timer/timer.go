package timer

import (
	"context"
	"das-pay/chain/chain_ckb"
	"das-pay/chain/chain_evm"
	"das-pay/chain/chain_sign"
	"das-pay/chain/chain_tron"
	"das-pay/config"
	"das-pay/dao"
	"das-pay/parser"
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/robfig/cron/v3"
	"github.com/scorpiotzh/mylog"
	"sync"
	"time"
)

type DasTimer struct {
	ChainEth      *chain_evm.ChainEvm
	ChainBsc      *chain_evm.ChainEvm
	ChainPolygon  *chain_evm.ChainEvm
	ChainTron     *chain_tron.ChainTron
	ChainCkb      *chain_ckb.ChainCkb
	SignClient    *chain_sign.RemoteSignClient
	TxBuilderBase *txbuilder.DasTxBuilderBase
	DasCore       *core.DasCore
	DbDao         *dao.DbDao
	Ctx           context.Context
	Wg            *sync.WaitGroup
	cron          *cron.Cron
}

var (
	log = mylog.NewLogger("timer", mylog.LevelDebug)
)

func (d *DasTimer) InitChain(kp *parser.KitParser) {
	if kp.ParserCkb != nil {
		d.ChainCkb = kp.ParserCkb.ChainCkb
	}
	if kp.ParserTron != nil {
		d.ChainTron = kp.ParserTron.ChainTron
	}
	if kp.ParserEth != nil {
		d.ChainEth = kp.ParserEth.ChainEvm
	}
	if kp.ParserPolygon != nil {
		d.ChainPolygon = kp.ParserPolygon.ChainEvm
	}
	if kp.ParserBsc != nil {
		d.ChainBsc = kp.ParserBsc.ChainEvm
	}
}

func (d *DasTimer) Run() error {
	if err := d.doUpdateTokenMap(); err != nil {
		return fmt.Errorf("doUpdateTokenMap err: %s", err.Error())
	}
	tickerHedge := time.NewTicker(time.Minute * 1)
	tickerToken := time.NewTicker(time.Minute * 2)
	tickerNode := time.NewTicker(time.Minute * 3)
	tickerDasInfo := time.NewTicker(time.Hour)
	tickerRejected := time.NewTicker(time.Minute * 20)
	tickerNormalCell := time.NewTicker(time.Minute * 20)

	d.Wg.Add(1)
	go func() {
		for {
			select {
			case <-tickerHedge.C:
				if config.Cfg.Server.HedgeUrl != "" {
					log.Info("doOrderHedge start ...")
					if err := d.doOrderHedge(); err != nil {
						log.Error("doOrderHedge err: ", err.Error())
					}
					log.Info("doOrderHedge end ...")
				}
			case <-tickerDasInfo.C:
				log.Info("doDasInfo start ...")
				if err := d.doDasInfo(); err != nil {
					log.Error("doDasInfo err: ", err.Error())
				}
				log.Info("doDasInfo end ...")
			case <-tickerRejected.C:
				log.Info("doRejected start ...")
				if err := d.doRejected(); err != nil {
					log.Error("doRejected err: ", err.Error())
				}
				log.Info("doRejected end ...")
			case <-tickerNormalCell.C:
				log.Info("doNormalCell start ...")
				if err := d.doNormalCell(); err != nil {
					log.Error("doNormalCell err: ", err.Error())
				}
				log.Info("doNormalCell end ...")
			case <-tickerToken.C:
				if err := d.doUpdateTokenMap(); err != nil {
					log.Error("doUpdateTokenMap err:", err)
				}
			case <-tickerNode.C:
				d.doNodeCheck()
			case <-d.Ctx.Done():
				log.Warn("timer done")
				d.Wg.Done()
				return
			default:
				time.Sleep(time.Second)
			}
		}
	}()
	return nil
}

func (d *DasTimer) DoOrderRefund() error {
	if config.Cfg.Server.CronSpec == "" {
		return nil
	}
	log.Info("DoOrderRefund:", config.Cfg.Server.CronSpec)
	d.cron = cron.New(cron.WithSeconds())
	_, err := d.cron.AddFunc(config.Cfg.Server.CronSpec, func() {
		log.Info("doOrderRefund start ...")
		if err := d.doOrderRefund(); err != nil {
			log.Error("doOrderRefund err: ", err.Error())
		}
		log.Info("doOrderRefund end ...")
	})
	if err != nil {
		return fmt.Errorf("c.AddFunc err: %s", err.Error())
	}
	d.cron.Start()
	return nil
}

func (d *DasTimer) CloseCron() {
	log.Warn("cron done")
	if d.cron != nil {
		d.cron.Stop()
	}
}
