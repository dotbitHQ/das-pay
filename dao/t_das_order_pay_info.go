package dao

import (
	"das-pay/tables"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

type RefundOrderInfo struct {
	OrderId      string            `json:"order_id" gorm:"column:order_id"`
	Hash         string            `json:"hash" gorm:"column:hash"`
	ChainType    common.ChainType  `json:"chain_type" gorm:"column:chain_type"`
	Address      string            `json:"address" gorm:"column:address"`
	PayTokenId   tables.PayTokenId `json:"pay_token_id" gorm:"column:pay_token_id"`
	PayAmount    decimal.Decimal   `json:"pay_amount" gorm:"column:pay_amount"`
	RefundStatus tables.TxStatus   `json:"refund_status" gorm:"column:refund_status"`
}

func (d *DbDao) GetNeedRefundOrderList() (list []RefundOrderInfo, err error) {
	sql := `SELECT p.order_id,p.hash,p.chain_type,p.address,o.pay_token_id,o.pay_amount,p.refund_status 
FROM t_das_order_pay_info p JOIN t_das_order_info o 
ON p.refund_status=? AND o.order_type=? AND p.order_id=o.order_id `
	err = d.db.Raw(sql, tables.TxStatusSending, tables.OrderTypeSelf).Find(&list).Error
	return
}

func (d *DbDao) CreateOrderPays(list []tables.TableDasOrderPayInfo) error {
	return d.db.Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns([]string{
			"chain_type", "address", "status", "account_id", "refund_status", "refund_hash",
		}),
	}).Create(&list).Error
}

func (d *DbDao) UpdateRefundStatus(hashList []string, oldStatus, newStatus tables.TxStatus) error {
	return d.db.Model(tables.TableDasOrderPayInfo{}).
		Where("hash IN(?) AND refund_status=?", hashList, oldStatus).
		Updates(map[string]interface{}{
			"refund_status": newStatus,
		}).Error
}

func (d *DbDao) UpdateRefundHash(hashList []string, refundHash string) error {
	return d.db.Model(tables.TableDasOrderPayInfo{}).
		Where("hash IN(?)", hashList).Updates(map[string]interface{}{
		"refund_hash": refundHash,
	}).Error
}

func (d *DbDao) GetMaybeRejectedPayInfoList() (count int64, err error) {
	err = d.db.Model(tables.TableDasOrderPayInfo{}).
		Where("status=?", tables.OrderStatusDefault).Count(&count).Error
	return
}

func (d *DbDao) GetUnRefundTxCount() (count int64, err error) {
	err = d.db.Model(tables.TableDasOrderPayInfo{}).
		Where("refund_status=?", tables.TxStatusSending).Count(&count).Error
	return
}
