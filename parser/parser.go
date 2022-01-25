package parser

import (
	"context"
	"das-pay/chain/chain_ckb"
	"das-pay/chain/chain_evm"
	"das-pay/chain/chain_tron"
	"das-pay/config"
	"das-pay/dao"
	"das-pay/parser/parser_ckb"
	"das-pay/parser/parser_common"
	"das-pay/parser/parser_evm"
	"das-pay/parser/parser_tron"
	"das-pay/tables"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"strings"
	"sync"
)

type KitParser struct {
	ParserEth     *parser_evm.ParserEvm
	ParserBsc     *parser_evm.ParserEvm
	ParserPolygon *parser_evm.ParserEvm
	ParserCkb     *parser_ckb.ParserCkb
	ParserTron    *parser_tron.ParserTron

	Ctx    context.Context
	Cancel context.CancelFunc
	Wg     *sync.WaitGroup
	DbDao  *dao.DbDao
}

func NewKitParser(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, dbDao *dao.DbDao) (*KitParser, error) {
	kp := KitParser{
		Ctx:    ctx,
		Cancel: cancel,
		Wg:     wg,
		DbDao:  dbDao,
	}
	if err := kp.initParserEth(); err != nil {
		return nil, err
	}
	if err := kp.initParserBsc(); err != nil {
		return nil, err
	}
	if err := kp.initParserTron(); err != nil {
		return nil, err
	}
	if err := kp.initParserPolygon(); err != nil {
		return nil, err
	}
	if err := kp.initParserCkb(); err != nil {
		return nil, err
	}
	return &kp, nil
}

func (k *KitParser) initParserEth() error {
	if config.Cfg.Chain.Eth.Switch {
		chainEth, err := chain_evm.Initialize(k.Ctx, config.Cfg.Chain.Eth.Node, config.Cfg.Chain.Eth.RefundAddFee)
		if err != nil {
			return fmt.Errorf("chain_evm.Initialize eth err: %s", err.Error())
		}
		k.ParserEth = &parser_evm.ParserEvm{
			PayTokenId: tables.TokenIdEth,
			ChainEvm:   chainEth,
			ParserCommon: parser_common.ParserCommon{
				Ctx:                k.Ctx,
				Wg:                 k.Wg,
				DbDao:              k.DbDao,
				ParserType:         tables.ParserTypeETH,
				Address:            config.Cfg.Chain.Eth.Address,
				CurrentBlockNumber: config.Cfg.Chain.Eth.CurrentBlockNumber,
				ConcurrencyNum:     config.Cfg.Chain.Eth.ConcurrencyNum,
				ConfirmNum:         config.Cfg.Chain.Eth.ConfirmNum,
			},
		}
	}
	return nil
}

func (k *KitParser) initParserBsc() error {
	if config.Cfg.Chain.Bsc.Switch {
		chainBsc, err := chain_evm.Initialize(k.Ctx, config.Cfg.Chain.Bsc.Node, config.Cfg.Chain.Bsc.RefundAddFee)
		if err != nil {
			return fmt.Errorf("chain_evm.Initialize bsc err: %s", err.Error())
		}
		k.ParserBsc = &parser_evm.ParserEvm{
			PayTokenId: tables.TokenIdBnb,
			ChainEvm:   chainBsc,
			ParserCommon: parser_common.ParserCommon{
				Ctx:                k.Ctx,
				Wg:                 k.Wg,
				DbDao:              k.DbDao,
				ParserType:         tables.ParserTypeBSC,
				Address:            config.Cfg.Chain.Bsc.Address,
				CurrentBlockNumber: config.Cfg.Chain.Bsc.CurrentBlockNumber,
				ConcurrencyNum:     config.Cfg.Chain.Bsc.ConcurrencyNum,
				ConfirmNum:         config.Cfg.Chain.Bsc.ConfirmNum,
			},
		}
	}
	return nil
}

