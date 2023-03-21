package example

import (
	"context"
	"fmt"
	"github.com/dotbitHQ/das-lib/bitcoin"
	"github.com/shopspring/decimal"
	"testing"
)

func getRpcClient() *bitcoin.BaseRequest {
	baseRep := bitcoin.BaseRequest{
		RpcUrl:   "",
		User:     "",
		Password: "",
		Proxy:    "socks5://127.0.0.1:8838",
	}
	return &baseRep
}

func TestDogeTx(t *testing.T) {
	rpcClient := getRpcClient()
	txTool := bitcoin.TxTool{
		RpcClient:        rpcClient,
		Ctx:              context.Background(),
		RemoteSignClient: nil,
		DustLimit:        bitcoin.DustLimitDoge,
		Params:           bitcoin.GetDogeMainNetParams(),
	}

	// get utxo
	addr := "DP86MSmWjEZw8GKotxcvAaW5D4e3qoEh6f"
	privateKey := ""
	payAmount := int64(102153551)
	_, uos, err := txTool.GetUnspentOutputsDoge(addr, privateKey, payAmount)
	if err != nil {
		t.Fatal(err)
	}

	// transfer
	tx, err := txTool.NewTx(uos, []string{addr}, []int64{payAmount}, "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = txTool.LocalSignTx(tx, uos)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := txTool.SendTx(tx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("hash:", hash)
}

func TestValue(t *testing.T) {
	//fmt.Println(decimal.NewFromInt(133333333).DivRound(decimal.NewFromInt(1e8), 8))
	//fmt.Println(decimal.NewFromFloat(float64(1.33333333) * 1e8).String())
	//fmt.Println(decimal.NewFromFloat(float64(1.33333333) * 1e8).Cmp(decimal.NewFromInt(133333333)))
	decValue := decimal.NewFromFloat(1.33333333)
	decValue = decValue.Mul(decimal.NewFromInt(1e8))
	fmt.Println(decValue.String())
}
