package chain_tron

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"strings"
	"unicode"
)

type ChainTron struct {
	Ctx    context.Context
	Client api.WalletClient
}

func Initialize(ctx context.Context, node string) (*ChainTron, error) {
	conn, err := grpc.DialContext(ctx, node, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &ChainTron{
		Ctx:    ctx,
		Client: api.NewWalletClient(conn),
	}, nil
}

func GetMemo(s []byte) string {
	str := make([]rune, 0, len(s))
	for _, v := range string(s) {
		if unicode.IsControl(v) {
			continue
		}
		str = append(str, v)
	}
	return strings.Replace(string(str), " ", "", 1)
}

func TransactionToHexString(tx *core.Transaction) (string, error) {
	data, err := proto.Marshal(tx)
	if err != nil {
		return "", fmt.Errorf("marshal tx:%v", err)
	}
	return hex.EncodeToString(data), nil
}

func NewTransactionFromHexString(raw string) (*core.Transaction, error) {
	data, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("hex decode:%s %v", raw, err)
	}
	tx := core.Transaction{}
	if err := proto.Unmarshal(data, &tx); err != nil {
		return nil, fmt.Errorf("unmashal err:%v", err)
	}
	return &tx, nil
}
