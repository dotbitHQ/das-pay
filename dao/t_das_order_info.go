package dao

import (
	"das-pay/tables"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (d *DbDao) GetOrderByOrderId(orderId string) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("order_id=? AND order_type=?",
		orderId, tables.OrderTypeSelf).Find(&order).Error
	return
}

func (d *DbDao) GetOrderByAddrWithPayAmount(chainType common.ChainType, addr string, payAmount decimal.Decimal) (order tables.TableDasOrderInfo, err error) {
	err = d.db.Where("chain_type=? AND address=? AND pay_amount=? AND order_type=?",
		chainType, addr, payAmount, tables.OrderTypeSelf).
		Order("id DESC").Limit(1).Find(&order).Error
	return
}

func (d *DbDao) UpdatePayStatus(payInfo *tables.TableDasOrderPayInfo) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(tables.TableDasOrderInfo{}).
			Where("order_id=? AND order_type=? AND pay_status=?",
				payInfo.OrderId, tables.OrderTypeSelf, tables.TxStatusDefault).
			Updates(map[string]interface{}{
				"pay_status":      tables.TxStatusSending,
				"register_status": tables.RegisterStatusApplyRegister,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(tables.TableDasOrderPayInfo{}).
			Where("order_id=? AND `hash`!=? AND status=? AND refund_status=?",
				payInfo.OrderId, payInfo.Hash, tables.OrderTxStatusConfirm, tables.TxStatusDefault).
			Updates(map[string]interface{}{
				"refund_status": tables.TxStatusSending,
			}).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.Insert{
			Modifier: "IGNORE",
		}).Create(&payInfo).Error; err != nil {
			return err
		}

		if err := tx.Model(tables.TableDasOrderPayInfo{}).
			Where("`hash`=? AND order_id=?", payInfo.Hash, payInfo.OrderId).
			Updates(map[string]interface{}{
				"chain_type": payInfo.ChainType,
				"address":    payInfo.Address,
				"status":     payInfo.Status,
				"account_id": payInfo.AccountId,
			}).Error; err != nil {
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

func (d *DbDao) GetOrders(orderIds []string) (list []tables.TableDasOrderInfo, err error) {
	if len(orderIds) == 0 {
		return
	}
	err = d.db.Select("order_id,action,pay_status,register_status,order_status").
		Where("order_id IN(?)", orderIds).Find(&list).Error
	return
}
