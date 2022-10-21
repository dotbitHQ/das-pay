package example

import (
	"context"
	"das-pay/chain/chain_evm"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"math/big"
	"strings"
	"sync"
	"testing"
	"time"
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
	address := ""
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

func TestPolygonOut(t *testing.T) {
	const url = "https://rpc.ankr.com/polygon" // url string
	rpcClient, err := ethclient.Dial(url)
	if err != nil {
		t.Fatal(err)
	}

	logs, err := rpcClient.FilterLogs(context.Background(), ethereum.FilterQuery{
		BlockHash: nil,
		FromBlock: big.NewInt(34497375),
		ToBlock:   big.NewInt(34497376),
		Addresses: []common.Address{common.HexToAddress("")},
		Topics:    nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(logs, err)
	for _, v := range logs {
		fmt.Println(v.Address, string(v.Data))
	}

}

func TestBscOut(t *testing.T) {
	const url = "https://rpc.ankr.com/polygon" // url string
	rpcClient, err := ethclient.Dial(url)
	if err != nil {
		t.Fatal(err)
	}
	chainId, err := rpcClient.ChainID(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	var orderIdMap = make(map[string]int)
	var lock sync.Mutex
	updateMap := func(orderId string) {
		lock.Lock()
		orderIdMap[orderId]++
		lock.Unlock()
	}
	blockNumStart := int64(34495948) //int64(22303059)
	blockNumEnd := int64(34584611)   //int64(22314584) //22283059
	//blockNumEnd := blockNumStart + 10000

	var wg sync.WaitGroup
	blockNumChan := make(chan int64, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			for {
				select {
				case blockNum, ok := <-blockNumChan:
					fmt.Println(blockNum)
					if ok {
						block, err := rpcClient.BlockByNumber(context.Background(), big.NewInt(blockNum))
						if err != nil {
							fmt.Println(err.Error())
						}
						if block == nil {
							fmt.Println("block is nil", blockNum)
						} else {
							for _, v := range block.Transactions() {
								asMessage, err := v.AsMessage(types.LatestSignerForChainID(chainId), v.GasPrice())
								if err != nil {
									t.Fatal(err)
								}
								fromHex := asMessage.From().Hex()
								if strings.EqualFold("0x497F300c628cd37Bc4f2E1A2864b11570E0f22A8", fromHex) {
									orderId := string(v.Data())
									updateMap(orderId)
								}
							}
						}
					} else {
						wg.Done()
						return
					}
				}
			}
		}()
	}
	//
	for i := blockNumStart; i < blockNumEnd; i++ {
		blockNumChan <- i
	}
	close(blockNumChan)

	log.Info("blockNumChan wait:", time.Now().String())
	wg.Wait()
	log.Info("blockNumChan ok:", time.Now().String())
	fmt.Println("len:", len(orderIdMap))
	for k, v := range orderIdMap {
		fmt.Println(k, ",", v)
	}
}

func TestBscNonce(t *testing.T) {
	const url = "https://rpc.ankr.com/bsc" // url string
	rpcClient, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}
	fmt.Println(rpcClient.PendingNonceAt(context.Background(), common.HexToAddress("")))

	str := "INSERT INTO t_das_order_pay_info(`hash`,order_id,chain_type,address,`status`,refund_status)" +
		"VALUES('0x1-%d','order_id_matic_20221021',1,'',1,1);"

	for i := 0; i < 100; i++ {
		fmt.Println(fmt.Sprintf(str, i))
	}

}

func TestBscTx(t *testing.T) {
	str := ``

	list := strings.Split(str, "\n")
	fmt.Println("len:", len(list))
	const url = "https://rpc.ankr.com/polygon" //"https://rpc.ankr.com/bsc" // url string
	rpcClient, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}

	var txListNotFund []string
	for _, v := range list {
		_, _, err := rpcClient.TransactionByHash(context.Background(), common.HexToHash(v))
		if err != nil {
			fmt.Println(err.Error(), v)
			txListNotFund = append(txListNotFund, v)
		} else {
			//fmt.Println(tx.Nonce(), v)
			//txListConfirm = append(txListConfirm, v)
		}
	}
	fmt.Println(len(txListNotFund))
}
