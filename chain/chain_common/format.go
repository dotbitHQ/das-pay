package chain_common

import (
	"math/big"
	"strconv"
)

func HexFormat(s string) string {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return s
}

func BigIntFromHex(h string) *big.Int {
	h = HexFormat(h)
	b, _ := new(big.Int).SetString(h, 16)
	return b
}

func HexToUint64(hex string) (uint64, error) {
	n, err := strconv.ParseUint(HexFormat(hex), 16, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}
