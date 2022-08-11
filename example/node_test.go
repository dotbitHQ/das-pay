package example

import (
	"context"
	"das-pay/chain/chain_evm"
	"das-pay/chain/chain_tron"
	"fmt"
	"testing"
)

func TestTronNode(t *testing.T) {
	node := "grpc.trongrid.io:50051"
	client, err := chain_tron.Initialize(context.Background(), node)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(client.GetBlockNumber())
}

func TestEthNode(t *testing.T) {
	node := "https://rpc.ankr.com/polygon"
	chainEth, err := chain_evm.Initialize(context.Background(), node, 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(chainEth.BestBlockNumber())
}

func TestCkb(t *testing.T) {
	c, err := getClientTestnet2()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(c.GetTipBlockNumber(context.Background()))
}
