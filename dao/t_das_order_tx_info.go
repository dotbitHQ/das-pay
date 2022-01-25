package dao

import "das-pay/tables"

func (d *DbDao) GetMaybeRejectedRegisterTxs(timestamp int64) (list []tables.TableDasOrderTxInfo, err error) {
	err = d.db.Where("timestamp<? AND status=?", timestamp, tables.OrderTxStatusDefault).Find(&list).Error
	return
}
