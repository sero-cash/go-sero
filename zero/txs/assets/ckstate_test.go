package assets

import (
	"testing"

	"github.com/sero-cash/go-czero-import/c_type"

	"github.com/sero-cash/go-sero/zero/utils"
)

var sero_token = Token{
	utils.CurrencyToUint256("SERO"),
	utils.NewU256(100),
}

var tk_ticket = Ticket{
	utils.CurrencyToUint256("TK"),
	c_type.RandUint256(),
}

var token_asset = Asset{&sero_token, nil}

var ticket_asset = Asset{nil, &tk_ticket}

var asset = Asset{&sero_token, &tk_ticket}

func TestCkState_OutPlus(t *testing.T) {
	ck := NewCKState(true, &sero_token)
	if ck.Check() == nil {
		t.Fail()
	}
	ck.AddIn(&token_asset)
	if ck.Check() != nil {
		t.Fail()
	}
	ck.AddIn(&ticket_asset)
	if ck.Check() == nil {
		t.Fail()
	}

	ck.AddOut(&asset)
	if ck.Check() == nil {
		t.Fail()
	}

	tkns, tkts := ck.GetList()
	if len(tkns) != 1 {
		t.Fail()
	}
	if len(tkts) != 0 {
		t.Fail()
	}
	ck.AddIn(&token_asset)
	if ck.Check() != nil {
		t.Fail()
	}
	tkns, tkts = ck.GetList()
	if len(tkns) != 0 {
		t.Fail()
	}
	if len(tkts) != 0 {
		t.Fail()
	}

}

func TestCkState_InPlus(t *testing.T) {
	ck := NewCKState(false, &sero_token)
	if ck.Check() == nil {
		t.Fail()
	}
	ck.AddIn(&token_asset)
	if ck.Check() != nil {
		t.Fail()
	}
	ck.AddIn(&ticket_asset)
	if ck.Check() == nil {
		t.Fail()
	}

	tkns, tkts := ck.GetList()
	if len(tkns) != 0 {
		t.Fail()
	}
	if len(tkts) != 1 {
		t.Fail()
	}

	ck.AddOut(&asset)
	if ck.Check() == nil {
		t.Fail()
	}

	ck.AddIn(&token_asset)
	if ck.Check() != nil {
		t.Fail()
	}
	tkns, tkts = ck.GetList()
	if len(tkns) != 0 {
		t.Fail()
	}
	if len(tkts) != 0 {
		t.Fail()
	}

}
