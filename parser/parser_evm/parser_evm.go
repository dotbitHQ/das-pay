package parser_evm

import (
	"das-pay/chain/chain_common"
	"das-pay/chain/chain_evm"
	"das-pay/config"
	"das-pay/notify"
	"das-pay/parser/parser_common"
	"das-pay/tables"
	"fmt"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/scorpiotzh/mylog"
	"github.com/scorpiotzh/toolib"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var log = mylog.NewLogger("parser_eth", mylog.LevelDebug)

type ParserEvm struct {
	PayTokenId tables.PayTokenId
	ChainEvm   *chain_evm.ChainEvm
	parser_common.ParserCommon
}

func (p *ParserEvm) Parser() {
	currentBlockNumber, err := p.ChainEvm.BestBlockNumber()
	if err != nil {
		log.Error("BestBlockNumber err: ", p.ParserType.ToString(), err.Error())
		return
	}
	if err := p.InitCurrentBlockNumber(currentBlockNumber); err != nil {
		log.Error("initCurrentBlockNumber err:", p.ParserType.ToString(), err.Error())
		return
	}
	atomic.AddUint64(&p.CurrentBlockNumber, 1)
	p.Wg.Add(1)
	for {
		select {
		default:
			latestBlockNumber, err := p.ChainEvm.BestBlockNumber()
			if err != nil {
				log.Error("BestBlockNumber err:", p.ParserType.ToString(), err.Error())
			} else {
				if p.ConcurrencyNum > 1 && p.CurrentBlockNumber < (latestBlockNumber-p.ConfirmNum-p.ConcurrencyNum) {
					nowTime := time.Now()
					if err = p.parserConcurrencyMode(); err != nil {
						log.Error("parserConcurrencyMode err:", p.ParserType.ToString(), err.Error(), p.CurrentBlockNumber)
						if !strings.Contains(err.Error(), "GetBlockByNumber data is nil") {
							notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, p.ParserType.ToString()+" Parse", notify.GetLarkTextNotifyStr("parserConcurrencyMode", "", err.Error()))
						}
					}
					log.Warn("parserConcurrencyMode time:", p.ParserType.ToString(), time.Since(nowTime).Seconds())
				} else if p.CurrentBlockNumber < (latestBlockNumber - p.ConfirmNum) { // check rollback
					nowTime := time.Now()
					if err = p.parserSubMode(); err != nil {
						log.Error("parserSubMode err:", p.ParserType.ToString(), err.Error(), p.CurrentBlockNumber)
						if !strings.Contains(err.Error(), "GetBlockByNumber data is nil") {
							notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, p.ParserType.ToString()+" Parse", notify.GetLarkTextNotifyStr("parserSubMode", "", err.Error()))
						}
					}
					log.Warn("parserSubMode time:", p.ParserType.ToString(), time.Since(nowTime).Seconds())
				} else {
					log.Info("RunParser:", p.ParserType.ToString(), p.CurrentBlockNumber, latestBlockNumber)
					time.Sleep(time.Second * 10)
				}
				time.Sleep(time.Second)
			}
		case <-p.Ctx.Done():
			log.Warnf("%s parse done", p.ParserType.ToString())
			p.Wg.Done()
			return
		}
	}
}

func (p *ParserEvm) parsingBlockData(block *chain_common.Block) error {
	for _, tx := range block.Transactions {
		switch strings.ToLower(ethcommon.HexToAddress(tx.To).Hex()) {
		case strings.ToLower(p.Address):
			orderId := string(ethcommon.FromHex(tx.Input))
			log.Info("ParsingBlockData:", p.ParserType.ToString(), tx.Hash, tx.From, orderId, tx.Value)
			if orderId == "" {
				continue
			}
			// select order by order id which in tx memo
			order, err := p.DbDao.GetOrderByOrderId(orderId)
			if err != nil {
				return fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
			} else if order.Id == 0 {
				log.Warn("GetOrderByOrderId is not exist:", p.ParserType.ToString(), orderId)
				continue
			}
			if order.PayTokenId != p.PayTokenId {
				log.Warn("order token id not match", order.OrderId, p.PayTokenId)
				continue
			}
			// check value is equal amount or not
			decValue := decimal.NewFromBigInt(chain_common.BigIntFromHex(tx.Value), 0)
			if decValue.Cmp(order.PayAmount) == -1 {
				log.Warn("tx value less than order amount:", p.ParserType.ToString(), decValue, order.PayAmount.String())
				continue
			}
			// change the status to confirm
			payInfo := tables.TableDasOrderPayInfo{
				Id:           0,
				Hash:         tx.Hash,
				OrderId:      order.OrderId,
				ChainType:    p.ParserType.ToChainType(),
				Address:      ethcommon.HexToAddress(tx.From).Hex(),
				Status:       tables.OrderTxStatusConfirm,
				AccountId:    order.AccountId,
				RefundStatus: tables.TxStatusDefault,
				RefundHash:   "",
				Timestamp:    time.Now().UnixNano() / 1e6,
			}
			if err := p.DbDao.UpdatePayStatus(&payInfo); err != nil {
				return fmt.Errorf("UpdatePayStatus err: %s", err.Error())
			}
		}
		continue
	}
	return nil
}