func (k *KitParser) initParserTron() error {
	if config.Cfg.Chain.Tron.Switch {
		chainTron, err := chain_tron.Initialize(k.Ctx, config.Cfg.Chain.Tron.Node)
		if err != nil {
			return fmt.Errorf("chain_ckb.Initialize tron err: %s", err.Error())
		}
		address := config.Cfg.Chain.Tron.Address
		if strings.HasPrefix(address, common.TronBase58PreFix) {
			if address, err = common.TronBase58ToHex(address); err != nil {
				return fmt.Errorf("TronBase58ToHex err: %s", err.Error())
			}
		}
		k.ParserTron = &parser_tron.ParserTron{
			ChainTron: chainTron,
			ParserCommon: parser_common.ParserCommon{
				Ctx:                k.Ctx,
				Wg:                 k.Wg,
				DbDao:              k.DbDao,
				ParserType:         tables.ParserTypeTRON,
				Address:            address,
				CurrentBlockNumber: config.Cfg.Chain.Tron.CurrentBlockNumber,
				ConcurrencyNum:     config.Cfg.Chain.Tron.ConcurrencyNum,
				ConfirmNum:         config.Cfg.Chain.Tron.ConfirmNum,
			},
		}
	}
	return nil
}

func (k *KitParser) initParserPolygon() error {
	if config.Cfg.Chain.Polygon.Switch {
		chainPolygon, err := chain_evm.Initialize(k.Ctx, config.Cfg.Chain.Polygon.Node, config.Cfg.Chain.Polygon.RefundAddFee)
		if err != nil {
			return fmt.Errorf("chain_evm.Initialize polygon err: %s", err.Error())
		}
		k.ParserPolygon = &parser_evm.ParserEvm{
			PayTokenId: tables.TokenIdMatic,
			ChainEvm:   chainPolygon,
			ParserCommon: parser_common.ParserCommon{
				Ctx:                k.Ctx,
				Wg:                 k.Wg,
				DbDao:              k.DbDao,
				ParserType:         tables.ParserTypePOLYGON,
				Address:            config.Cfg.Chain.Polygon.Address,
				CurrentBlockNumber: config.Cfg.Chain.Polygon.CurrentBlockNumber,
				ConcurrencyNum:     config.Cfg.Chain.Polygon.ConcurrencyNum,
				ConfirmNum:         config.Cfg.Chain.Polygon.ConfirmNum,
			},
		}
	}
	return nil
}

func (k *KitParser) initParserCkb() error {
	chainCkb, err := chain_ckb.Initialize(k.Ctx, config.Cfg.Chain.Ckb.CkbUrl, config.Cfg.Chain.Ckb.IndexUrl)
	if err != nil {
		return fmt.Errorf("chain_ckb.Initialize ckb err: %s", err.Error())
	}
	k.ParserCkb = &parser_ckb.ParserCkb{
		ChainCkb: chainCkb,
		ParserCommon: parser_common.ParserCommon{
			Ctx:                k.Ctx,
			Wg:                 k.Wg,
			DbDao:              k.DbDao,
			ParserType:         tables.ParserTypeCKB,
			Address:            config.Cfg.Chain.Ckb.Address,
			CurrentBlockNumber: config.Cfg.Chain.Ckb.CurrentBlockNumber,
			ConcurrencyNum:     config.Cfg.Chain.Ckb.ConcurrencyNum,
			ConfirmNum:         config.Cfg.Chain.Ckb.ConfirmNum,
		},
	}
	return nil
}

func (k *KitParser) Run() {
	if config.Cfg.Chain.Ckb.Switch {
		go k.ParserCkb.Parser()
	}
	if config.Cfg.Chain.Eth.Switch {
		go k.ParserEth.Parser()
	}
	if config.Cfg.Chain.Tron.Switch {
		go k.ParserTron.Parser()
	}
	if config.Cfg.Chain.Bsc.Switch {
		go k.ParserBsc.Parser()
	}
	if config.Cfg.Chain.Polygon.Switch {
		go k.ParserPolygon.Parser()
	}
}
