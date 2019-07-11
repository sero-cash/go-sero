package light

import (
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-czero-import/keys"
	"encoding/binary"
	"fmt"
	"github.com/sero-cash/go-sero/rlp"
	"sync/atomic"
	"github.com/robfig/cron"
	"github.com/sero-cash/go-sero/zero/txtool/flight"
	"github.com/sero-cash/go-sero/zero/txtool"
)

type LightNode struct {
	db     *serodb.LDBDatabase
	txPool *core.TxPool

	sri flight.SRI

	lastNumber uint64
}

var (
	pkrPrefix = []byte("PKr")
	nilPrefix = []byte("NIL")
)

func NewLightNode(dbPath string, txPool *core.TxPool) (lightNode *LightNode) {

	db, err := serodb.NewLDBDatabase(dbPath, 1024, 1024)
	if err != nil {
		panic(err)
	}
	lightNode = &LightNode{
		txPool: txPool,
		sri:    flight.SRI_Inst,
		db:     db,
	}
	current_light = lightNode

	AddJob("0/10 * * * * ?", lightNode.fetchBlockInfo)

	log.Info("Init NewLightNode success")
	return
}

var fetchCount = uint64(5000)

func (self *LightNode) getLastNumber() (num uint64) {
	if self.lastNumber == 0 {
		value, err := self.db.Get(numKey())
		if err != nil {
			return 0
		}
		self.lastNumber = bytesToUint64(value)
	}
	return self.lastNumber

}

func numKey() []byte {
	return []byte("LIGHT_SYNC_NUM")
}

func (self *LightNode) fetchBlockInfo() {
	if txtool.Ref_inst.Bc == nil || !txtool.Ref_inst.Bc.IsValid() {
		return
	}

	fmt.Println("fetchBlockInfo begin")
	start := self.getLastNumber()
	fmt.Println("fetchBlockInfo ,start=", start)
	fmt.Println("fetchBlockInfo ,GetBlocksInfo=", start+1, fetchCount)
	blocks, err := self.sri.GetBlocksInfo(start+1, fetchCount)
	if err != nil {
		log.Error("light GetBlocksInfo err:", err.Error())

	}
	fmt.Println("fereturntchBlockInfo ,GetBlocksInfo, len(blocks=", len(blocks))
	if len(blocks) == 0 {
		return
	}
	var count uint64 = 0
	batch := self.db.NewBatch()
	for _, block := range blocks {
		// PKR -> Outs
		outs := block.Outs

		pkrMap := make(map[keys.PKr][]txtool.Out)

		for _, out := range outs {
			var pkr keys.PKr
			if out.State.OS.Out_Z != nil {
				pkr = out.State.OS.Out_Z.PKr
			}
			if out.State.OS.Out_O != nil {
				pkr = out.State.OS.Out_O.Addr
			}
			if value,ok := pkrMap[pkr];ok {
				v:=value
				v = append(v,out)
				pkrMap[pkr]= v
			}else{
				pkrMap[pkr]= []txtool.Out{out}
			}

		}
		for pkr,v := range pkrMap{
			data, err := rlp.EncodeToBytes(v)
			if err != nil {
				return
			}
			batch.Put(pkrKey(pkr.ToUint512(), uint64(block.Num)), data)
		}

		nils := block.Nils
		if len(nils) > 0 {
			for _, Nil := range nils {
				batch.Put(nilKey(Nil,uint64(block.Num)), uint64ToBytes(1))
			}
		}
		count ++
	}
	if count == 0 {
		return
	}

	lastNumber := self.lastNumber
	if count < fetchCount {
		lastNumber = start + count
	} else {
		lastNumber = start + fetchCount
	}
	batch.Put(numKey(), uint64ToBytes(lastNumber))
	err = batch.Write()
	if err == nil {
		self.lastNumber = lastNumber
	}
	return
}

func uint64ToBytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func bytesToUint64(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

func nilKey(Nil keys.Uint256, num uint64) []byte {
	key := append(nilPrefix, Nil[:]...)
	return append(key, uint64ToBytes(num)...)
}

func pkrKey(pkr keys.Uint512, num uint64) []byte {
	key := append(pkrPrefix, pkr[:]...)
	return append(key, uint64ToBytes(num)...)
}

func AddJob(spec string, run RunFunc) *cron.Cron {
	c := cron.New()
	c.AddJob(spec, &RunJob{run: run})
	c.Start()
	return c
}

type (
	RunFunc func()
)

type RunJob struct {
	runing int32
	run    RunFunc
}

func (r *RunJob) Run() {
	x := atomic.LoadInt32(&r.runing)
	if x == 1 {
		return
	}

	atomic.StoreInt32(&r.runing, 1)
	defer func() {
		atomic.StoreInt32(&r.runing, 0)
	}()

	r.run()
}
