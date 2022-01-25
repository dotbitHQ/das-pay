package chain_ckb

import (
	"context"
	"fmt"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/transaction"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/nervosnetwork/ckb-sdk-go/utils"
	"github.com/scorpiotzh/mylog"
)

var (
	log = mylog.NewLogger("chain_ckb", mylog.LevelDebug)
)

type ChainCkb struct {
	CkbUrl     string
	IndexerUrl string
	Client     rpc.Client
	Ctx        context.Context
}

func Initialize(ctx context.Context, ckbUrl, indexerUrl string) (*ChainCkb, error) {
	rpcClient, err := rpc.DialWithIndexer(ckbUrl, indexerUrl)
	if err != nil {
		return nil, fmt.Errorf("rpc.DialWithIndexer err:%s", err.Error())
	}
	return &ChainCkb{
		CkbUrl:     ckbUrl,
		IndexerUrl: indexerUrl,
		Client:     rpcClient,
		Ctx:        ctx,
	}, nil
}

func (c *ChainCkb) GetTipBlockNumber() (uint64, error) {
	if blockNumber, err := c.Client.GetTipBlockNumber(c.Ctx); err != nil {
		log.Error("GetTipBlockNumber err:", err.Error())
		return 0, err
	} else {
		return blockNumber, nil
	}
}

func (c *ChainCkb) GetBlockByNumber(blockNumber uint64) (*types.Block, error) {
	return c.Client.GetBlockByNumber(c.Ctx, blockNumber)
}

func (c *ChainCkb) GetTransaction(hash types.Hash) (*types.TransactionWithStatus, error) {
	return c.Client.GetTransaction(c.Ctx, hash)
}

func (c *ChainCkb) CreateTransaction(inputs []*types.CellInput, outputs []*types.CellOutput, outputsData [][]byte) (tx *types.Transaction, group []int, witnessArgs *types.WitnessArgs, err error) {
	systemScripts, err := utils.NewSystemScripts(c.Client)
	if err != nil {
		err = fmt.Errorf("NewSystemScripts err:%s", err.Error())
		return
	}
	tx = transaction.NewSecp256k1SingleSigTx(systemScripts)
	tx.Outputs = append(tx.Outputs, outputs...)
	tx.OutputsData = outputsData
	group, witnessArgs, err = transaction.AddInputsForTransaction(tx, inputs)
	if err != nil {
		err = fmt.Errorf("AddInputsForTransaction err:%s", err.Error())
		return
	}
	return
}
