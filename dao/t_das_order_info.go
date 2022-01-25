package dao

import (
	"das-pay/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (d *DbDao) GetOrderByOrderId(orderId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("order_id=? AND order_type=?", orderId, tables.OrderTypeSelf).Find(&order).Error
	return
}

func (d *DbDao) UpdatePayStatus(payInfo *tables.TableDasOrderPayInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := d.db.Model(tables.TableDasOrderInfo{}).
			Where("order_id=? AND order_type=? AND pay_status=?",
				payInfo.OrderId, tables.OrderTypeSelf, tables.TxStatusDefault).
			Updates(map[string]interface{}{
				"pay_status":      tables.TxStatusSending,
				"register_status": tables.RegisterStatusApplyRegister,
			}).Error; err != nil {
			return err
		}

		if err := d.db.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"chain_type", "address", "status", "account_id", "refund_status", "refund_hash",
			}),
		}).Create(&payInfo).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DbDao) GetNeedHedgeOrderList() (list []tables.TableDasOrderInfo, err error) {
	err = d.db.Where("order_type=? AND hedge_status=?", tables.OrderTypeSelf, tables.TxStatusSending).Find(&list).Error
	return
}

func (d *DbDao) UpdateHedgeStatus(orderId string, oldStatus, newStatus tables.TxStatus) error {
	return d.db.Model(tables.TableDasOrderInfo{}).
		Where("order_id=? AND order_type=? AND hedge_status=?",
			orderId, tables.OrderTypeSelf, oldStatus).
		Updates(map[string]interface{}{
			"hedge_status": newStatus,
		}).Error
}
