// Copyright 2015 The sero.cash Authors
// This file is part of the go-sero library.
//
// The go-sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sero library. If not, see <http://www.gnu.org/licenses/>.

package txs

import (
    "github.com/sero-cash/go-sero/zero/txs/zstate"
    "github.com/sero-cash/go-sero/zero/txs/tx"
    "github.com/sero-cash/go-czero-import/keys"
    "errors"
    "fmt"
    "sort"
    "github.com/sero-cash/go-sero/zero/utils"
)


type z2zs struct {
    z2z map[keys.Uint256]utils.I256
}

func newZ2Zs() (ret z2zs) {
    ret.z2z=make(map[keys.Uint256]utils.I256)
    return
}

func (self *z2zs) get(k *keys.Uint256) (utils.I256){
    if k==nil {
        return utils.NewI256(0)
    }
    z,ok:=self.z2z[*k]
    if !ok {
        self.z2z[*k]=utils.NewI256(0)
        return utils.NewI256(0)
    } else {
        return *z.ToRef()
    }
}
func (self *z2zs) add(k *keys.Uint256,z2z *utils.I256) {
    z,ok:=self.z2z[*k]
    if !ok {
        self.z2z[*k]=*z2z.ToRef()
    } else {
        z.AddI(z2z)
        self.z2z[*k]=*z.ToRef()
    }
}

func (self *z2zs) del(k *keys.Uint256,z2z *utils.I256) {
    z,ok:=self.z2z[*k]
    if !ok {
        z=utils.NewI256(0)
        z.SubI(z2z)
        self.z2z[*k]=*z.ToRef()
    } else {
        z.SubI(z2z)
        self.z2z[*k]=*z.ToRef()
    }
}

func (self *z2zs) clone() (ret z2zs){
    ret=newZ2Zs()
    for k,v:=range self.z2z {
        ret.z2z[k]=*v.ToRef()
    }
    return
}

func (self *z2zs) clear() {
    for k,_:=range self.z2z {
        self.z2z[k]=utils.NewI256(0);
    }
    return
}

func (self *z2zs) sortcut(l int)(ret []utils.I256) {
    ks:=utils.Uint256s{}
    for k,_:=range self.z2z {
        ks=append(ks,k)
    }
    sort.Sort(ks)
    ks=utils.Uint256SliceCut(ks,l)
    for _,currency:=range ks {
        ret=append(ret,*self.z2z[currency].ToRef())
    }
    for len(ret)<l {
        ret=append(ret, utils.NewI256(0))
    }
    return
}


type preTxDesc_Z struct {
    currency keys.Uint256
    in *zstate.OutState1
    out *tx.Out
    z2z z2zs
}


type preTxDesc_O struct {
    currency keys.Uint256
    fee utils.U256
    z2o utils.I256
    ins []zstate.OutState1
    outs []tx.Out
}

type c2i struct {
    c2index map[keys.Uint256]int
    currencys []keys.Uint256
}

func newC2I() (ret c2i) {
    ret.c2index=make(map[keys.Uint256]int)
    return
}

func (self *c2i)addI(index int) (ret keys.Uint256){
    if index>=len(self.currencys) {
        panic("Currency index is too big")
    }
    if len(self.currencys)>index {
        copy(ret[:],self.currencys[index][:])
        return
    } else {
        return
    }
}
func (self *c2i)addC(currency* keys.Uint256) int {
    if currency==nil {
        panic("Currency is nil")
    }
    if index,ok:=self.c2index[*currency];!ok {
        ret:=len(self.currencys)
        self.c2index[*currency]=ret
        self.currencys=append(self.currencys,*currency)
        return ret
    } else {
        return index
    }
}

type preTx struct {
    last_anchor keys.Uint256
    uouts []zstate.OutState1
    desc_zs []preTxDesc_Z
    desc_os map[keys.Uint256]*preTxDesc_O
    C2I c2i
}

type preOut struct {
    Out tx.Out
    currency keys.Uint256
}

type tempZDesc struct {
    z_ins []zstate.OutState1
    z_outs []preOut
}

