package example

import (
	"context"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/sign"
	"github.com/DeAccountSystems/das-lib/txbuilder"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"sync"
	"testing"
)

func getClientTestnet2() (rpc.Client, error) {
	ckbUrl := "http://100.77.204.22:8224"
	indexerUrl := "http://100.77.204.22:8226"
	return rpc.DialWithIndexer(ckbUrl, indexerUrl)
}

func getNewDasCoreTestnet2() (*core.DasCore, error) {
	client, err := getClientTestnet2()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	contractArgs := "0xbc502a34a430e3e167c82a24db6f9237b15ebf35"
	contractCodeHash := "0x00000000000000000000000000000000000000000000000000545950455f4944"
	thqCodeHash := "0x96248cdefb09eed910018a847cfb51ad044c2d7db650112931760e3ef34a7e9a"
	ops := []core.DasCoreOption{
		core.WithClient(client),
		core.WithDasContractArgs(contractArgs),
		core.WithDasContractCodeHash(contractCodeHash),
		core.WithDasNetType(common.DasNetTypeTestnet2),
		core.WithTHQCodeHash(thqCodeHash),
	}
	dc := core.NewDasCore(context.Background(), &wg, ops...)
	// contract
	mapDasContractTypeArgs := map[common.DasContractName]string{
		//common.DasContractNameAccountCellType:       "0x6f0b8328b703617508d62d1f017b0d91fab2056de320a7b7faed4c777a356b7b",
		//common.DasContractNameBalanceCellType:       "0x27560fe2daa6150b771621300d1d4ea127832b7b326f2d70eed63f5333b4a8a9",
		common.DasContractNameConfigCellType: "0x34363fad2018db0b3b6919c26870f302da74c3c4ef4456e5665b82c4118eda51",
		//common.DasContractNameDispatchCellType:      "0xeedd10c7d8fee85c119daf2077fea9cf76b9a92ddca546f1f8e0031682e65aee",
		//common.DasContractNameAccountSaleCellType:   "0xed5d7fc00a3f8605bfe3f6717747bb0ed529fa064c2b8ce56e9677a0c46c2c1c",
		//common.DasContractNameAlwaysSuccess:         "0x7821c662b7efd50e7f6cf2b036efe53e07eccaf2e3447a2a470ee07ae455ab92",
		//common.DasContractNameIncomeCellType:        "0xd7b9d8213671aec03f3a3ab95171e0e79481db2c084586b9ea99914c00ff3716",
		//common.DasContractNamePreAccountCellType:    "0xd3f7ad59632a2ebdc2fe9d41aa69708ed1069b074cd8b297b205f835335d3a6b",
		//common.DASContractNameOfferCellType:         "0x443b2d1b3b00ffab1a2287d84c47b2c31a11aad24b183d732c213a69e3d6d390",
		//common.DasContractNameApplyRegisterCellType: "0xc78fa9066af1624e600ccfb21df9546f900b2afe5d7940d91aefc115653f90d9",
	}
	dc.InitDasContract(mapDasContractTypeArgs)
	// config cell
	if err = dc.InitDasConfigCell(); err != nil {
		return nil, err
	}
	//// so script
	//if err = dc.InitDasSoScript(); err != nil {
	//	return nil, err
	//}
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

	amount := uint64(33548859339) //* common.OneCkb
	orderid := "2047945383965376ad7cf6bcfa8e29e3"
	fromAddr := "ckt1qyqvsej8jggu4hmr45g4h8d9pfkpd0fayfksz44t9q"
	toAddr := "ckt1qyqvsej8jggu4hmr45g4h8d9pfkpd0fayfksz44t9q"
	fromParseAddress, err := address.Parse(fromAddr)
	if err != nil {
		t.Fatal(err)
	}
	txBuilderBase := getTxBuilderBase(dc, common.Bytes2Hex(fromParseAddress.Script.Args))
	toParseAddress, err := address.Parse(toAddr)
	if err != nil {
		t.Fatal(err)
	}
	liveCells, total, err := core.GetSatisfiedCapacityLiveCell(dc.Client(), nil, fromParseAddress.Script, nil, amount, common.MinCellOccupiedCkb)
	if err != nil {
		t.Fatal(err)
	}
	//
	var txParams txbuilder.BuildTransactionParams
	for _, v := range liveCells {
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
	fee := uint64(1e4)
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
