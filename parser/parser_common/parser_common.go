package parser_common

import (
	"context"
	"das-pay/dao"
	"das-pay/tables"
	"github.com/scorpiotzh/mylog"
	"sync"
)

var log = mylog.NewLogger("parser_common", mylog.LevelDebug)

type ParserCommon struct {
	Ctx                context.Context
	Wg                 *sync.WaitGroup
	DbDao              *dao.DbDao
	ParserType         tables.ParserType
	Address            string
	CurrentBlockNumber uint64
	ConcurrencyNum     uint64
	ConfirmNum         uint64
}

func (p *ParserCommon) CheckFork(parentHash string) (bool, error) {
	block, err := p.DbDao.FindBlockInfoByBlockNumber(p.ParserType, p.CurrentBlockNumber-1)
	if err != nil {
		return false, err
	}
	if block.Id == 0 {
		return false, nil
	} else if block.BlockHash != parentHash {
		log.Warn("CheckFork:", p.CurrentBlockNumber, parentHash, block.BlockHash)
		return true, nil
	}
	return false, nil
}

func (p *ParserCommon) InitCurrentBlockNumber(currentBlockNumber uint64) error {
	if block, err := p.DbDao.FindBlockInfo(p.ParserType); err != nil {
		return err
	} else if block.Id > 0 {
		p.CurrentBlockNumber = block.BlockNumber
	} else if p.CurrentBlockNumber == 0 && currentBlockNumber > 0 {
		p.CurrentBlockNumber = currentBlockNumber
	}
	return nil
}
