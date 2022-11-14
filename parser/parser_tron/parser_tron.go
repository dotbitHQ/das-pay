package parser_tron

import (
	"das-pay/chain/chain_tron"
	"das-pay/config"
	"das-pay/notify"
	"das-pay/parser/parser_common"
	"das-pay/tables"
	"encoding/hex"
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/protobuf/proto"
	"github.com/scorpiotzh/mylog"
	"github.com/shopspring/decimal"
	"sync"
	"sync/atomic"
	"time"
)

var log = mylog.NewLogger("parser_tron", mylog.LevelDebug)

type ParserTron struct {
	ChainTron *chain_tron.ChainTron
	parser_common.ParserCommon
}

func (p *ParserTron) Parser() {
	currentBlockNumber, err := p.ChainTron.GetBlockNumber()
	if err != nil {
		log.Error("GetTipBlockNumber err: ", p.ParserType.ToString(), err.Error())
		return
	}

	if err := p.InitCurrentBlockNumber(uint64(currentBlockNumber)); err != nil {
		log.Error("initCurrentBlockNumber err:", p.ParserType.ToString(), err.Error())
		return
	}
	atomic.AddUint64(&p.CurrentBlockNumber, 1)
	p.Wg.Add(1)
	for {
		select {
		default:
			latestBlockNumber, err := p.ChainTron.GetBlockNumber()
			if err != nil {
				log.Error("BestBlockNumber err:", p.ParserType.ToString(), err.Error())
			} else {
				if p.ConcurrencyNum > 1 && p.CurrentBlockNumber < (uint64(latestBlockNumber)-p.ConfirmNum-p.ConcurrencyNum) {
					nowTime := time.Now()
					if err = p.parserConcurrencyMode(); err != nil {
						log.Error("parserConcurrencyMode err:", p.ParserType.ToString(), err.Error(), p.CurrentBlockNumber)
						notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "Tron Parse", notify.GetLarkTextNotifyStr("parserConcurrencyMode", "", err.Error()))
					}
					log.Warn("parserConcurrencyMode time:", p.ParserType.ToString(), time.Since(nowTime).Seconds())
				} else if p.CurrentBlockNumber < (uint64(latestBlockNumber) - p.ConfirmNum) { // check rollback
					nowTime := time.Now()
					if err = p.parserSubMode(); err != nil {
						log.Error("parserSubMode err:", p.ParserType.ToString(), err.Error(), p.CurrentBlockNumber)
						notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "Tron Parse", notify.GetLarkTextNotifyStr("parserSubMode", "", err.Error()))
					}
					log.Warn("parserSubMode time:", p.ParserType.ToString(), time.Since(nowTime).Seconds())
				} else {
					log.Info("RunParser:", p.ParserType.ToString(), p.CurrentBlockNumber, latestBlockNumber)
					time.Sleep(time.Second * 10)
				}
				time.Sleep(time.Second)
			}
		case <-p.Ctx.Done():
			log.Warn("tron parse done")
			p.Wg.Done()
			return
		}
	}
}

