package light

import (
	"github.com/sero-cash/go-czero-import/cpt"
	"github.com/sero-cash/go-czero-import/keys"
)

type SLI struct {
}

var SLI_Inst = SLI{}

func (self *SLI) CreateKr() (kr Kr) {
	rnd := keys.RandUint256()
	zsk := keys.RandUint256()
	vsk := keys.RandUint256()
	zsk = cpt.Force_Fr(&zsk)
	vsk = cpt.Force_Fr(&vsk)

	skr := keys.PKr{}
	copy(skr[:], zsk[:])
	copy(skr[32:], vsk[:])
	copy(skr[64:], rnd[:])

	sk := keys.Uint512{}
	copy(sk[:], skr[:])
	pk := keys.Sk2PK(&sk)

	pkr := keys.Addr2PKr(&pk, &rnd)
	kr.PKr = pkr
	kr.SKr = skr
	return
}

func (self *SLI) DecOuts(outs []Out, skr *keys.PKr) (douts []DOut, e error) {
	return
}

func (self *SLI) GenTx(param *GenTxParam) (gtx GTx, e error) {
	return
}
