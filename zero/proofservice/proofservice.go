package proofservice

import (
	"errors"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/flight"
	"github.com/sero-cash/go-sero/zero/wallet/light"
)

type ServiceFee struct {
	ZinFee   *big.Int
	OinFee   *big.Int
	OutFee   *big.Int
	FixedFee *big.Int
}

type Job struct {
	TxHash    common.Hash
	Timestamp time.Time
	Error     error

	tx    *stx.T
	param *txtool.GTxParam
}

func newJob(tx *stx.T, param *txtool.GTxParam) *Job {
	return &Job{tx: tx, param: param}
}

var instance *ProofService

type Config struct {
	PKr            c_type.PKr
	MaxWorkNumber  int
	MaxQueueNumber int
	Fee            ServiceFee
}

type ProofService struct {
	rpc    string
	config *Config

	queueChan chan *Job
	workChan  chan *Job
	jobs      sync.Map
	workNum   int32
	client    SeroClient
	// redisClient *RedisClient
	storage Storage
}

func Instance() *ProofService {
	return instance
}

type Backend interface {
	CommitTx(tx *txtool.GTx) error
	CheckNil(Nils []c_type.Uint256) (nilResps []light.NilValue, e error)
}

type SeroClient interface {
	CheckNils(nils []c_type.Uint256) bool
	CommitTx(tx *txtool.GTx) error
}

type Storage interface {
	Exists(common.Hash) bool
	Save(job *Job)
	Get(hash common.Hash) *Job
}

type MapStorage struct {
	cache map[common.Hash]*Job
}

func newMapStorage() *MapStorage {
	return &MapStorage{make(map[common.Hash]*Job)}
}

func (storage *MapStorage) Exists(hash common.Hash) bool {
	_, ok := storage.cache[hash]
	return ok
}

func (storage *MapStorage) Save(job *Job) {
	hash := job.tx.Tx1.Tx1_Hash()
	storage.cache[common.BytesToHash(hash[:])] = job
}

func (storage *MapStorage) Get(hash common.Hash) *Job {
	return storage.cache[hash]
}

func NewProofService(rpc string, backend Backend, config *Config) *ProofService {
	proof := &ProofService{
		rpc:       rpc,
		config:    config,
		queueChan: make(chan *Job, config.MaxQueueNumber),
	}

	proof.client = NewLocalClient(backend)
	proof.storage = newMapStorage()

	instance = proof
	go proof.loop()
	log.Info("ProofService start", "config:", config)
	return proof
}

func (proof *ProofService) FindTxHash(hash common.Hash) common.Hash {
	job := proof.storage.Get(hash)
	if job != nil {
		return job.TxHash
	}
	return common.Hash{}
}

var sero = *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256()

func (proof *ProofService) checkFee(param *txtool.GTxParam) bool {
	if proof.config.Fee.FixedFee.Sign() > 0 {
		for _, out := range param.Outs {
			if out.PKr == proof.config.PKr {
				if out.Asset.Tkn != nil {
					if out.Asset.Tkn.Currency == sero {
						amount := out.Asset.Tkn.Value.ToInt()
						return amount.Cmp(proof.config.Fee.FixedFee) >= 0
					}
				}
				break
			}
		}
	} else {
		fee := big.NewInt(0)
		for _, in := range param.Ins {
			if in.Out.State.OS.Out_P != nil {
				fee = new(big.Int).Add(fee, proof.config.Fee.ZinFee)
			} else if in.Out.State.OS.Out_C != nil {
				fee = new(big.Int).Add(fee, proof.config.Fee.OinFee)
			} else {
				return false
			}
		}
		fee = new(big.Int).Add(fee, new(big.Int).Mul(proof.config.Fee.OutFee, big.NewInt(int64(len(param.Outs)-1))))
		for _, out := range param.Outs {
			if out.PKr == proof.config.PKr {
				if out.Asset.Tkn != nil {
					if out.Asset.Tkn.Currency == sero {
						amount := out.Asset.Tkn.Value.ToInt()
						return amount.Cmp(fee) >= 0
					}
				}
				break
			}
		}
	}

	return false
}

func (proof *ProofService) Fee() ServiceFee {
	return proof.config.Fee
}

func (proof *ProofService) SubmitWork(tx *stx.T, param *txtool.GTxParam) error {
	hash := tx.Tx1_Hash()
	if !proof.checkFee(param) {
		log.Error("check fee error", "txHash", common.Bytes2Hex(hash[:]))
		errors.New("checkFee error")
	}

	if proof.storage.Exists(common.BytesToHash(hash[:])) {
		log.Warn("already exists", "txHash", common.Bytes2Hex(hash[:]))
		return errors.New("already exists")
	}

	job := newJob(tx, param)
	if TryEnqueue(job, proof.queueChan) {
		proof.storage.Save(job)
		return nil
	}
	return errors.New("server is busy")
}

func (proof *ProofService) processJob(job *Job) {
	gtx, err := flight.ProveTx1(job.tx, job.param)
	if err != nil {
		log.Error("processJob error", "error", err)
		job.Error = err
		proof.storage.Save(job)
		return
	}
	if err := proof.client.CommitTx(&gtx); err != nil {
		log.Error("processJob error", "error", err)
		job.Error = err
		proof.storage.Save(job)
		return
	}
	txHash := gtx.Tx.ToHash()
	job.TxHash = common.BytesToHash(txHash[:])
	proof.storage.Save(job)
}

func (proof *ProofService) loop() {
	clear := time.NewTicker(time.Minute * 10)
	defer clear.Stop()

	for {
		for proof.workNum >= 5 {
			time.Sleep(time.Second)
		}
		select {
		case job := <-proof.queueChan:
			atomic.AddInt32(&proof.workNum, 1)
			go func() {
				defer atomic.AddInt32(&proof.workNum, -1)
				proof.processJob(job)
			}()
		case <-clear.C:
			proof.jobs.Range(func(key, value interface{}) bool {
				job := value.(*Job)
				if job.Timestamp.Add(time.Hour * 2).Before(time.Now()) {
					proof.jobs.Delete(key)
				}
				return true
			})
		}
	}
}

func TryEnqueue(job *Job, jobChan chan *Job) bool {
	select {
	case jobChan <- job:
		return true
	default:
		return false
	}
}
