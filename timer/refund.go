package timer

import (
	"das-pay/chain/chain_evm"
	"das-pay/chain/chain_sign"
	"das-pay/config"
	"das-pay/dao"
	"das-pay/notify"
	"das-pay/tables"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"github.com/dotbitHQ/das-lib/bitcoin"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/dotbitHQ/das-lib/witness"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	addressCkb "github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/shopspring/decimal"
	"strings"
)

func (d *DasTimer) doOrderRefund() error {
	list, err := d.DbDao.GetNeedRefundOrderList()
	if err != nil {
		return fmt.Errorf("GetNeedRefundOrderList err: %s", err.Error())
	}
	if len(list) == 0 {
		return nil
	}
	log.Info("doOrderRefund:", len(list))

	nonceEth, nonceBsc, noncePolygon := uint64(0), uint64(0), uint64(0)
	if d.ChainEth != nil {
		nonce, err := d.ChainEth.NonceAt(config.Cfg.Chain.Eth.Address)
		if err != nil {
			return fmt.Errorf("NonceAt eth err: %s %s", err.Error(), config.Cfg.Chain.Eth.Address)
		}
		nonceEth = nonce
	}
	if d.ChainBsc != nil {
		nonce, err := d.ChainBsc.NonceAt(config.Cfg.Chain.Bsc.Address)
		if err != nil {
			return fmt.Errorf("NonceAt bsc err: %s %s", err.Error(), config.Cfg.Chain.Bsc.Address)
		}
		nonceBsc = nonce
	}
	if d.ChainPolygon != nil {
		nonce, err := d.ChainPolygon.NonceAt(config.Cfg.Chain.Polygon.Address)
		if err != nil {
			return fmt.Errorf("NonceAt polygon err: %s %s", err.Error(), config.Cfg.Chain.Polygon.Address)
		}
		noncePolygon = nonce
	}
	var ckbOrderList []*dao.RefundOrderInfo
	var dogeOrderList []*dao.RefundOrderInfo
	for i, v := range list {
		if v.RefundStatus != tables.TxStatusSending {
			continue
		}
		switch v.PayTokenId {
		case tables.TokenIdCkb, tables.TokenIdDas:
			ckbOrderList = append(ckbOrderList, &list[i])

		case tables.TokenIdEth:
			req := doOrderRefundEvmReq{
				order:      &list[i],
				nonce:      nonceEth,
				refund:     config.Cfg.Chain.Eth.Refund,
				chainEvm:   d.ChainEth,
				tokenId:    tables.TokenIdEth,
				chainType:  common.ChainTypeEth,
				from:       config.Cfg.Chain.Eth.Address,
				signMethod: chain_sign.SignMethodEvm,
				private:    config.Cfg.Chain.Eth.Private,
			}
			if hash, err := d.doOrderRefundEvm(&req); err != nil {
				log.Error("doOrderRefundEvm eth err:", err.Error(), v.OrderId)
				notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund eth", notify.GetLarkTextNotifyStr("doOrderRefundEvm", v.OrderId, err.Error()))
			} else if hash != "" {
				log.Info("doOrderRefundEvm eth ok:", v.OrderId, hash)
				nonceEth += 1
			}
		case tables.TokenIdBnb:
			req := doOrderRefundEvmReq{
				order:      &list[i],
				nonce:      nonceBsc,
				refund:     config.Cfg.Chain.Bsc.Refund,
				chainEvm:   d.ChainBsc,
				tokenId:    tables.TokenIdBnb,
				chainType:  common.ChainTypeEth,
				from:       config.Cfg.Chain.Bsc.Address,
				signMethod: chain_sign.SignMethodEvm,
				private:    config.Cfg.Chain.Bsc.Private,
			}
			if hash, err := d.doOrderRefundEvm(&req); err != nil {
				log.Error("doOrderRefundEvm bnb err:", err.Error(), v.OrderId, nonceBsc)
				notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund bnb", notify.GetLarkTextNotifyStr("doOrderRefundEvm", v.OrderId, fmt.Sprintf("%s[%d]", err.Error(), nonceBsc)))
				if strings.Contains(err.Error(), "nonce too low") {
					nonceBsc += 1
				}
			} else if hash != "" {
				log.Info("doOrderRefundEvm bnb ok:", v.OrderId, hash, nonceBsc)
				nonceBsc += 1
			}
		case tables.TokenIdMatic:
			req := doOrderRefundEvmReq{
				order:      &list[i],
				nonce:      noncePolygon,
				refund:     config.Cfg.Chain.Polygon.Refund,
				chainEvm:   d.ChainPolygon,
				tokenId:    tables.TokenIdMatic,
				chainType:  common.ChainTypeEth,
				from:       config.Cfg.Chain.Polygon.Address,
				signMethod: chain_sign.SignMethodEvm,
				private:    config.Cfg.Chain.Polygon.Private,
			}
			if hash, err := d.doOrderRefundEvm(&req); err != nil {
				log.Error("doOrderRefundEvm polygon err:", err.Error(), v.OrderId)
				notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund polygon", notify.GetLarkTextNotifyStr("doOrderRefundEvm", v.OrderId, err.Error()))
			} else if hash != "" {
				log.Info("doOrderRefundEvm polygon ok:", v.OrderId, hash)
				noncePolygon += 1
			}
		case tables.TokenIdTrx:
			if hash, err := d.doOrderRefundTrx(&list[i]); err != nil {
				log.Error("doOrderRefundTrx err:", err.Error(), v.OrderId)
				notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund trx", notify.GetLarkTextNotifyStr("doOrderRefundTrx", v.OrderId, err.Error()))
			} else if hash != "" {
				log.Info("doOrderRefundTrx ok:", v.OrderId, hash)
			}
		case tables.TokenIdDoge:
			dogeOrderList = append(dogeOrderList, &list[i])
		}
	}
	if len(ckbOrderList) > 0 {
		if hash, err := d.doOrderRefundCkb(ckbOrderList); err != nil {
			log.Error("doOrderRefundCkb ckb err:", err.Error(), ckbOrderList)
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund ckb", notify.GetLarkTextNotifyStr("doOrderRefundCkb", "", err.Error()))
		} else if hash != "" {
			log.Info("doOrderRefundCkb ckb ok:", ckbOrderList, hash)
		}
	}
	if len(dogeOrderList) > 0 {
		if hash, err := d.doOrderRefundDoge(dogeOrderList); err != nil {
			log.Error("doOrderRefundDoge err:", err.Error(), dogeOrderList)
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund doge", notify.GetLarkTextNotifyStr("doOrderRefundDoge", "", err.Error()))
		} else if hash != "" {
			log.Info("doOrderRefundDoge ok:", dogeOrderList, hash)
		}
	}
	return nil
}

