package chain_evm

import (
	"context"
	"das-pay/chain/chain_common"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/scorpiotzh/mylog"
	"github.com/shopspring/decimal"
)

var (
	log = mylog.NewLogger("chain_evm", mylog.LevelDebug)
)

type ChainEvm struct {
	chain_common.ChainCommon
	Client *ethclient.Client
	Ctx    context.Context
}

func Initialize(ctx context.Context, node string, refundAddFee float64) (*ChainEvm, error) {
	ethClient, err := ethclient.Dial(node)
	if err != nil {
		return nil, fmt.Errorf("ethclient.Dial err: %s", err.Error())
	}
	return &ChainEvm{
		ChainCommon: chain_common.ChainCommon{Node: node, RefundAddFee: refundAddFee},
		Client:      ethClient,
		Ctx:         ctx,
	}, nil
}

func (c *ChainEvm) EstimateGas(from, to string, value decimal.Decimal, input []byte, addFee float64) (gasPrice, gasLimit decimal.Decimal, err error) {
	fromAddr := common.HexToAddress(from)
	toAddr := common.HexToAddress(to)
	call := ethereum.CallMsg{From: fromAddr, To: &toAddr, Value: value.BigInt(), Data: input}
	limit, err := c.Client.EstimateGas(c.Ctx, call)
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("EstimateGas err: %s", err.Error())
	}
	gasLimit, _ = decimal.NewFromString(fmt.Sprintf("%d", limit))
	fee, err := c.Client.SuggestGasPrice(c.Ctx)
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("SuggestGasPrice err: %s", err.Error())
	}
	gasPrice, _ = decimal.NewFromString(fmt.Sprintf("%d", fee))

	log.Info("EstimateGas:", from, to, value, gasPrice, gasLimit, addFee)
	if addFee > 1 && addFee < 2 {
		gasPrice = gasPrice.Mul(decimal.NewFromFloat(addFee))
	}
	return
}

func (c *ChainEvm) NewTransaction(from, to string, value decimal.Decimal, data []byte, nonce uint64, addFee float64) (*types.Transaction, error) {
	toAddr := common.HexToAddress(to)
	gasPrice, gasLimit, err := c.EstimateGas(from, to, value, data, addFee)
	if err != nil {
		return nil, err
	}
	log.Info("NewTransaction:", from, to, value, nonce, gasPrice, gasLimit, addFee)
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddr,
		Value:    value.BigInt(),
		Gas:      gasLimit.BigInt().Uint64(),
		GasPrice: gasPrice.BigInt(),
		Data:     data,
	})
	return tx, nil
}

func (c *ChainEvm) NonceAt(address string) (uint64, error) {
	return c.Client.NonceAt(c.Ctx, common.HexToAddress(address), nil)
}

func (c *ChainEvm) SignWithPrivateKey(private string, tx *types.Transaction) (*types.Transaction, error) {
	privateKey, err := crypto.HexToECDSA(chain_common.HexFormat(private))
	if err != nil {
		return nil, fmt.Errorf("crypto.HexToECDSA err: %s", err.Error())
	}

	chainID, err := c.Client.NetworkID(c.Ctx)
	if err != nil {
		return nil, fmt.Errorf("NetworkID err: %s", err.Error())
	}
	sigTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, fmt.Errorf("SignTx err: %s", err.Error())
	}
	return sigTx, nil
}

func (c *ChainEvm) SendTransaction(tx *types.Transaction) error {
	return c.Client.SendTransaction(c.Ctx, tx)
}