func (p *ParserEvm) parserSubMode() error {
	log.Info("parserSubMode:", p.ParserType.ToString(), p.CurrentBlockNumber)
	block, err := p.ChainEvm.GetBlockByNumber(p.CurrentBlockNumber)
	if err != nil {
		return fmt.Errorf("GetBlockByNumber err: %s", err.Error())
	} else {
		blockHash := block.Hash
		parentHash := block.ParentHash
		if block.Hash == "" || block.ParentHash == "" {
			log.Info("GetBlockByNumber:", p.CurrentBlockNumber, toolib.JsonString(&block))
			return fmt.Errorf("GetBlockByNumber data is nil: [%d]", p.CurrentBlockNumber)
		}
		log.Info("parserSubMode:", p.ParserType.ToString(), blockHash, parentHash)
		if fork, err := p.CheckFork(parentHash); err != nil {
			return fmt.Errorf("CheckFork err: %s", err.Error())
		} else if fork {
			log.Warn("CheckFork is true:", p.ParserType.ToString(), p.CurrentBlockNumber, blockHash, parentHash)
			atomic.AddUint64(&p.CurrentBlockNumber, ^uint64(0))
		} else if err := p.parsingBlockData(block); err != nil {
			//log.Error("ParsingBlockData err:", p.ParserType.ToString(), err.Error())
			return fmt.Errorf("parsingBlockData err: %s", err.Error())
		} else {
			blockInfo := tables.TableBlockParserInfo{
				ParserType:  p.ParserType,
				BlockNumber: p.CurrentBlockNumber,
				BlockHash:   blockHash,
				ParentHash:  parentHash,
			}
			if err = p.DbDao.CreateBlockInfo(&blockInfo); err != nil {
				return fmt.Errorf("CreateBlockInfo err: %s", err.Error())
			} else {
				atomic.AddUint64(&p.CurrentBlockNumber, 1)
			}
			if err = p.DbDao.DeleteBlockInfo(p.ParserType, p.CurrentBlockNumber-20); err != nil {
				return fmt.Errorf("DeleteBlockInfo err: %s", err.Error())
			}
		}
	}
	return nil
}

func (p *ParserEvm) parserConcurrencyMode() error {
	log.Info("parserConcurrencyMode:", p.ParserType.ToString(), p.CurrentBlockNumber)
	var errList = make([]error, p.ConcurrencyNum)
	var blockList = make([]tables.TableBlockParserInfo, p.ConcurrencyNum)
	var blocks = make([]*chain_common.Block, p.ConcurrencyNum)

	for i := uint64(0); i < p.ConcurrencyNum; i++ {
		bn := p.CurrentBlockNumber + i
		block, err := p.ChainEvm.GetBlockByNumber(bn)
		if err != nil {
			return fmt.Errorf("GetBlockByNumber err:%s [%d]", err.Error(), bn)
		}
		if block.Hash == "" || block.ParentHash == "" {
			log.Info("GetBlockByNumber:", bn, toolib.JsonString(&block))
			return fmt.Errorf("GetBlockByNumber data is nil: [%d]", bn)
		}
		hash := block.Hash
		parentHash := block.ParentHash
		blockList[i] = tables.TableBlockParserInfo{
			ParserType:  p.ParserType,
			BlockNumber: bn,
			BlockHash:   hash,
			ParentHash:  parentHash,
		}
		blocks[i] = block
	}
	//
	wg := sync.WaitGroup{}
	for i := uint64(0); i < p.ConcurrencyNum; i++ {
		wg.Add(1)
		go func(index uint64) {
			defer wg.Done()
			bn := p.CurrentBlockNumber + index
			if err := p.parsingBlockData(blocks[index]); err != nil {
				errList[index] = fmt.Errorf("ParsingBlockData err:%s [%d]", err.Error(), bn)
			}
		}(i)
	}
	wg.Wait()
	// err check
	for i := uint64(0); i < p.ConcurrencyNum; i++ {
		bn := p.CurrentBlockNumber + i
		if errList[i] != nil {
			return fmt.Errorf("ParsingBlockData err:%s [%d]", errList[i].Error(), bn)
		}
	}
	// block
	if err := p.DbDao.CreateBlockInfoList(blockList); err != nil {
		return fmt.Errorf("AddBlockInfoList err:%s", err.Error())
	} else {
		atomic.AddUint64(&p.CurrentBlockNumber, p.ConcurrencyNum)
	}
	if err := p.DbDao.DeleteBlockInfo(p.ParserType, p.CurrentBlockNumber-20); err != nil {
		log.Error("DeleteBlockInfo err:", p.ParserType.ToString(), err.Error())
	}
	return nil
}
