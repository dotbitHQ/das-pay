package timer

import (
	"das-pay/config"
	"das-pay/notify"
	"das-pay/tables"
	"fmt"
	"time"
)

func (d *DasTimer) doDasInfo() error {
	if config.Cfg.Notify.LarkDasInfoKey == "" {
		return nil
	}
	accountNum, err := d.DbDao.GetAccountCount()
	if err != nil {
		return fmt.Errorf("GetAccountCount err: %s", err.Error())
	}
	ownerNum, err := d.DbDao.GetOwnerCount()
	if err != nil {
		return fmt.Errorf("GetOwnerCount err: %s", err.Error())
	}
	applyNum, preNum, proNum, confirmNum := int64(0), int64(0), int64(0), int64(0)
	list, err := d.DbDao.GetRegisterStatusCount()
	for _, v := range list {
		switch v.RegisterStatus {
		case tables.RegisterStatusApplyRegister:
			applyNum = v.CountNum
		case tables.RegisterStatusPreRegister:
			preNum = v.CountNum
		case tables.RegisterStatusProposal:
			preNum = v.CountNum
		case tables.RegisterStatusConfirmProposal:
			confirmNum = v.CountNum
		}
	}
	msg := `- accounts: %d
- owners: %d
- apply_register: %d
- pre_register: %d
- propose: %d
- confirm_proposal: %d
- time：%s`
	msg = fmt.Sprintf(msg, accountNum, ownerNum, applyNum, preNum, proNum, confirmNum, time.Now().Format("2006-01-02 15:04:05"))
	notify.SendLarkTextNotify(config.Cfg.Notify.LarkDasInfoKey, "register info", msg)
	return nil
}
