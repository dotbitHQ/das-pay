package timer

import (
	"das-pay/config"
	"das-pay/notify"
	"fmt"
)

var (
	ethBlockNumber     uint64
	bscBlockNumber     uint64
	polygonBlockNumber uint64
	ckbBlockNumber     uint64
	tronBlockNumber    int64
)

func (d *DasTimer) doNodeCheck() {
	d.doNodeCheckEth()
	d.doNodeCheckBsc()
	d.doNodeCheckPolygon()
	d.doNodeCheckCKb()
	d.doNodeCheckTron()
}

func (d *DasTimer) doNodeCheckEth() {
	if d.ChainEth == nil {
		return
	}
	bn, err := d.ChainEth.BestBlockNumber()
	if err != nil {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "node check", notify.GetLarkTextNotifyStr("BestBlockNumber", "ETH", err.Error()))
	} else if bn <= ethBlockNumber {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "node check", notify.GetLarkTextNotifyStr("BestBlockNumber", "ETH", fmt.Sprintf("block number: %d", bn)))
	} else {
		log.Info("doNodeCheckEth:", ethBlockNumber, bn)
		ethBlockNumber = bn
	}
}

func (d *DasTimer) doNodeCheckBsc() {
	if d.ChainBsc == nil {
		return
	}
	bn, err := d.ChainBsc.BestBlockNumber()
	if err != nil {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "node check", notify.GetLarkTextNotifyStr("BestBlockNumber", "BSC", err.Error()))
	} else if bn <= bscBlockNumber {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "node check", notify.GetLarkTextNotifyStr("BestBlockNumber", "BSC", fmt.Sprintf("block number: %d", bn)))
	} else {
		log.Info("doNodeCheckBsc:", bscBlockNumber, bn)
		bscBlockNumber = bn
	}
}

func (d *DasTimer) doNodeCheckPolygon() {
	if d.ChainPolygon == nil {
		return
	}
	bn, err := d.ChainPolygon.BestBlockNumber()
	if err != nil {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "node check", notify.GetLarkTextNotifyStr("BestBlockNumber", "POLYGON", err.Error()))
	} else if bn <= polygonBlockNumber {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "node check", notify.GetLarkTextNotifyStr("BestBlockNumber", "POLYGON", fmt.Sprintf("block number: %d", bn)))
	} else {
		log.Info("doNodeCheckPolygon:", polygonBlockNumber, bn)
		polygonBlockNumber = bn
	}
}

func (d *DasTimer) doNodeCheckCKb() {
	if d.ChainCkb == nil {
		return
	}
	bn, err := d.ChainCkb.GetTipBlockNumber()
	if err != nil {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "node check", notify.GetLarkTextNotifyStr("GetTipBlockNumber", "CKB", err.Error()))
	} else if bn <= ckbBlockNumber {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "node check", notify.GetLarkTextNotifyStr("GetTipBlockNumber", "CKB", fmt.Sprintf("block number: %d", bn)))
	} else {
		log.Info("doNodeCheckCKb:", ckbBlockNumber, bn)
		ckbBlockNumber = bn
	}
}

func (d *DasTimer) doNodeCheckTron() {
	if d.ChainTron == nil {
		return
	}
	bn, err := d.ChainTron.GetBlockNumber()
	if err != nil {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "node check", notify.GetLarkTextNotifyStr("GetBlockNumber", "TRON", err.Error()))
	} else if bn <= tronBlockNumber {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "node check", notify.GetLarkTextNotifyStr("GetBlockNumber", "TRON", fmt.Sprintf("block number: %d", bn)))
	} else {
		log.Info("doNodeCheckTron:", tronBlockNumber, bn)
		tronBlockNumber = bn
	}
}