func preGen(ts *tx.T,state *zstate.State) (p preTx,e error) {
    state1:=zstate.LoadState1(&state.State0)
    p.desc_os=make(map[keys.Uint256]*preTxDesc_O)
    p.last_anchor=state.State0.Cur.Tree.RootKey()
    z_temp_descs:=make(map[keys.Uint256]*tempZDesc)
    p.C2I=newC2I()
    for _,t:=range ts.CTxs {
        p.C2I.addC(&t.Currency)
        
        temp_desc,ok:=z_temp_descs[t.Currency]
        if !ok {
            temp_desc=&tempZDesc{}
            z_temp_descs[t.Currency]=temp_desc
        } else {}
        
        desc_o, ok := p.desc_os[t.Currency]
        if !ok {
            desc_o = &preTxDesc_O{}
            desc_o.currency = t.Currency
            p.desc_os[t.Currency] = desc_o
        }
        
        desc_o.fee.AddU(&t.Fee)
        
        for _, in := range t.Ins {
            if src, err := state1.GetOut(&in.Root); err == nil {
                if src.Out_O.Currency != t.Currency {
                    e = errors.New("currency type not match!")
                    return
                }
                if src.Out_O.Out.Value.Cmp(&utils.U256_0)==0 {
                } else {
                    if src.Desc_Z == nil {
                        desc_o.ins = append(desc_o.ins, *src)
                    } else {
                        temp_desc.z_ins = append(temp_desc.z_ins, *src)
                    }
                }
                p.uouts = append(p.uouts, *src)
            } else {
                e = err
                return
            }
        }
        
        for _, outdata := range t.Outs {
            switch outdata.Z {
            case tx.TYPE_N:
                fallthrough
            case tx.TYPE_O:
                desc_o.outs = append(desc_o.outs, outdata)
            default:
                temp_desc.z_outs = append(temp_desc.z_outs, preOut{
                    outdata,
                    t.Currency,
                })
            }
        }
    }
    
    z2z:=newZ2Zs()
    for currency,temp_desc:=range z_temp_descs {
        z_ins:=temp_desc.z_ins
        z_outs:=temp_desc.z_outs
        
        var desc_n int
        if len(z_ins)>len(z_outs) {
            desc_n=len(z_ins)
        } else {
            desc_n=len(z_outs)
        }
    
        for i:=0;i<desc_n;i++ {
            var desc_z preTxDesc_Z
            desc_z.currency=currency
            if len(z_ins)>i {
                desc_z.in=&z_ins[i]
                z2z.add(
                    &desc_z.in.Out_O.Currency,
                    desc_z.in.Out_O.Out.Value.ToI256().ToRef(),
                )
            } else {}
            if len(z_outs)>i {
                desc_z.out=&z_outs[i].Out
                z2z.del(
                    &z_outs[i].currency,
                    desc_z.out.Value.ToI256().ToRef(),
                )
            } else {}
            desc_z.z2z=z2z.clone()
            p.desc_zs=append(p.desc_zs,desc_z)
        }
    }
    
    if len(p.desc_zs)>0 {
        last:=p.desc_zs[len(p.desc_zs)-1]
        for k,v:=range last.z2z.z2z {
            if desc_o,ok:=p.desc_os[k];!ok {
                e=errors.New("can not find desc_o recive last.z2z");
                return;
            } else {
                desc_o.z2o=v
            }
        }
        last.z2z.clear()
    }
    
    o2o:=newZ2Zs()
    for currency,desc_o:=range p.desc_os {
        o2o.add(&currency,&desc_o.z2o)
    
        r_fee:=desc_o.fee.ToI256()
        r_fee.Reverse()
        o2o.add(&currency,&r_fee)
        for _,in:=range desc_o.ins {
            o2o.add(&desc_o.currency,in.Out_O.Out.Value.ToI256().ToRef());
        }
    
        for _,out:=range desc_o.outs {
            o2o.del(&desc_o.currency,out.Value.ToI256().ToRef());
        }
    }
    
    for currency,o:=range o2o.z2z {
        if o.Cmp(&utils.I256_0)!=0 {
            e=fmt.Errorf("currency %v banlance != 0",currency)
            return
        } else {}
    }
    return
}

