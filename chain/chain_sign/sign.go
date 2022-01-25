package chain_sign

import (
	"context"
	"crypto/sha256"
	"das-pay/chain/chain_tron"
	"encoding/hex"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/golang/protobuf/proto"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
)

type reqParam struct {
	Errno  int         `json:"errno"`
	Errmsg interface{} `json:"errmsg"`
	Data   interface{} `json:"data"`
}

type RemoteSignClient struct {
	ctx    context.Context
	client rpc.Client
}

func NewRemoteSignClient(ctx context.Context, apiUrl string) (*RemoteSignClient, error) {
	client, err := rpc.Dial(apiUrl)
	if err != nil {
		return nil, err
	}

	return &RemoteSignClient{
		ctx:    ctx,
		client: client,
	}, nil
}

const (
	SignMethodEvm  string = "wallet_eTHSignMsg"
	SignMethodTron string = "wallet_tronSignMsg"
	SignMethodCkb  string = "wallet_cKBSignMsg"
)

func (r *RemoteSignClient) SignCkbMessage(ckbSignerAddress, message string) ([]byte, error) {
	if common.Has0xPrefix(message) {
		message = message[2:]
	}
	reply := reqParam{}
	param := struct {
		Address     string `json:"address"`
		CkbBuildRet string `json:"ckb_build_ret"`
		Tx          string `json:"tx"`
	}{
		Address:     ckbSignerAddress,
		CkbBuildRet: "",
		Tx:          message,
	}
	if err := r.client.CallContext(r.ctx, &reply, SignMethodCkb, param); err != nil {
		return nil, fmt.Errorf("remoteRpcClient.Call err: %s", err.Error())
	}
	if reply.Errno == 0 {
		signTxStr := reply.Data.(string)
		signTxBys, err := hex.DecodeString(signTxStr)
		if err != nil {
			return nil, fmt.Errorf("hex.DecodeString signed tx err: %s", err.Error())
		}
		return signTxBys, nil
	} else {
		return nil, fmt.Errorf("remoteRpcClient.Call err: %s", reply.Errmsg)
	}
}

func (r *RemoteSignClient) SignEvmTx(method, address string, tx *types.Transaction) (*types.Transaction, error) {
	reply := reqParam{}
	txRlpBys, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, fmt.Errorf("rlp.EncodeToBytes err: %s", err.Error())
	}
	param := struct {
		Address string `json:"address"`
		Tx      string `json:"tx"`
	}{
		Address: address,
		Tx:      hex.EncodeToString(txRlpBys),
	}
	if err := r.client.CallContext(r.ctx, &reply, method, param); err != nil {
		return nil, fmt.Errorf("client.CallContext err: %s", err.Error())
	}
	if reply.Errno == 0 {
		signTxStr := reply.Data.(string)
		signTxBys, err := hex.DecodeString(signTxStr)
		if err != nil {
			return nil, fmt.Errorf("hex.DecodeString signed tx err: %s", err.Error())
		}
		signTx := types.Transaction{}
		if err = rlp.DecodeBytes(signTxBys, &signTx); err != nil {
			return nil, fmt.Errorf("rlp.DecodeBytes signed tx err: %s", err.Error())
		}
		return &signTx, nil
	} else {
		return nil, fmt.Errorf("client.CallContext err: %s", reply.Errmsg)
	}
}

func (r *RemoteSignClient) SignTrxTx(address string, tx *api.TransactionExtention) (*api.TransactionExtention, error) {
	rawTx, err := chain_tron.TransactionToHexString(tx.Transaction)
	if err != nil {
		return nil, fmt.Errorf("TransactionToHexString err: %s", err.Error())
	}
	if err, signTx := r.signTrxTx(address, rawTx); err != nil {
		return nil, fmt.Errorf("signTrxTx err: %s", err.Error())
	} else {
		if coreTx, err := chain_tron.NewTransactionFromHexString(signTx); err != nil {
			return nil, fmt.Errorf("NewTransactionFromHexString err: %s", err.Error())
		} else {
			if raw, err := proto.Marshal(coreTx.GetRawData()); err != nil {
				return nil, fmt.Errorf("Marshal err: %s", err.Error())
			} else {
				txIdStr := fmt.Sprintf("%x", sha256.Sum256(raw))
				txId, _ := hex.DecodeString(txIdStr)
				retTx := api.TransactionExtention{
					Transaction:    coreTx,
					Txid:           txId,
					ConstantResult: tx.ConstantResult,
					Result:         tx.Result,
				}
				return &retTx, nil
			}
		}
	}
}

func (r *RemoteSignClient) signTrxTx(address, tronTxHexStr string) (error, string) {
	reply := reqParam{}
	type addressInfo struct {
		Address string `json:"address"`
		Index   int64  `json:"index"`
	}
	param := struct {
		Addrs []addressInfo `json:"addrs"`
		Tx    string        `json:"tx"`
	}{
		Addrs: []addressInfo{
			{
				Address: address,
			},
		},
		Tx: tronTxHexStr,
	}
	if err := r.client.CallContext(r.ctx, &reply, SignMethodTron, param); err != nil {
		return fmt.Errorf("client.CallContext err: %s", err.Error()), ""
	}
	if reply.Errno == 0 {
		return nil, reply.Data.(string)
	} else {
		return fmt.Errorf("client.CallContext err: %s", reply.Errmsg), ""
	}
}