type doOrderRefundEvmReq struct {
	order      *dao.RefundOrderInfo
	nonce      uint64
	refund     bool
	chainEvm   *chain_evm.ChainEvm
	tokenId    tables.PayTokenId
	chainType  common.ChainType
	from       string
	signMethod string
	private    string
}

func (d *DasTimer) doOrderRefundEvm(req *doOrderRefundEvmReq) (string, error) {
	if !req.refund || req.chainEvm == nil {
		return "", nil
	}
	if req.order == nil || req.order.PayTokenId != req.tokenId || req.order.ChainType != req.chainType || req.order.PayAmount.Cmp(decimal.Zero) != 1 {
		return "", nil
	}
	// tx and sign
	data := []byte(req.order.OrderId)
	refundAmount := req.order.PayAmount
	addFee := req.chainEvm.RefundAddFee
	gasPrice, gasLimit, err := req.chainEvm.EstimateGas(req.from, req.order.Address, req.order.PayAmount, data, addFee)
	if err != nil {
		return "", fmt.Errorf("EstimateGas err: %s", err.Error())
	}
	fee := gasPrice.Mul(gasLimit)
	if refundAmount.Cmp(fee) == -1 {
		log.Info("doOrderRefundEvm:", req.order.OrderId, req.tokenId, "fee > refund amount")
		return "", nil
	} else {
		refundAmount = refundAmount.Sub(fee)
		log.Info("doOrderRefundEvm:", req.order.OrderId, req.order.PayAmount, refundAmount, fee)
	}
	tx, err := req.chainEvm.NewTransaction(req.from, req.order.Address, refundAmount, data, req.nonce, addFee)
	if err != nil {
		return "", fmt.Errorf("NewTransaction err: %s", err.Error())
	}
	if req.private != "" {
		tx, err = req.chainEvm.SignWithPrivateKey(req.private, tx)
		if err != nil {
			return "", fmt.Errorf("SignWithPrivateKey err:%s", err.Error())
		}
	} else if d.SignClient != nil {
		tx, err = d.SignClient.SignEvmTx(req.signMethod, req.from, tx)
		if err != nil {
			return "", fmt.Errorf("SignEvmTx err: %s %s", err.Error(), req.tokenId)
		}
	} else {
		return "", fmt.Errorf("need sign tx")
	}
	// update
	hashList := []string{req.order.Hash}
	if err := d.DbDao.UpdateRefundStatus(hashList, tables.TxStatusSending, tables.TxStatusOk); err != nil {
		return "", fmt.Errorf("UpdateRefundStatus err: %s", err.Error())
	}
	if err := req.chainEvm.SendTransaction(tx); err != nil {
		if err := d.DbDao.UpdateRefundStatus(hashList, tables.TxStatusOk, tables.TxStatusSending); err != nil {
			log.Info("UpdateRefundStatus err: ", err.Error(), req.order.OrderId)
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund", notify.GetLarkTextNotifyStr("UpdateRefundStatus", req.order.OrderId, err.Error()))
		}
		return "", fmt.Errorf("SendTransaction err: %s", err.Error())
	}
	refundHash := tx.Hash().Hex()
	if err := d.DbDao.UpdateRefundHash(hashList, refundHash); err != nil {
		log.Info("UpdateRefundHash err:", err.Error(), req.order.Hash, refundHash)
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund "+req.tokenId.ToChainString(), notify.GetLarkTextNotifyStr("UpdateRefundHash", req.order.OrderId+","+refundHash, err.Error()))
	}
	return refundHash, nil
}

