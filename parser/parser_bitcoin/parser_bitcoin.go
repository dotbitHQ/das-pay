package parser_bitcoin

import (
	"das-pay/config"
	"das-pay/notify"
	"das-pay/parser/parser_common"
	"das-pay/tables"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/txscript"
	"github.com/dotbitHQ/das-lib/bitcoin"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/scorpiotzh/mylog"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"sync"
	"sync/atomic"
	"time"
)

var log = mylog.NewLogger("parser_bitcoin", mylog.LevelDebug)

type ParserBitcoin struct {
	PayTokenId tables.PayTokenId
	parser_common.ParserCommon
	NodeRpc *bitcoin.BaseRequest
}

func (p *ParserBitcoin) initCurrentBlockNumber() error {
	if block, err := p.DbDao.FindBlockInfo(p.ParserType); err != nil {
		return fmt.Errorf("FindBlockInfo err: %s", err.Error())
	} else if block.Id > 0 {
		p.CurrentBlockNumber = block.BlockNumber
	} else {
		data, err := p.NodeRpc.GetBlockChainInfo()
		if err != nil {
			return fmt.Errorf("GetBlockChainInfo err: %s", err.Error())
		}
		p.CurrentBlockNumber = data.Blocks
	}
	return nil
}

func (p *ParserBitcoin) Parser() {
	log.Info("ParserBitcoin:")
	if err := p.initCurrentBlockNumber(); err != nil {
		log.Error("initCurrentBlockNumber err: ", err.Error())
		return
	}
	atomic.AddUint64(&p.CurrentBlockNumber, 1)
	p.Wg.Add(1)
	for {
		select {
		default:
			data, err := p.NodeRpc.GetBlockChainInfo()
			if err != nil {
				log.Error("GetBlockChainInfo err: %s", err.Error())
				time.Sleep(time.Second * 10)
			} else if p.ConcurrencyNum > 1 && p.CurrentBlockNumber < (data.Blocks-p.ConfirmNum-p.ConcurrencyNum) {
				nowTime := time.Now()
				if err := p.parserConcurrencyMode(); err != nil {
					log.Error("parserConcurrencyMode err:", p.ParserType.ToString(), err.Error(), p.CurrentBlockNumber)
				}
				log.Warn("parserConcurrencyMode time:", p.ParserType.ToString(), time.Since(nowTime).Seconds())
				time.Sleep(time.Second * 1)
			} else if p.CurrentBlockNumber < (data.Blocks - p.ConfirmNum) {
				nowTime := time.Now()
				if err := p.parserSubMode(); err != nil {
					log.Error("parserSubMode err:", p.ParserType.ToString(), err.Error(), p.CurrentBlockNumber)
				}
				log.Warn("parserSubMode time:", p.ParserType.ToString(), time.Since(nowTime).Seconds())
				time.Sleep(time.Second * 5)
			} else {
				log.Info("Parser:", p.ParserType.ToString(), p.CurrentBlockNumber, data.Blocks)
				time.Sleep(time.Second * 10)
			}
		case <-p.Ctx.Done():
			p.Wg.Done()
			return
		}
	}
}

