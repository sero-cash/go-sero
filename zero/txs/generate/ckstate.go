package generate

import (
	"fmt"

	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/zero/txs/assets"
	"github.com/sero-cash/go-sero/zero/utils"
)

type cyState struct {
	balance utils.I256
}

type cyStateMap map[keys.Uint256]*cyState

func (self cyStateMap) add(key *keys.Uint256, value *utils.U256) {
	if _, ok := self[*key]; ok {
		self[*key].balance.AddU(value)
	} else {
		self[*key] = &cyState{
			*value.ToI256().ToRef(),
		}
	}
}

func (self cyStateMap) sub(key *keys.Uint256, value *utils.U256) {
	if _, ok := self[*key]; ok {
		self[*key].balance.SubU(value)
	} else {
		self[*key] = &cyState{}
		self[*key].balance.SubU(value)
	}
}

func newcyStateMap() (ret cyStateMap) {
	ret = make(map[keys.Uint256]*cyState)
	return
}

type CKState struct {
	cy cyStateMap
	tk map[keys.Uint256]int
}

func NewCKState(fee *assets.Token) (ret CKState) {
	ret.cy = newcyStateMap()
	ret.cy.sub(&fee.Currency, &fee.Value)
	ret.tk = make(map[keys.Uint256]int)
	return
}

func (self *CKState) AddIn(asset *assets.Asset) (added bool, e error) {
	added = false
	if asset.Tkn != nil {
		self.cy.add(&asset.Tkn.Currency, &asset.Tkn.Value)
		added = true
	}
	if asset.Tkt != nil {
		if _, ok := self.tk[asset.Tkt.Value]; !ok {
			added = true
			self.tk[asset.Tkt.Value] = 1
			return
		} else {
			e = fmt.Errorf("in tkt duplicate: %v", asset.Tkt.Value)
			return
		}
	} else {
		return
	}
}

func (self *CKState) AddOut(asset *assets.Asset) (added bool, e error) {
	added = false
	if asset.Tkn != nil {
		self.cy.sub(&asset.Tkn.Currency, &asset.Tkn.Value)
		added = true
	}
	if asset.Tkt != nil {
		if c, ok := self.tk[asset.Tkt.Value]; ok {
			if c > 0 {
				added = true
				self.tk[asset.Tkt.Value] = c - 1
				return
			} else {
				e = fmt.Errorf("out tkt duplicate: %v", asset.Tkt.Value)
				return
			}
		} else {
			e = fmt.Errorf("out tkt not in ins: %v", asset.Tkt.Value)
			return
		}
	} else {
		return
	}
}

func (self *CKState) Check() (e error) {
	for currency, state := range self.cy {
		if state.balance.Cmp(&utils.I256_0) != 0 {
			e = fmt.Errorf("currency %v banlance != 0", currency)
			return
		}
	}

	for ticket, c := range self.tk {
		if c > 0 {
			e = fmt.Errorf("tikect not use %v", ticket)
			return
		}
	}
	return
}