func (d *DasTimer) doOrderRefundTrx(order *dao.RefundOrderInfo) (string, error) {
	if !config.Cfg.Chain.Ckb.Refund || d.ChainTron == nil {
		return "", nil
	}
	if order == nil || order.PayTokenId != tables.TokenIdTrx || order.ChainType != common.ChainTypeTron || order.PayAmount.Cmp(decimal.Zero) != 1 {
		return "", nil
	}
	// tx and sign
	fromHex := config.Cfg.Chain.Tron.Address
	if strings.HasPrefix(fromHex, common.TronBase58PreFix) {
		fromData, _ := address.Base58ToAddress(fromHex)
		fromHex = hex.EncodeToString(fromData)
	}

	tx, err := d.ChainTron.CreateTransaction(fromHex, order.Address, order.OrderId, order.PayAmount.IntPart())
	if err != nil {
		return "", fmt.Errorf("CreateTransaction err: %s", err.Error())
	}
	if config.Cfg.Chain.Tron.Private != "" {
		tx, err = d.ChainTron.AddSign(tx.Transaction, config.Cfg.Chain.Tron.Private)
		if err != nil {
			return "", fmt.Errorf("AddSign err:%s", err.Error())
		}
	} else if d.SignClient != nil {
		tx, err = d.SignClient.SignTrxTx(fromHex, tx)
		if err != nil {
			return "", fmt.Errorf("SignTrxTx err: %s", err.Error())
		}
	} else {
		return "", fmt.Errorf("need sign tx")
	}
	// update
	hashList := []string{order.Hash}
	if err := d.DbDao.UpdateRefundStatus(hashList, tables.TxStatusSending, tables.TxStatusOk); err != nil {
		return "", fmt.Errorf("UpdateRefundStatus err: %s", err.Error())
	}
	if err := d.ChainTron.SendTransaction(tx.Transaction); err != nil {
		if err := d.DbDao.UpdateRefundStatus(hashList, tables.TxStatusOk, tables.TxStatusSending); err != nil {
			log.Info("UpdateRefundStatus err: ", err.Error(), order.OrderId)
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund", notify.GetLarkTextNotifyStr("UpdateRefundStatus", order.OrderId, err.Error()))
		}
		return "", fmt.Errorf("SendTransaction err: %s", err.Error())
	}
	// order tx
	refundHash := hex.EncodeToString(tx.Txid)
	if err := d.DbDao.UpdateRefundHash(hashList, refundHash); err != nil {
		log.Info("UpdateRefundHash err:", err.Error(), order.Hash, refundHash)
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund "+order.PayTokenId.ToChainString(), notify.GetLarkTextNotifyStr("UpdateRefundHash", order.OrderId+","+refundHash, err.Error()))
	}
	return refundHash, nil
}

