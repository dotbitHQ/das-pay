package timer

import (
	"das-pay/config"
	"das-pay/notify"
	"fmt"
	"strings"
	"time"
)

func (d *DasTimer) doRejected() error {
	countPay, err := d.DbDao.GetMaybeRejectedPayInfoList()
	if err != nil {
		return fmt.Errorf("GetMaybeRejectedPayInfoList err: %s", err.Error())
	}
	countRefund, err := d.DbDao.GetUnRefundTxCount()
	if err != nil {
		return fmt.Errorf("GetUnRefundTxCount err: %s", err.Error())
	}
	list, err := d.DbDao.GetMaybeRejectedRegisterTxs(time.Now().Add(-time.Minute*10).UnixNano() / 1e6)
	if err != nil {
		return fmt.Errorf("GetMaybeRejectedRegisterTxs err: %s", err.Error())
	}
	if countPay == 0 && countRefund == 0 && len(list) == 0 {
		return nil
	}
	msg := `> unconfirmed payment: %d
> un refund txs: %d 
> rejected register tx: %d
%s`
	var orderList []string
	for _, v := range list {
		sinceMin := time.Since(time.Unix(v.Timestamp/1000, 0)).Minutes()
		orderList = append(orderList, fmt.Sprintf("%s : %s (%.2f min)", v.Action, v.OrderId, sinceMin))
	}
	msg = fmt.Sprintf(msg, countPay, countRefund, len(list), strings.Join(orderList, "\n"))
	if len(list) > 0 {
		notify.SendLarkTextNotifyAtAll(config.Cfg.Notify.LarkErrorKey, "Rejected Txs", msg)
	} else {
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "Rejected Txs", msg)
	}
	return nil
}
