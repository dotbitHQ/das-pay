package chain_tron

import (
	"encoding/hex"
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
)

func (c *ChainTron) GetBlockNumber() (int64, error) {
	block, err := c.Client.GetNowBlock2(c.Ctx, new(api.EmptyMessage))
	if err != nil {
		return 0, err
	}
	return block.BlockHeader.RawData.Number, nil
}

func (c *ChainTron) GetBlockByNumber(blockNumber uint64) (*api.BlockExtention, error) {
	num := int64(blockNumber)
	block, err := c.Client.GetBlockByNum2(c.Ctx, &api.NumberMessage{Num: num})
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (c *ChainTron) CreateTransaction(fromHex, toHex, memo string, amount int64) (*api.TransactionExtention, error) {
	fromAddr, err := hex.DecodeString(fromHex)
	if err != nil {
		return nil, fmt.Errorf("decode from hex:%s %v", fromHex, err)
	}
	toAddr, err := hex.DecodeString(toHex)
	if err != nil {
		return nil, fmt.Errorf("decode to hex:%s %v", toHex, err)
	}
	in := &core.TransferContract{
		OwnerAddress: fromAddr,
		ToAddress:    toAddr,
		Amount:       amount,
	}
	tx, err := c.Client.CreateTransaction2(c.Ctx, in)
	if err != nil {
		return nil, fmt.Errorf("create tx err:%v", err)
	}
	if tx.Result.Code != api.Return_SUCCESS {
		return nil, fmt.Errorf("create tx failed:%s", tx.Result.Message)
	}
	if memo != "" {
		data, err := hex.DecodeString(fmt.Sprintf("%x", memo))
		if err != nil {
			return nil, fmt.Errorf("hex decode:%s %v", memo, err)
		}
		tx.Transaction.RawData.Data = data
	}
	return tx, nil
}

func (c *ChainTron) AddSign(tx *core.Transaction, private string) (*api.TransactionExtention, error) {
	pri, err := hex.DecodeString(private)
	if err != nil {
		return nil, fmt.Errorf("decode private:%v", err)
	}

	ts, err := c.Client.AddSign(c.Ctx, &core.TransactionSign{Transaction: tx, PrivateKey: pri})
	if err != nil {
		return nil, fmt.Errorf("sign err:%v", err)
	}
	if ts.Result.Code != api.Return_SUCCESS {
		return nil, fmt.Errorf("sign failed:%s", ts.Result.Message)
	}
	return ts, nil
}

func (c *ChainTron) SendTransaction(in *core.Transaction) error {
	ret, err := c.Client.BroadcastTransaction(c.Ctx, in)
	if err != nil {
		return fmt.Errorf("broadcast tx err:%v", err)
	}
	if ret.Code != api.Return_SUCCESS {
		return fmt.Errorf("broadcast tx failed:%s", ret.Message)
	}
	return nil
}
