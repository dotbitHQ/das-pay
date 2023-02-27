package parser_bitcoin

import (
	"das-pay/parser/parser_common"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/dotbitHQ/das-lib/bitcoin"
	"github.com/scorpiotzh/mylog"
	"github.com/shopspring/decimal"
	"sync/atomic"
	"time"
)

var log = mylog.NewLogger("parser_bitcoin", mylog.LevelDebug)

type ParserBitcoin struct {
	parser_common.ParserCommon
	NodeRpc *bitcoin.BaseRequest
}

func (p *ParserBitcoin) initCurrentBlockNumber() error {
	if block, err := p.DbDao.FindBlockInfo(p.ParserType); err != nil {
		return fmt.Errorf("FindBlockInfo err: %s", err.Error())
	} else if block.Id > 0 {
		p.CurrentBlockNumber = block.BlockNumber
	} else {
		data, err := p.getBlockChainInfo()
		if err != nil {
			return fmt.Errorf("getBlockChainInfo err: %s", err.Error())
		}
		p.CurrentBlockNumber = data.Blocks
	}
	return nil
}

func (p *ParserBitcoin) getBlockChainInfo() (data bitcoin.BlockChainInfo, e error) {
	err := p.NodeRpc.Request(bitcoin.RpcMethodGetBlockChainInfo, nil, &data)
	if err != nil {
		e = fmt.Errorf("req RpcMethodGetBlockChainInfo err: %s", err.Error())
		return
	} else if data.Blocks == 0 {
		e = fmt.Errorf("blockChainInfo.Blocks is 0")
		return
	}
	return
}

func (p *ParserBitcoin) Parser() error {

	if err := p.initCurrentBlockNumber(); err != nil {
		return fmt.Errorf("initCurrentBlockNumber err: %s", err.Error())
	}
	atomic.AddUint64(&p.CurrentBlockNumber, 1)
	p.Wg.Add(1)
	for {
		select {
		default:
			data, err := p.getBlockChainInfo()
			if err != nil {
				log.Error("getBlockChainInfo err: %s", err.Error())
				time.Sleep(time.Second * 10)
			} else if p.ConcurrencyNum > 1 && p.CurrentBlockNumber < (data.Blocks-p.ConfirmNum-p.ConcurrencyNum) {
				// todo
				time.Sleep(time.Second * 1)
			} else if p.CurrentBlockNumber < (data.Blocks - p.ConfirmNum) {
				// todo
				time.Sleep(time.Second * 5)
			} else {
				log.Info("Parser:", p.ParserType.ToString(), p.CurrentBlockNumber, data.Blocks)
				time.Sleep(time.Second * 10)
			}
		case <-p.Ctx.Done():
			p.Wg.Done()
			return nil
		}
	}
}

func (p *ParserBitcoin) parserConcurrencyMode() error {
	return nil
}
func (p *ParserBitcoin) parserSubMode() error {
	return nil
}

func (p *ParserBitcoin) parsingBlockData(block *bitcoin.BlockInfo) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}
	for _, v := range block.Tx {
		// get tx info
		var data btcjson.TxRawResult
		err := p.NodeRpc.Request(bitcoin.RpcMethodGetRawTransaction, []interface{}{v, true}, data)
		if err != nil {
			return fmt.Errorf("req RpcMethodGetRawTransaction err: %s", err.Error())
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
		// check inputs & pay info & order id
		if isMyTx {
			if len(data.Vin) == 0 {
				return fmt.Errorf("tx vin is nil")
			}
			addrVin, err := bitcoin.VinScriptSigToAddress(data.Vin[0].ScriptSig, bitcoin.GetDogeMainNetParams())
			if err != nil {
				return fmt.Errorf("VinScriptSigToAddress err: %s", err.Error())
			}
			payInfo, err := p.DbDao.GetPayInfoByHash(v)
			if err != nil {
				return fmt.Errorf("GetPayInfoByHash err: %s", err.Error())
			} else if payInfo.Id == 0 {
				// todo notify
				continue
			} else if payInfo.Address != addrVin {
				// todo notify
				continue
			}
			orderInfo, err := p.DbDao.GetOrderByOrderId(payInfo.OrderId)
			if err != nil {
				return fmt.Errorf("GetOrderByOrderId err: %s", err.Error())
			} else if orderInfo.Id == 0 {
				// todo notify
				continue
			} else if orderInfo.Address != addrVin {
				// todo notify
				continue
			}
			decValue := decimal.NewFromFloat(value)
			if orderInfo.PayAmount.Cmp(decValue) != 0 {
				// todo notify
				continue
			}
			if err := p.DbDao.UpdatePayStatus(&payInfo); err != nil {
				return fmt.Errorf("UpdatePayStatus err: %s", err.Error())
			}
			break
		}
	}
	return nil
}
