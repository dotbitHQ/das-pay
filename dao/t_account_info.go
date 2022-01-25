package dao

import "das-pay/tables"

func (d *DbDao) GetAccountCount() (count int64, err error) {
	err = d.parserDb.Model(tables.TableAccountInfo{}).Where("account!=''").Count(&count).Error
	return
}

func (d *DbDao) GetOwnerCount() (count int64, err error) {
	err = d.parserDb.Model(tables.TableAccountInfo{}).Where("account!=''").Group("owner_chain_type,`owner`").Count(&count).Error
	return
}

type RegisterStatusCount struct {
	RegisterStatus tables.RegisterStatus `json:"register_status" gorm:"column:register_status"`
	CountNum       int64                 `json:"count_num" gorm:"column:count_num"`
}

func (d *DbDao) GetRegisterStatusCount() (list []RegisterStatusCount, err error) {
	err = d.db.Model(tables.TableDasOrderInfo{}).Select("register_status,count(*) AS count_num").
		Where("order_type=? AND order_status=? AND register_status>? AND register_status<?",
			tables.OrderTypeSelf, tables.OrderStatusDefault,
			tables.RegisterStatusConfirmPayment, tables.RegisterStatusRegistered).
		Group("register_status").Find(&list).Error
	return
}