func (d *DasTimer) doOrderRefundCkb(list []*dao.RefundOrderInfo) (string, error) {
	if !config.Cfg.Chain.Ckb.Refund || d.ChainCkb == nil {
		return "", nil
	}
	var txParams txbuilder.BuildTransactionParams
	totalAmount := decimal.Zero

	dasContract, err := core.GetDasContractInfo(common.DasContractNameDispatchCellType)
	if err != nil {
		return "", fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	balanceContract, err := core.GetDasContractInfo(common.DasContractNameBalanceCellType)
	if err != nil {
		return "", fmt.Errorf("GetDasContractInfo err: %s", err.Error())
	}
	var hashList []string
	for _, v := range list {
		hashList = append(hashList, v.Hash)
		totalAmount = totalAmount.Add(v.PayAmount)
		if v.PayTokenId != tables.TokenIdDas && v.PayTokenId != tables.TokenIdCkb {
			return "", nil
		}
		if v.PayAmount.Cmp(decimal.Zero) != 1 {
			return "", nil
		}
		parseAddr, err := addressCkb.Parse(v.Address)
		if err != nil {
			return "", fmt.Errorf("address.Parse err:%s %s", err.Error(), v.Address)
		}

		output := types.CellOutput{
			Capacity: v.PayAmount.BigInt().Uint64(),
			Lock:     parseAddr.Script,
			Type:     nil,
		}
		if dasContract.IsSameTypeId(parseAddr.Script.CodeHash) {
			ownerHex, _, err := d.DasCore.Daf().ArgsToHex(parseAddr.Script.Args)
			if err != nil {
				return "", fmt.Errorf("ArgsToHex err: %s", err.Error())
			}
			if ownerHex.DasAlgorithmId == common.DasAlgorithmIdEth712 {
				output.Type = balanceContract.ToScript(nil)
			}
		}
		txParams.Outputs = append(txParams.Outputs, &output)
		txParams.OutputsData = append(txParams.OutputsData, []byte(""))
	}
	// inputs
	fee := uint64(1e6)
	parseAddrFrom, err := addressCkb.Parse(config.Cfg.Chain.Ckb.Address)
	if err != nil {
		return "", fmt.Errorf("address.Parse err:%s %s", err.Error(), config.Cfg.Chain.Ckb.Address)
	}
	liveCells, total, err := d.DasCore.GetBalanceCells(&core.ParamGetBalanceCells{
		DasCache:          nil,
		LockScript:        parseAddrFrom.Script,
		CapacityNeed:      totalAmount.BigInt().Uint64() + fee,
		CapacityForChange: common.MinCellOccupiedCkb,
		SearchOrder:       indexer.SearchOrderDesc,
	})
	if err != nil {
		return "", fmt.Errorf("GetBalanceCells err: %s", err.Error())
	}
	for _, v := range liveCells {
		txParams.Inputs = append(txParams.Inputs, &types.CellInput{
			Since:          0,
			PreviousOutput: v.OutPoint,
		})
	}
	// change
	if change := total - totalAmount.BigInt().Uint64() - fee; change > 0 {
		txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
			Capacity: change,
			Lock:     parseAddrFrom.Script,
			Type:     nil,
		})
		txParams.OutputsData = append(txParams.OutputsData, []byte(""))
	}
	// witness
	actionWitness, err := witness.GenActionDataWitness("order_refund", nil)
	if err != nil {
		return "", fmt.Errorf("GenActionDataWitness err: %s", err.Error())
	}
	txParams.Witnesses = append(txParams.Witnesses, actionWitness)
	// tx
	txBuilder := txbuilder.NewDasTxBuilderFromBase(d.TxBuilderBase, nil)
	if err := txBuilder.BuildTransaction(&txParams); err != nil {
		return "", fmt.Errorf("BuildTransaction err: %s", err.Error())
	}
	//
	if err := d.DbDao.UpdateRefundStatus(hashList, tables.TxStatusSending, tables.TxStatusOk); err != nil {
		return "", fmt.Errorf("UpdateRefundStatus err: %s", err.Error())
	}
	if hash, err := txBuilder.SendTransactionWithCheck(false); err != nil {
		if err := d.DbDao.UpdateRefundStatus(hashList, tables.TxStatusOk, tables.TxStatusSending); err != nil {
			log.Info("UpdateRefundStatus err: ", err.Error(), hashList)
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund ckb", notify.GetLarkTextNotifyStr("UpdateRefundStatus", strings.Join(hashList, ","), err.Error()))
		}
		return "", fmt.Errorf("SendTransactionWithCheck err: %s", err.Error())
	} else {
		refundHash := hash.Hex()
		if err := d.DbDao.UpdateRefundHash(hashList, refundHash); err != nil {
			log.Info("UpdateRefundHash err:", err.Error(), hashList, refundHash)
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund ckb", notify.GetLarkTextNotifyStr("UpdateRefundHash", strings.Join(hashList, ",")+";"+refundHash, err.Error()))
		}
		return hash.Hex(), nil
	}
}

