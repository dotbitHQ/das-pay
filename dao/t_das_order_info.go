package dao

import (
	"das-pay/tables"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (d *DbDao) GetOrderByOrderId(orderId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("order_id=? AND order_type=?", orderId, tables.OrderTypeSelf).Find(&order).Error
	return
}

func (d *DbDao) UpdatePayStatus(payInfo *tables.TableDasOrderPayInfo) error {
	var oldPayInfo tables.TableDasOrderPayInfo
	if err := d.db.Where("`hash`!=? AND order_id=? AND status=? AND refund_status=",
		payInfo.Hash, payInfo.OrderId, tables.OrderTxStatusConfirm, tables.TxStatusDefault).Find(&oldPayInfo).Error; err != nil {
		return fmt.Errorf("get old pay info err: %s", err.Error())
	} else if oldPayInfo.Id > 0 {
		payInfo.RefundStatus = tables.TxStatusSending
	}

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
