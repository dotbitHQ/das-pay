package chain_common

import (
	"fmt"
	"github.com/shopspring/decimal"
)

func (c *ChainCommon) GetBalance(address string) (decimal.Decimal, error) {
	var balance string
	method := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBalance","params":["%s", "%s"],"id":1}`, address, "latest")

	if resp, err := c.Request(c.Node, method, &balance); err != nil {
		return decimal.Zero, err
	} else if resp.Error.Code != 0 {
		return decimal.Zero, fmt.Errorf("request err: %s [%d]", resp.Error.Message, resp.Error.Code)
	}

	bigBalance := BigIntFromHex(balance)
	decBalance, err := decimal.NewFromString(bigBalance.String())
	if err != nil {
		return decimal.Zero, err
	}
	return decBalance, nil
}

func (c *ChainCommon) GetBlockByNumber(blockNumber uint64) (*Block, error) {
	block := Block{}
	method := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%x", %v],"id":1}`, blockNumber, true)

	if resp, err := c.Request(c.Node, method, &block); err != nil {
		return nil, err
	} else if resp.Error.Code != 0 {
		return nil, fmt.Errorf("request err: %s [%d]", resp.Error.Message, resp.Error.Code)
	}
	return &block, nil
}

func (c *ChainCommon) BestBlockNumber() (uint64, error) {
	var number string
	method := `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`

	if resp, err := c.Request(c.Node, method, &number); err != nil {
		return 0, err
	} else if resp.Error.Code != 0 {
		return 0, fmt.Errorf("request err: %s [%d]", resp.Error.Message, resp.Error.Code)
	}

	height, err := HexToUint64(number)
	if err != nil {
		return 0, fmt.Errorf("HexToUint64 err:%s", err.Error())
	}
	return height, nil
}
