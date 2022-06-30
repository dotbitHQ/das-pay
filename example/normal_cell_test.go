package example

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"testing"
)

func TestNormalCell(t *testing.T) {
	addr := "ckt1qyqvsej8jggu4hmr45g4h8d9pfkpd0fayfksz44t9q"
	parseAddr, err := address.Parse(addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(parseAddr.Script)
	dc, err := getNewDasCoreTestnet2()
	if err != nil {
		t.Fatal(err)
	}
	liveCells, total, err := core.GetSatisfiedCapacityLiveCell(dc.Client(), nil, parseAddr.Script, nil, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("total:", total)
	for _, v := range liveCells {
		if len(v.OutputData) > 0 {
			fmt.Println(v.OutPoint.TxHash, v.OutPoint.Index)
		}
	}

}
