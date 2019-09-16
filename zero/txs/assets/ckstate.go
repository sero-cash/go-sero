package assets

import (
	"encoding/hex"
	"fmt"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/zero/utils"
)

type cyState struct {
	balance utils.I256
}

type cyStateMap map[c_type.Uint256]*cyState

func (self cyStateMap) add(key *c_type.Uint256, value *utils.U256) {
	if _, ok := self[*key]; ok {
		self[*key].balance.AddU(value)
	} else {
		self[*key] = &cyState{
			*value.ToI256().ToRef(),
		}
	}
}

func (self cyStateMap) sub(key *c_type.Uint256, value *utils.U256) {
	if _, ok := self[*key]; ok {
		self[*key].balance.SubU(value)
	} else {
		self[*key] = &cyState{}
		self[*key].balance.SubU(value)
	}
}

func newcyStateMap() (ret cyStateMap) {
	ret = make(cyStateMap)
	return
}

type cgState struct {
	count int
}

type cgStateMap map[c_type.Uint512]*cgState

func newcgStateMap() (ret cgStateMap) {
	ret = make(cgStateMap)
	return
}

func (self cgStateMap) add(category *c_type.Uint256, value *c_type.Uint256) {
	ticket := c_type.Uint512{}
	copy(ticket[:], category[:])
	copy(ticket[32:], value[:])
	if _, ok := self[ticket]; ok {
		state := self[ticket]
		state.count++
	} else {
		self[ticket] = &cgState{1}
	}
}

func (self cgStateMap) sub(category *c_type.Uint256, value *c_type.Uint256) {
	ticket := c_type.Uint512{}
	copy(ticket[:], category[:])
	copy(ticket[32:], value[:])
	if _, ok := self[ticket]; ok {
		state := self[ticket]
		state.count--
	} else {
		self[ticket] = &cgState{-1}
	}
}

type CKState struct {
	outPlus bool
	cy      cyStateMap
	tk      cgStateMap
}

func NewCKState(outPlus bool, fee *Token) (ret CKState) {
	ret.outPlus = outPlus
	ret.cy = newcyStateMap()
	ret.tk = newcgStateMap()
	if outPlus {
		ret.cy.add(&fee.Currency, &fee.Value)
	} else {
		ret.cy.sub(&fee.Currency, &fee.Value)
	}
	return
}

func (self *CKState) Tkns() (ret []Token) {
	for c, v := range self.cy {
		if v.balance.Cmp(&utils.I256_0) > 0 {
			ret = append(ret, Token{c, utils.U256(v.balance)})
		}
	}
	return
}

func (self *CKState) Tkts() (ret []Ticket) {
	for c, v := range self.tk {
		if v.count > 0 {
			category := c_type.Uint256{}
			value := c_type.Uint256{}
			copy(category[:], c[:32])
			copy(value[:], c[32:])
			ret = append(ret, Ticket{category, value})
		}
	}
	return
}

func (self *CKState) GetList() (tkns []Token, tkts []Ticket) {
	tkns = self.Tkns()
	tkts = self.Tkts()
	return
}

func (self *CKState) AddIn(asset *Asset) (added bool) {
	added = false
	if asset.Tkn != nil {
		if asset.Tkn.Currency != c_type.Empty_Uint256 {
			if asset.Tkn.Value.ToUint256() != c_type.Empty_Uint256 {
				if self.outPlus {
					self.cy.sub(&asset.Tkn.Currency, &asset.Tkn.Value)
				} else {
					self.cy.add(&asset.Tkn.Currency, &asset.Tkn.Value)
				}
				added = true
			}
		}
	}
	if asset.Tkt != nil {
		if asset.Tkt.Category != c_type.Empty_Uint256 {
			if asset.Tkt.Value != c_type.Empty_Uint256 {
				if self.outPlus {
					self.tk.sub(&asset.Tkt.Category, &asset.Tkt.Value)
				} else {
					self.tk.add(&asset.Tkt.Category, &asset.Tkt.Value)
				}
				added = true
			}
		}
	}
	return
}

func (self *CKState) AddOut(asset *Asset) (added bool) {
	added = false
	if asset.Tkn != nil {
		if self.outPlus {
			self.cy.add(&asset.Tkn.Currency, &asset.Tkn.Value)
		} else {
			self.cy.sub(&asset.Tkn.Currency, &asset.Tkn.Value)
		}
		added = true
	}
	if asset.Tkt != nil {
		if self.outPlus {
			self.tk.add(&asset.Tkt.Category, &asset.Tkt.Value)
		} else {
			self.tk.sub(&asset.Tkt.Category, &asset.Tkt.Value)
		}
		added = true
	}
	return
}

func (self *CKState) CheckToken() (e error) {
	for currency, state := range self.cy {
		if state.balance.Cmp(&utils.I256_0) != 0 {
			e = fmt.Errorf("currency %v value %v not balance", utils.BytesToCurrency(currency[:]), state.balance.ToIntRef())
			return
		}
	}
	return
}

func (self *CKState) CheckTicket() (e error) {
	for c, v := range self.tk {
		category := c_type.Uint256{}
		value := c_type.Uint256{}
		copy(category[:], c[:32])
		copy(value[:], c[32:])
		if v.count != 0 {
			e = fmt.Errorf("category %v value %v not balance", utils.BytesToCurrency(category[:]), hex.EncodeToString(value[:]))
			return
		}
	}
	return
}

func (self *CKState) Check() (e error) {
	if e = self.CheckToken(); e != nil {
		return
	}
	if e = self.CheckTicket(); e != nil {
		return
	}
	return
}
