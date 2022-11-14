package example

import (
	"context"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/sign"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"sync"
	"testing"
)

func getClientTestnet2() (rpc.Client, error) {
	ckbUrl := "https://testnet.ckb.dev/"
	indexerUrl := "https://testnet.ckb.dev/indexer"
	return rpc.DialWithIndexer(ckbUrl, indexerUrl)
}

func getNewDasCoreTestnet2() (*core.DasCore, error) {
	client, err := getClientTestnet2()
	if err != nil {
		return nil, err
	}

	env := core.InitEnvOpt(common.DasNetTypeTestnet2,
		common.DasContractNameConfigCellType,
		//common.DasContractNameAccountCellType,
		//common.DasContractNameDispatchCellType,
		common.DasContractNameBalanceCellType,
		common.DasContractNameAlwaysSuccess,
		common.DasContractNameIncomeCellType,
		//common.DASContractNameSubAccountCellType,
		//common.DasContractNamePreAccountCellType,
		common.DASContractNameEip712LibCellType,
	)
	var wg sync.WaitGroup
	ops := []core.DasCoreOption{
		core.WithClient(client),
		core.WithDasContractArgs(env.ContractArgs),
		core.WithDasContractCodeHash(env.ContractCodeHash),
		core.WithDasNetType(common.DasNetTypeTestnet2),
		core.WithTHQCodeHash(env.THQCodeHash),
	}
	dc := core.NewDasCore(context.Background(), &wg, ops...)
	// contract
	dc.InitDasContract(env.MapContract)
	// config cell
	if err = dc.InitDasConfigCell(); err != nil {
		return nil, err
	}
	// so script
	if err = dc.InitDasSoScript(); err != nil {
		return nil, err
	}
	return dc, nil
}

func getTxBuilderBase(dasCore *core.DasCore, args string) *txbuilder.DasTxBuilderBase {
	private := ""
	handleSign := sign.LocalSign(private)
	txBuilderBase := txbuilder.NewDasTxBuilderBase(context.Background(), dasCore, handleSign, args)
	return txBuilderBase
}

func TestTransactionCkb(t *testing.T) {
	dc, err := getNewDasCoreTestnet2()
	if err != nil {
		t.Fatal(err)
	}

	amount := uint64(996916) * common.OneCkb
	fee := uint64(1e6)
	orderid := ""
	fromAddr := "ckt1qyqvsej8jggu4hmr45g4h8d9pfkpd0fayfksz44t9q"
	toAddr := "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qgre86nk8v9x44kq3flsemppzyd3xstveadq0yl2wcas56kkcz987r8vyyg3ky6pdn8457te6y8"
	fromParseAddress, err := address.Parse(fromAddr)
	if err != nil {
		t.Fatal(err)
	}
	txBuilderBase := getTxBuilderBase(dc, common.Bytes2Hex(fromParseAddress.Script.Args))
	toParseAddress, err := address.Parse(toAddr)
	if err != nil {
		t.Fatal(err)
	}
	liveCells, total, err := dc.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          nil,
		LockScript:        fromParseAddress.Script,
		CapacityNeed:      amount + fee,
		CapacityForChange: common.MinCellOccupiedCkb,
		SearchOrder:       indexer.SearchOrderAsc,
	})
	if err != nil {
		t.Fatal(err, total)
	}
	fmt.Println(len(liveCells))
	//
	var txParams txbuilder.BuildTransactionParams
	for i, v := range liveCells {
		fmt.Println(i)
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			Since:          0,
			PreviousOutput: v.OutPoint,
		})
	}
	txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
		Capacity: amount,
		Lock:     toParseAddress.Script,
		Type:     nil,
	})
	txParams.OutputsData = append(txParams.OutputsData, []byte(orderid))
	//

	if change := total - amount - fee; change > 0 {
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: change,
			Lock:     fromParseAddress.Script,
			Type:     nil,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte{})
	}

	//
	txBuilder := txbuilder.NewDasTxBuilderFromBase(txBuilderBase, nil)
	if err := txBuilder.BuildTransaction(&txParams); err != nil {
		t.Fatal(err)
	}
	if hash, err := txBuilder.SendTransactionWithCheck(false); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println("hash:", hash)
	}
}
