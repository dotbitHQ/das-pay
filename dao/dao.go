package dao

import (
	"das-pay/config"
	"fmt"
	"github.com/scorpiotzh/toolib"
	"gorm.io/gorm"
)

type DbDao struct {
	db       *gorm.DB
	parserDb *gorm.DB
}

func NewGormDB(dbMysql, parserMysql config.DbMysql) (*DbDao, error) {
	db, err := toolib.NewGormDB(dbMysql.Addr, dbMysql.User, dbMysql.Password, dbMysql.DbName, dbMysql.MaxOpenConn, dbMysql.MaxIdleConn)
	if err != nil {
		return nil, fmt.Errorf("toolib.NewGormDB db err: %s", err.Error())
	}
	parserDb, err := toolib.NewGormDB(parserMysql.Addr, parserMysql.User, parserMysql.Password, parserMysql.DbName, parserMysql.MaxOpenConn, parserMysql.MaxIdleConn)
	if err != nil {
		return nil, fmt.Errorf("toolib.NewGormDB parser db err: %s", err.Error())
	}
	return &DbDao{db: db, parserDb: parserDb}, nil
}