func (p *ParserBitcoin) parserConcurrencyMode() error {
	log.Info("parserConcurrencyMode:", p.ParserType.ToString(), p.CurrentBlockNumber, p.ConcurrencyNum)

	var blockList = make([]tables.TableBlockParserInfo, p.ConcurrencyNum)
	var blockCh = make(chan bitcoin.BlockInfo, p.ConcurrencyNum)

	blockLock := &sync.Mutex{}
	blockGroup := &errgroup.Group{}
	blockChCount := uint64(0)

	for i := uint64(0); i < p.ConcurrencyNum; i++ {
		bn := p.CurrentBlockNumber + i
		index := i
		blockGroup.Go(func() error {
			blockHash, err := p.NodeRpc.GetBlockHash(bn)
			if err != nil {
				return fmt.Errorf("req GetBlockHash err: %s", err.Error())
			}

			block, err := p.NodeRpc.GetBlock(blockHash)
			if err != nil {
				return fmt.Errorf("req GetBlock err: %s", err.Error())
			}

			hash := block.Hash
			parentHash := block.PreviousBlockHash

			blockLock.Lock()
			blockList[index] = tables.TableBlockParserInfo{
				ParserType:  p.ParserType,
				BlockNumber: bn,
				BlockHash:   hash,
				ParentHash:  parentHash,
			}
			blockLock.Unlock()
			blockCh <- block
			if atomic.AddUint64(&blockChCount, 1) == p.ConcurrencyNum {
				close(blockCh)
			}
			return nil
		})
	}

	blockGroup.Go(func() error {
		for v := range blockCh {
			if err := p.parsingBlockData(&v); err != nil {
				return fmt.Errorf("parsingBlockData err: %s", err.Error())
			}
		}
		return nil
	})

	if err := blockGroup.Wait(); err != nil {
		return fmt.Errorf("errGroup.Wait() err: %s", err.Error())
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
func (p *ParserBitcoin) parserSubMode() error {
	log.Info("parserSubMode:", p.ParserType.ToString(), p.CurrentBlockNumber)

	hash, err := p.NodeRpc.GetBlockHash(p.CurrentBlockNumber)
	if err != nil {
		return fmt.Errorf("req GetBlockHash err: %s", err.Error())
	}

	block, err := p.NodeRpc.GetBlock(hash)
	if err != nil {
		return fmt.Errorf("req GetBlock err: %s", err.Error())
	}
	blockHash := block.Hash
	parentHash := block.PreviousBlockHash
	log.Info("parserSubMode:", p.ParserType.ToString(), blockHash, parentHash)
	if fork, err := p.CheckFork(parentHash); err != nil {
		return fmt.Errorf("CheckFork err: %s", err.Error())
	} else if fork {
		log.Warn("CheckFork is true:", p.ParserType.ToString(), p.CurrentBlockNumber, blockHash, parentHash)
		atomic.AddUint64(&p.CurrentBlockNumber, ^uint64(0))
	} else if err := p.parsingBlockData(&block); err != nil {
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
	return nil
}

func (p *ParserBitcoin) parsingBlockData(block *bitcoin.BlockInfo) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}
	log.Info("parsingBlockData:", p.ParserType.ToString(), block.Height, block.Hash, len(block.Tx))
	for _, v := range block.Tx {
		// get tx info
		data, err := p.NodeRpc.GetRawTransaction(v)
		if err != nil {
			return fmt.Errorf("req GetRawTransaction err: %s", err.Error())
		}
		// check address of outputs
		isMyTx, value := false, float64(0)
		for _, vOut := range data.Vout {
			for _, addr := range vOut.ScriptPubKey.Addresses {
				if addr == p.Address {
					isMyTx = true
					value = vOut.Value
					break
				}
			}
			if isMyTx {
				break
			}
		}
		decValue := decimal.NewFromFloat(value)
		// check inputs & pay info & order id
		if isMyTx {
			log.Info("parsingBlockData:", p.ParserType.ToString(), v)
			if len(data.Vin) == 0 {
				return fmt.Errorf("tx vin is nil")
			}
			_, addrPayload, err := bitcoin.VinScriptSigToAddress(data.Vin[0].ScriptSig, bitcoin.GetDogeMainNetParams())
			if err != nil {
				return fmt.Errorf("VinScriptSigToAddress err: %s", err.Error())
			}

			if ok, err := p.checkOpReturn(data, decValue, addrPayload); err != nil {
				return fmt.Errorf("checkOpReturn err: %s", err.Error())
			} else if ok {
				continue
			}
			if err = p.checkHashAndAmount(data, decValue, addrPayload); err != nil {
				return fmt.Errorf("checkHashAndAmount err: %s", err.Error())
			}
		}
	}
	return nil
}

func (p *ParserBitcoin) checkOpReturn(data btcjson.TxRawResult, decValue decimal.Decimal, addrPayload string) (bool, error) {
	var orderId string
	for _, vOut := range data.Vout {
		switch vOut.ScriptPubKey.Type {
		case txscript.NullDataTy.String():
			bys := common.Hex2Bytes(vOut.ScriptPubKey.Hex)
			if len(bys) <= 2 {
				continue
			}
			orderId = string(bys[2:])
			break
		}
	}
	log.Info("checkOpReturn:", orderId, addrPayload)
	if orderId == "" {
		return false, nil
	}
	order, err := p.DbDao.GetOrderByOrderId(orderId)
	if err != nil {
		return false, fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
	} else if order.Id == 0 {
		log.Warn("GetOrderByOrderId is not exist:", p.ParserType.ToString(), orderId)
		return false, nil
	}
	if order.PayTokenId != p.PayTokenId {
		log.Warn("order token id not match", order.OrderId)
		return false, nil
	}
	payAmount := order.PayAmount.DivRound(decimal.NewFromInt(1e8), 8)
	if payAmount.Cmp(decValue) != 0 {
		log.Warn("tx value not match order amount:", decValue.String(), payAmount.String())
		return false, nil
	}
	// change the status to confirm
	payInfo := tables.TableDasOrderPayInfo{
		Id:           0,
		Hash:         data.Txid,
		OrderId:      order.OrderId,
		ChainType:    p.ParserType.ToChainType(),
		Address:      addrPayload,
		Status:       tables.OrderTxStatusConfirm,
		AccountId:    order.AccountId,
		RefundStatus: tables.TxStatusDefault,
		RefundHash:   "",
		Timestamp:    time.Now().UnixNano() / 1e6,
	}
	if err := p.DbDao.UpdatePayStatus(&payInfo); err != nil {
		return false, fmt.Errorf("UpdatePayStatus err: %s", err.Error())
	}

	return true, nil
}

func (p *ParserBitcoin) checkHashAndAmount(data btcjson.TxRawResult, decValue decimal.Decimal, addrPayload string) error {
	payInfo, err := p.DbDao.GetPayInfoByHash(data.Txid)
	if err != nil {
		return fmt.Errorf("GetPayInfoByHash err: %s", err.Error())
	}
	returnHash, addressMatch, valueMatch := false, false, false
	var orderInfo tables.TableDasOrderInfo
	if payInfo.Id > 0 {
		returnHash = true
		orderInfo, err = p.DbDao.GetOrderByOrderId(payInfo.OrderId)
		if err != nil {
			return fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
		}
		if payInfo.Address == addrPayload && orderInfo.Address == addrPayload {
			addressMatch = true
		}
		payAmount := orderInfo.PayAmount.DivRound(decimal.NewFromInt(1e8), 8)
		if payAmount.Cmp(decValue) == 0 {
			valueMatch = true
		}
	}
	if orderInfo.Id == 0 {
		orderInfo, err = p.DbDao.GetOrderByAddrWithPayAmount(p.ParserType.ToChainType(), addrPayload, decValue)
		if err != nil {
			return fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
		}
		if orderInfo.Id > 0 {
			addressMatch = true
			valueMatch = true
		}
		payInfo = tables.TableDasOrderPayInfo{
			Id:           0,
			Hash:         data.Txid,
			OrderId:      orderInfo.OrderId,
			ChainType:    p.ParserType.ToChainType(),
			Address:      addrPayload,
			Status:       tables.OrderTxStatusConfirm,
			AccountId:    orderInfo.AccountId,
			RefundStatus: tables.TxStatusDefault,
			RefundHash:   "",
			Timestamp:    time.Now().UnixNano() / 1e6,
		}
	}
	if orderInfo.Id > 0 && orderInfo.PayTokenId != p.PayTokenId {
		log.Warn("order token id not match", orderInfo.OrderId)
		return nil
	}
	log.Info("checkHashAndAmount:", orderInfo.OrderId, data.Txid, returnHash, addressMatch, valueMatch)
	if addressMatch && valueMatch {
		payInfo.Status = tables.OrderTxStatusConfirm
		if err := p.DbDao.UpdatePayStatus(&payInfo); err != nil {
			return fmt.Errorf("UpdatePayStatus err: %s", err.Error())
		}
	} else {
		msg := `hash: %s
addrPayload: %s`
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "checkHashAndAmount", fmt.Sprintf(msg, data.Txid, addrPayload))
	}
	return nil
}
