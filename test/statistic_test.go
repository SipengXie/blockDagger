package test

import (
	"blockDagger/helper"
	"testing"
)

func TestTxStat(t *testing.T) {
	helper.TransactionCounting(18999949, 51)
}
