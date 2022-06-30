package tables

import "github.com/dotbitHQ/das-lib/common"

type TableAccountInfo struct {
	Id                  uint64                `json:"id" gorm:"column:id;primary_key;AUTO_INCREMENT"`
	BlockNumber         uint64                `json:"block_number" gorm:"column:block_number"`
	Outpoint            string                `json:"outpoint" gorm:"column:outpoint"`
	AccountId           string                `json:"account_id" gorm:"account_id"`
	Account             string                `json:"account" gorm:"column:account"`
	OwnerChainType      common.ChainType      `json:"owner_chain_type" gorm:"column:owner_chain_type"`
	Owner               string                `json:"owner" gorm:"column:owner"`
	OwnerAlgorithmId    common.DasAlgorithmId `json:"owner_algorithm_id" gorm:"column:owner_algorithm_id"`
	ManagerChainType    common.ChainType      `json:"manager_chain_type" gorm:"column:manager_chain_type"`
	Manager             string                `json:"manager" gorm:"column:manager"`
	ManagerAlgorithmId  common.DasAlgorithmId `json:"manager_algorithm_id" gorm:"column:manager_algorithm_id"`
	Status              AccountStatus         `json:"status" gorm:"column:status"`
	RegisteredAt        int64                 `json:"registered_at" gorm:"column:registered_at"`
	ExpiredAt           int64                 `json:"expired_at" gorm:"column:expired_at"`
	ConfirmProposalHash string                `json:"confirm_proposal_hash" gorm:"column:confirm_proposal_hash"`
}

type AccountStatus int

const (
	AccountStatusNotOpenRegister AccountStatus = -1
	AccountStatusNormal          AccountStatus = 0
	AccountStatusOnSale          AccountStatus = 1
	AccountStatusOnAuction       AccountStatus = 2
	TableNameAccountInfo                       = "t_account_info"
)

func (t *TableAccountInfo) TableName() string {
	return TableNameAccountInfo
}
