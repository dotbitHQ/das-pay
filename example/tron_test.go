package example

import (
	"context"
	"das-pay/chain/chain_tron"
	"encoding/hex"
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/protobuf/proto"
	"testing"
)

const (
	TronNode = ""
)

func TestTronGetBlockByNumber(t *testing.T) {
	client, err := chain_tron.Initialize(context.Background(), TronNode)
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.GetBlockByNumber(22734291)
	if err != nil {
		t.Fatal(err)
	}
	for _, tx := range res.Transactions {
		if len(tx.Transaction.RawData.Contract) != 1 {
			continue
		}
		orderId := chain_tron.GetMemo(tx.Transaction.RawData.Data)
		if orderId == "" {
			continue
		} else if len(orderId) > 64 {
			continue
		}
		fmt.Println(orderId, len(orderId))
		switch tx.Transaction.RawData.Contract[0].Type {
		case core.Transaction_Contract_TransferContract:
			instance := core.TransferContract{}
			if err := proto.Unmarshal(tx.Transaction.RawData.Contract[0].Parameter.Value, &instance); err != nil {
				continue
			}
			fromAddr, toAddr := hex.EncodeToString(instance.OwnerAddress), hex.EncodeToString(instance.ToAddress)
			fmt.Println(fromAddr, toAddr)
		case core.Transaction_Contract_TransferAssetContract:
		case core.Transaction_Contract_TriggerSmartContract:
		}
	}
}

func TestTransactionTrx(t *testing.T) {
	client, err := chain_tron.Initialize(context.Background(), TronNode)
	if err != nil {
		t.Fatal(err)
	}

	fromHex, _ := address.Base58ToAddress("TQoLh9evwUmZKxpD1uhFttsZk3EBs8BksV")
	fromAddr := hex.EncodeToString(fromHex)

	toHex, _ := address.Base58ToAddress("TFUg8zKThCj23acDSwsVjQrBVRywMMQGP1")
	toAddr := hex.EncodeToString(toHex)
	fmt.Println(fromAddr, toAddr)

	orderId := "1ab8a3ca8bd77f17f970d78d0017b8bc"
	amount := int64(154815242)
	private := ""

	tx, err := client.CreateTransaction(fromAddr, toAddr, orderId, amount)
	if err != nil {
		t.Fatal(err)
	}

	txSign, err := client.AddSign(tx.Transaction, private)
	if err != nil {
		t.Fatal(err)
	}
	hash := hex.EncodeToString(txSign.Txid)
	fmt.Println("tx hash:", hash)
	err = client.SendTransaction(txSign.Transaction)
	if err != nil {
		t.Fatal(err)
	}
}
