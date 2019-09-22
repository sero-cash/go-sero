package proofservice

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/zero/wallet/light"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/zero/txs/stx"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/flight"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
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
	Status    int // 1:penging,2 handing, 3 doned, 4:timeout, 5:error

	tx    *stx.T
	param *txtool.GTxParam
}

func newJob(tx *stx.T, param *txtool.GTxParam) *Job {
	return &Job{tx: tx, param: param, Status: 1};
}

var instance *ProofService

type Config struct {
	PKr            c_type.PKr
	MaxWorkNumber  int
	MaxQueueNumber int
	Fee            ServiceFee
	RedisConfig    RedisConfig
}

type ProofService struct {
	rpc    string
	config *Config

	queueChan   chan *Job
	workChan    chan *Job
	jobs        sync.Map
	workNum     int32;
	client      SeroClient
	redisClient *RedisClient
}

func Instance() *ProofService {
	return instance
}

type Backend interface {
	CommitTx(tx *txtool.GTx) error
	CheckNil(Nils []c_type.Uint256) (nilResps []light.NilValue, e error)
}

func NewProofService(rpc string, backend Backend, config *Config) *ProofService {
	proof := &ProofService{
		rpc:       rpc,
		config:    config,
		queueChan: make(chan *Job, config.MaxQueueNumber),
	}
	if rpc != "" {
		proof.client = NewRemoteClient(rpc)
	} else {
		proof.client = NewLocalClient(backend)
	}
	// proof.redisClient = NewRedisClient()
	instance = proof
	go proof.loop()
	log.Info("ProofService start", "config:", config)
	return proof
}

func (self *ProofService) FindTxHash(hash common.Hash) common.Hash {
	job := self.redisClient.GetJob(hash)
	if job != nil {
		return job.TxHash
	}
	return common.Hash{}
}

var sero = *common.BytesToHash(common.LeftPadBytes([]byte("SERO"), 32)).HashToUint256()

func (self *ProofService) checkFee(param *txtool.GTxParam) bool {
	if self.config.Fee.FixedFee.Sign() > 0 {
		for _, out := range param.Outs {
			if out.PKr == self.config.PKr {
				if out.Asset.Tkn != nil {
					if out.Asset.Tkn.Currency == sero {
						amount := out.Asset.Tkn.Value.ToInt()
						return amount.Cmp(self.config.Fee.FixedFee) >= 0
					}
				}
				break
			}
		}
	} else {
		fee := big.NewInt(0)
		for _, in := range param.Ins {
			if in.Out.State.OS.Out_P != nil {
				fee = new(big.Int).Add(fee, self.config.Fee.ZinFee)
			} else if in.Out.State.OS.Out_C != nil {
				fee = new(big.Int).Add(fee, self.config.Fee.OinFee)
			} else {
				return false
			}
		}
		fee = new(big.Int).Add(fee, new(big.Int).Mul(self.config.Fee.OutFee, big.NewInt(int64(len(param.Outs)-1))))
		for _, out := range param.Outs {
			if out.PKr == self.config.PKr {
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

func (self *ProofService) Fee() ServiceFee {
	return self.config.Fee
}

func (self *ProofService) SubmitWork(tx *stx.T, param *txtool.GTxParam) bool {
	hash := tx.Tx1_Hash()
	if !self.checkFee(param) {
		log.Error("check fee error", "txHash", common.Bytes2Hex(hash[:]))
		return false
	}

	if self.redisClient.Exists(common.BytesToHash(hash[:])) {
		job := newJob(tx, param)
		if TryEnqueue(job, self.queueChan) {
			self.redisClient.SetJob(job);
			return true
		}
	}
	return false
}

type SeroClient interface {
	CheckNils(nils []c_type.Uint256) bool
	CommitTx(tx *txtool.GTx) error
}

func (self *ProofService) processJob(job *Job) {
	job.Status = 2;
	gtx, err := flight.ProveTx1(job.tx, job.param)
	if err != nil {
		log.Error("processJob error", "error", err)
		job.Error = err
		job.Status = 5;
		self.redisClient.UpdateJob(job);
		return
	}
	if err := self.client.CommitTx(&gtx); err != nil {
		log.Error("processJob error", "error", err)
		job.Error = err
		job.Status = 5
		self.redisClient.UpdateJob(job);
		return
	}
	txHash := gtx.Tx.ToHash()
	job.TxHash = common.BytesToHash(txHash[:])
	job.Status = 3;
	self.redisClient.UpdateJob(job);
}

func (self *ProofService) loop() {
	clear := time.NewTicker(time.Minute * 10)
	defer clear.Stop()

	for {
		for self.workNum >= 5 {
			time.Sleep(time.Second)
		}
		select {
		case job := <-self.queueChan:
			if job.Status == 1 {
				atomic.AddInt32(&self.workNum, 1)
				go func() {
					defer atomic.AddInt32(&self.workNum, -1);
					self.processJob(job)
				}()
			}
		case <-clear.C:
			self.jobs.Range(func(key, value interface{}) bool {
				job := value.(*Job)
				if job.Timestamp.Add(time.Hour * 2).Before(time.Now()) {
					if job.Status == 1 || job.Status == 2 {
						job.Status = 4;
					}

					self.jobs.Delete(key);
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
