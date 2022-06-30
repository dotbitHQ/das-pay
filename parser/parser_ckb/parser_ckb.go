package parser_ckb

import (
	"das-pay/chain/chain_ckb"
	"das-pay/config"
	"das-pay/notify"
	"das-pay/parser/parser_common"
	"das-pay/tables"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/mylog"
	"github.com/shopspring/decimal"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var log = mylog.NewLogger("parser_ckb", mylog.LevelDebug)

type ParserCkb struct {
	ChainCkb *chain_ckb.ChainCkb
	parser_common.ParserCommon

	addressArgs string
}

func (p *ParserCkb) Parser() {
	parseAdd, err := address.Parse(p.Address)
	if err != nil {
		log.Error("address.Parse err:", p.ParserType.ToString(), err.Error())
		return
	}
	p.addressArgs = common.Bytes2Hex(parseAdd.Script.Args)

	currentBlockNumber, err := p.ChainCkb.GetTipBlockNumber()
	if err != nil {
		log.Error("GetTipBlockNumber err: ", p.ParserType.ToString(), err.Error())
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
			latestBlockNumber, err := p.ChainCkb.GetTipBlockNumber()
			if err != nil {
				log.Error("BestBlockNumber err:", p.ParserType.ToString(), err.Error())
			} else {
				if p.ConcurrencyNum > 1 && p.CurrentBlockNumber < (latestBlockNumber-p.ConfirmNum-p.ConcurrencyNum) {
					nowTime := time.Now()
					if err = p.parserConcurrencyMode(); err != nil {
						log.Error("parserConcurrencyMode err:", p.ParserType.ToString(), err.Error(), p.CurrentBlockNumber)
						notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "Ckb Parse", notify.GetLarkTextNotifyStr("parserConcurrencyMode", "", err.Error()))
					}
					log.Warn("parserConcurrencyMode time:", p.ParserType.ToString(), time.Since(nowTime).Seconds())
				} else if p.CurrentBlockNumber < (latestBlockNumber - p.ConfirmNum) { // check rollback
					nowTime := time.Now()
					if err = p.parserSubMode(); err != nil {
						log.Error("parserSubMode err:", p.ParserType.ToString(), err.Error(), p.CurrentBlockNumber)
						notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "Ckb Parse", notify.GetLarkTextNotifyStr("parserSubMode", "", err.Error()))
					}
					log.Warn("parserSubMode time:", p.ParserType.ToString(), time.Since(nowTime).Seconds())
				} else {
					log.Info("RunParser:", p.ParserType.ToString(), p.CurrentBlockNumber, latestBlockNumber)
					time.Sleep(time.Second * 10)
				}
				time.Sleep(time.Millisecond * 300)
			}
		case <-p.Ctx.Done():
			log.Warn("ckb parse done")
			p.Wg.Done()
			return
		}
	}
}

func (p *ParserCkb) parsingBlockData(block *types.Block) error {
	for _, tx := range block.Transactions {
		for i, v := range tx.Outputs {
			if p.addressArgs != common.Bytes2Hex(v.Lock.Args) {
				continue
			}
			orderId := string(tx.OutputsData[i])
			if orderId == "" {
				continue
			}
			log.Info("parsingBlockData:", orderId, tx.Hash.Hex())
			capacity, _ := decimal.NewFromString(strconv.FormatUint(v.Capacity, 10))
			order, err := p.DbDao.GetOrderByOrderId(orderId)
			if err != nil {
				return fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
			} else if order.Id == 0 {
				log.Warn("GetOrderByOrderId is not exist:", p.ParserType.ToString(), orderId)
				continue
			}
			if order.PayTokenId != tables.TokenIdCkb && order.PayTokenId != tables.TokenIdDas {
				log.Warn("order token id not match", order.OrderId)
				continue
			}
			if capacity.Cmp(order.PayAmount) == -1 {
				log.Warn("tx value less than order amount:", capacity.String(), order.PayAmount.String())
				continue
			}
			txInputs, err := p.ChainCkb.GetTransaction(tx.Inputs[0].PreviousOutput.TxHash)
			if err != nil {
				return fmt.Errorf("GetTransaction err:%s", err.Error())
			}
			mode := address.Mainnet
			if config.Cfg.Server.Net != common.DasNetTypeMainNet {
				mode = address.Testnet
			}

			fromAddr, err := common.ConvertScriptToAddress(mode, txInputs.Transaction.Outputs[tx.Inputs[0].PreviousOutput.Index].Lock)
			if err != nil {
				return fmt.Errorf("common.ConvertScriptToAddress err:%s", err.Error())
			}
			// change the status to confirm
			payInfo := tables.TableDasOrderPayInfo{
				Id:           0,
				Hash:         tx.Hash.Hex(),
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
			break
		}
	}
	return nil
}

func (p *ParserCkb) parserSubMode() error {
	log.Info("parserSubMode:", p.ParserType.ToString(), p.CurrentBlockNumber)
	block, err := p.ChainCkb.GetBlockByNumber(p.CurrentBlockNumber)
	if err != nil {
		return fmt.Errorf("GetBlockByNumber err: %s", err.Error())
	} else {
		blockHash := block.Header.Hash.Hex()
		parentHash := block.Header.ParentHash.Hex()
		log.Info("parserSubMode:", p.ParserType.ToString(), blockHash, parentHash)
		if fork, err := p.CheckFork(parentHash); err != nil {
			return fmt.Errorf("CheckFork err: %s", err.Error())
		} else if fork {
			log.Warn("CheckFork is true:", p.ParserType.ToString(), p.CurrentBlockNumber, blockHash, parentHash)
			atomic.AddUint64(&p.CurrentBlockNumber, ^uint64(0))
		} else if err := p.parsingBlockData(block); err != nil {
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

func (p *ParserCkb) parserConcurrencyMode() error {
	log.Info("parserConcurrencyMode:", p.ParserType.ToString(), p.CurrentBlockNumber)
	var errList = make([]error, p.ConcurrencyNum)
	var blockList = make([]tables.TableBlockParserInfo, p.ConcurrencyNum)
	var blocks = make([]*types.Block, p.ConcurrencyNum)

	for i := uint64(0); i < p.ConcurrencyNum; i++ {
		bn := p.CurrentBlockNumber + i
		block, err := p.ChainCkb.GetBlockByNumber(bn)
		if err != nil {
			return fmt.Errorf("GetBlockByNumber err:%s [%d]", err.Error(), bn)
		}
		hash := block.Header.Hash.Hex()
		parentHash := block.Header.ParentHash.Hex()
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
