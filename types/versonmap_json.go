package types

import (
	"encoding/json"

	"github.com/ledgerwatch/erigon-lib/common"
)

type TempSturct struct {
	Map map[common.Address]map[string]interface{} `json:"output"`
}

func (t *TempSturct) ToString() string {
	b, _ := json.Marshal(t)
	return string(b)
}
