package example

import (
	"context"
	"das-pay/chain/chain_evm"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"strings"
	"testing"
)

const (
	EthNode = "https://data-seed-prebsc-1-s1.binance.org:8545"
	//EthNode = "https://matic-mumbai.chainstacklabs.com"
)

func TestTransactionEvm(t *testing.T) {
	ethClient, err := chain_evm.Initialize(context.Background(), EthNode, 0)
	if err != nil {
		t.Fatal(err)
	}
	private := "" //
	from := "0xc9f53b1d85356B60453F867610888D89a0B667Ad"
	to := "0xD43B906Be6FbfFFFF60977A0d75EC93696e01dC7"
	amount := decimal.NewFromInt(22379548536000000)
	nonce, err := ethClient.NonceAt(from)
	if err != nil {
		t.Fatal(err)
	}
	orderId := "62440abe309e51137b4429d52761e410"

	tx, err := ethClient.NewTransaction(from, to, amount, []byte(orderId), nonce, 0)
	if err != nil {
		t.Fatal(err)
	}
	tx, err = ethClient.SignWithPrivateKey(private, tx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(tx.Hash().Hex())
	if err := ethClient.SendTransaction(tx); err != nil {
		fmt.Println("err:", err)
		t.Fatal(err)
	}
}

func TestGetBalance(t *testing.T) {
	chainEth, err := chain_evm.Initialize(context.Background(), EthNode, 0)
	if err != nil {
		t.Fatal(err)
	}
	address := "0xD43B906Be6FbfFFFF60977A0d75EC93696e01dC7"
	bal, err := chainEth.GetBalance(address)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(bal)
}

func TestGetBlockByNumber(t *testing.T) {
	chainEth, err := chain_evm.Initialize(context.Background(), "https://rpc.ankr.com/polygon", 0)
	if err != nil {
		t.Fatal(err)
	}
	block, err := chainEth.GetBlockByNumber(31959896)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(block.Hash, block.ParentHash, len(block.Transactions))
}

func TestBestBlockNumber(t *testing.T) {
	chainEth, err := chain_evm.Initialize(context.Background(), EthNode, 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(chainEth.BestBlockNumber())

}

func TestBscTx(t *testing.T) {
	str := ``

	list := strings.Split(str, "\n")
	fmt.Println("len:", len(list))
	list = nil
	const url = "https://rpc.ankr.com/bsc" // url string
	rpcClient, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
	fmt.Println(rpcClient.PendingNonceAt(context.Background(), common.HexToAddress("")))

	//for _, v := range list {
	//	tx, isP, err := rpcClient.TransactionByHash(context.Background(), common.HexToHash(v))
	//	if err != nil {
	//		fmt.Println(err.Error())
	//	} else {
	//		fmt.Println(tx.Nonce(), isP, v)
	//	}
	//}

}
