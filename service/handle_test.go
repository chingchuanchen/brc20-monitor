package service

import (
	"github/HyprNetwork/brc20-balance-monitor/constant"
	"github/HyprNetwork/brc20-balance-monitor/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleDeploy(t *testing.T) {
	handle := NewHandler("0x01", "0x02", &model.InscriptionBRC20{
		Proto:        constant.BRC20_P,
		Operation:    constant.BRC20_OP_DEPLOY,
		BRC20Tick:    "test",
		BRC20Max:     "100000",
		BRC20Amount:  "100000",
		BRC20Decimal: "18",
		BRC20Limit:   "100000",
	})
	err := handle.HandleDepoly()
	assert.Nil(t, err)
}

func TestHandleMint(t *testing.T) {
	handle := NewHandler("0x01", "0x02", &model.InscriptionBRC20{
		Proto:        constant.BRC20_P,
		Operation:    constant.BRC20_OP_MINT,
		BRC20Tick:    "test",
		BRC20Decimal: "18",
		BRC20Amount:  "10",
	})
	err := handle.HandleMint(1)
	assert.Nil(t, err)
}

func TestHandleTransfer(t *testing.T) {
	handle := NewHandler("0x01", "0x02", &model.InscriptionBRC20{
		Proto:       constant.BRC20_P,
		Operation:   constant.BRC20_OP_TRANSFER,
		BRC20Tick:   "test",
		BRC20Max:    "100000",
		BRC20Amount: "10",
	})
	err := handle.HandleTransfer()
	assert.Nil(t, err)
}
