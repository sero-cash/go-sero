package ethapi

import (
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/zero/proofservice"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txtool"
)

type ProofServiceApi struct {
}

func NewProofServiceApi() *ProofServiceApi {
	return &ProofServiceApi{}
}

func (nodeApi *ProofServiceApi) Fee() map[string]hexutil.Big {
	fee := proofservice.Instance().Fee()
	ret := make(map[string]hexutil.Big)
	ret["fixedFee"] = hexutil.Big(*fee.FixedFee)
	return ret
}

func (nodeApi *ProofServiceApi) SubmitProofWork(tx *stx.T, param *txtool.GTxParam) error {
	return proofservice.Instance().SubmitWork(tx, param)
}

func (nodeApi *ProofServiceApi) FindTxHash(hash common.Hash) common.Hash {
	return proofservice.Instance().FindTxHash(hash)
}
