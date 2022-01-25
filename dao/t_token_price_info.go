package dao

import "das-pay/tables"

func (d *DbDao) GetTokenPriceList() (list []tables.TableTokenPriceInfo, err error) {
	err = d.parserDb.Order("id DESC").Find(&list).Error
	return
}
