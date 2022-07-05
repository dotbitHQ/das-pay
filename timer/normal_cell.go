package timer

import (
	"das-pay/config"
	"das-pay/notify"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"time"
)

func (d *DasTimer) doNormalCell() error {
	if config.Cfg.Chain.Ckb.Address == "" {
		return nil
	}
	parseAddr, err := address.Parse(config.Cfg.Chain.Ckb.Address)
	if err != nil {
		return fmt.Errorf("address.Parse err: %s", err.Error())
	}
	liveCells, total, err := d.DasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          nil,
		LockScript:        parseAddr.Script,
		CapacityNeed:      0,
		CapacityForChange: 0,
		SearchOrder:       indexer.SearchOrderDesc,
	})
	if err != nil {
		return fmt.Errorf("GetBalanceCells err: %s", err.Error())
	}
	log.Info("doNormalCell:", len(liveCells), total)
	capacity := total / common.OneCkb
	msg := `- Count：%d
- Capacity: %d
- Time：%s`
	msg = fmt.Sprintf(msg, len(liveCells), capacity, time.Now().Format("2006-01-02 15:04:05"))

	if capacity < 1000000 {
		notify.SendLarkTextNotifyAtAll(config.Cfg.Notify.LarkDasInfoKey, "Live Cells", msg)
	} else {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkDasInfoKey, "Live Cells", msg)
	}
	return nil
}
