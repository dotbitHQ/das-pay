package timer

import (
	"das-pay/tables"
	"fmt"
	"sync"
)

var (
	tokenLock sync.RWMutex
	mapToken  map[tables.PayTokenId]tables.TableTokenPriceInfo
)

func (d *DasTimer) doUpdateTokenMap() error {
	tokenLock.Lock()
	defer tokenLock.Unlock()

	list, err := d.DbDao.GetTokenPriceList()
	if err != nil {
		return fmt.Errorf("GetTokenPriceList err:%s", err.Error())
	}
	mapToken = make(map[tables.PayTokenId]tables.TableTokenPriceInfo)
	for i, v := range list {
		mapToken[v.TokenId] = list[i]
	}
	return nil
}

func GetTokenInfo(tokenId tables.PayTokenId) tables.TableTokenPriceInfo {
	if tokenId == tables.TokenIdDas {
		tokenId = tables.TokenIdCkb
	}
	tokenLock.RLock()
	defer tokenLock.RUnlock()
	t, _ := mapToken[tokenId]
	return t
}