func (p *ParserTron) parsingBlockData(block *api.BlockExtention) error {
	for _, tx := range block.Transactions {
		if len(tx.Transaction.RawData.Contract) != 1 {
			continue
		}
		orderId := chain_tron.GetMemo(tx.Transaction.RawData.Data)
		//log.Info("orderId:", orderId, tx.Transaction.RawData.Contract[0].Type)
		if orderId == "" {
			continue
		} else if len(orderId) > 64 {
			continue
		}

		switch tx.Transaction.RawData.Contract[0].Type {
		case core.Transaction_Contract_TransferContract:
			instance := core.TransferContract{}
			if err := proto.Unmarshal(tx.Transaction.RawData.Contract[0].Parameter.Value, &instance); err != nil {
				log.Error(" proto.Unmarshal err:", err.Error())
				continue
			}
			fromAddr, toAddr := hex.EncodeToString(instance.OwnerAddress), hex.EncodeToString(instance.ToAddress)
			//log.Info("orderId:", fromAddr, toAddr, p.Address)
			if toAddr != p.Address {
				continue
			}
			log.Info("parsingBlockData orderId:", orderId, hex.EncodeToString(tx.Txid))
			// check order id
			order, err := p.DbDao.GetOrderByOrderId(orderId)
			if err != nil {
				return fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
			} else if order.Id == 0 {
				log.Warn("GetOrderByOrderId is not exist:", p.ParserType.ToString(), orderId)
				continue
			}
			if order.PayTokenId != tables.TokenIdTrx {
				log.Warn("order token id not match", order.OrderId)
				continue
			}

			amountValue := decimal.New(instance.Amount, 0)
			if amountValue.Cmp(order.PayAmount) == -1 {
				log.Warn("tx value less than order amount:", amountValue.String(), order.PayAmount.String())
				continue
			}
			// change the status to confirm
			payInfo := tables.TableDasOrderPayInfo{
				Id:           0,
				Hash:         hex.EncodeToString(tx.Txid),
				OrderId:      order.OrderId,
				ChainType:    p.ParserType.ToChainType(),
				Address:      fromAddr,
				Status:       tables.OrderTxStatusConfirm,
				AccountId:    order.AccountId,
				RefundStatus: tables.TxStatusDefault,
				RefundHash:   "",
				Timestamp:    time.Now().UnixNano() / 1e6,
			}
			if err := p.DbDao.UpdatePayStatus(&payInfo); err != nil {
				return fmt.Errorf("UpdatePayStatus err: %s", err.Error())
			}
		case core.Transaction_Contract_TransferAssetContract:
		case core.Transaction_Contract_TriggerSmartContract:
		}
	}
	return nil
}

func (p *ParserTron) parserSubMode() error {
	log.Info("parserSubMode:", p.ParserType.ToString(), p.CurrentBlockNumber)
	block, err := p.ChainTron.GetBlockByNumber(p.CurrentBlockNumber)
	if err != nil {
		return fmt.Errorf("GetBlockByNumber err: %s", err.Error())
	} else {
		blockHash := hex.EncodeToString(block.Blockid)
		if block.BlockHeader == nil {
			return fmt.Errorf("parserSubMode: block.BlockHeader is nil")
		} else if block.BlockHeader.RawData == nil {
			return fmt.Errorf("parserSubMode: block.BlockHeader.RawData is nil")
		}
		parentHash := hex.EncodeToString(block.BlockHeader.RawData.ParentHash)
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

func (p *ParserTron) parserConcurrencyMode() error {
	log.Info("parserConcurrencyMode:", p.ParserType.ToString(), p.CurrentBlockNumber)
	var errList = make([]error, p.ConcurrencyNum)
	var blockList = make([]tables.TableBlockParserInfo, p.ConcurrencyNum)
	var blocks = make([]*api.BlockExtention, p.ConcurrencyNum)

	for i := uint64(0); i < p.ConcurrencyNum; i++ {
		bn := p.CurrentBlockNumber + i
		block, err := p.ChainTron.GetBlockByNumber(bn)
		if err != nil {
			return fmt.Errorf("GetBlockByNumber err:%s [%d]", err.Error(), bn)
		}
		if block.BlockHeader == nil {
			return fmt.Errorf("parserSubMode: block.BlockHeader is nil")
		} else if block.BlockHeader.RawData == nil {
			return fmt.Errorf("parserSubMode: block.BlockHeader.RawData is nil")
		}
		blockHash := hex.EncodeToString(block.Blockid)
		parentHash := hex.EncodeToString(block.BlockHeader.RawData.ParentHash)
		blockList[i] = tables.TableBlockParserInfo{
			ParserType:  p.ParserType,
			BlockNumber: bn,
			BlockHash:   blockHash,
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
