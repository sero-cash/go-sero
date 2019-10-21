package proofservice

import (
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/rlp"
	"gopkg.in/redis.v3"
	"time"
)

type RedisConfig struct {
	Endpoint string `json:"endpoint"`
	Password string `json:"password"`
	Database int64  `json:"database"`
	PoolSize int    `json:"poolSize"`
}

type RedisClient struct {
	client *redis.Client
	prefix string
}

func NewRedisClient(cfg *RedisConfig, prefix string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Endpoint,
		Password: cfg.Password,
		DB:       cfg.Database,
		PoolSize: cfg.PoolSize,
	})
	return &RedisClient{client: client, prefix: prefix}
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}

func (r *RedisClient) Check() (string, error) {
	return r.client.Ping().Result()
}

func (r *RedisClient) BgSave() (string, error) {
	return r.client.BgSave().Result()
}

func (r *RedisClient) Exists(txHash common.Hash) bool {
	ret := r.client.Exists(r.formatKey(txHash))
	return ret.Val()
}

// func (r *RedisClient) Save(job *Job) bool {
// 	bytes, err := rlp.EncodeToBytes(job)
// 	if err != nil {
// 		return false
// 	}
// 	ret := r.client.SetNX(r.formatKey(job.TxHash), string(bytes), 0)
// 	return ret.Val()
// }

func (r *RedisClient) Save(job *Job) {
	bytes, err := rlp.EncodeToBytes(job)
	if err != nil {
		log.Error("redis save err", err)
	}
	r.client.Set(r.formatKey(job.TxHash), string(bytes), time.Hour*2)
}

func (r *RedisClient) Get(txHash common.Hash) *Job {
	ret := r.client.Get(r.formatKey(txHash))
	if datas, err := ret.Bytes(); err == nil {
		job := Job{}
		err = rlp.DecodeBytes(datas, &job)
		if err != nil {
			log.Error("GetJob", err)
			return nil
		}
		return &job;
	} else {
		log.Error("GetJob", err)
		return nil
	}
}

func (r *RedisClient) formatKey(txHash common.Hash) string {
	return r.prefix + ":" + txHash.Hex()
}