func (d *DasTimer) doOrderRefundDoge(list []*dao.RefundOrderInfo) (string, error) {
	var orderIds []string
	for _, v := range list {
		orderIds = append(orderIds, v.OrderId)
	}
	// check order
	orders, err := d.DbDao.GetOrders(orderIds)
	if err != nil {
		return "", fmt.Errorf("GetOrders err: %s", err.Error())
	}
	var notRefundMap = make(map[string]struct{})
	for _, v := range orders {
		if v.Action == common.DasActionApplyRegister && v.RegisterStatus == tables.RegisterStatusRegistered {
			notRefundMap[v.OrderId] = struct{}{}
		}
	}
	//
	var hashList []string
	var addresses []string
	var values []int64
	var total int64
	for _, v := range list {
		if _, ok := notRefundMap[v.OrderId]; ok {
			log.Warn("notRefundMap:", v.OrderId)
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "doOrderRefundDoge", "notRefundMap: "+v.OrderId)
			continue
		}
		hashList = append(hashList, v.Hash)
		addresses = append(addresses, v.Address)
		value := v.PayAmount.IntPart()
		total += value
		values = append(values, value)
	}
	if len(addresses) == 0 || len(values) == 0 {
		return "", nil
	}

	// get utxo
	_, uos, err := d.ChainDoge.GetUnspentOutputsDoge(config.Cfg.Chain.Doge.Address, config.Cfg.Chain.Doge.Private, total)
	if err != nil {
		return "", fmt.Errorf("GetUnspentOutputsDoge err: %s", err.Error())
	}

	// build tx
	tx, err := d.ChainDoge.NewTx(uos, addresses, values)
	if err != nil {
		return "", fmt.Errorf("NewTx err: %s", err.Error())
	}

	// sign
	var signTx *wire.MsgTx
	if config.Cfg.Chain.Doge.Private != "" {
		_, err = d.ChainDoge.LocalSignTx(tx, uos)
		if err != nil {
			return "", fmt.Errorf("LocalSignTx err: %s", err.Error())
		}
		signTx = tx
	} else {
		signTx, err = d.ChainDoge.RemoteSignTx(bitcoin.RemoteSignMethodDogeTx, tx, uos)
		if err != nil {
			return "", fmt.Errorf("LocalSignTx err: %s", err.Error())
		}
	}

	//
	if err := d.DbDao.UpdateRefundStatus(hashList, tables.TxStatusSending, tables.TxStatusOk); err != nil {
		return "", fmt.Errorf("UpdateRefundStatus err: %s", err.Error())
	}
	// send tx
	hash, err := d.ChainDoge.SendTx(signTx)
	if err != nil {
		if err := d.DbDao.UpdateRefundStatus(hashList, tables.TxStatusOk, tables.TxStatusSending); err != nil {
			log.Info("UpdateRefundStatus err: ", err.Error(), hashList)
			notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund doge", notify.GetLarkTextNotifyStr("UpdateRefundStatus", strings.Join(hashList, ","), err.Error()))
		}
		return "", fmt.Errorf("SendTx err: %s", err.Error())
	}
	if err := d.DbDao.UpdateRefundHash(hashList, hash); err != nil {
		log.Info("UpdateRefundHash err:", err.Error(), hashList, hash)
		notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "order refund doge", notify.GetLarkTextNotifyStr("UpdateRefundHash", strings.Join(hashList, ",")+";"+hash, err.Error()))
	}

	return hash, nil
}
