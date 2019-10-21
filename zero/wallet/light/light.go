package light

import (
	"encoding/binary"
	"math/big"
	"sync/atomic"

	"github.com/robfig/cron"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core"
	"github.com/sero-cash/go-sero/core/rawdb"
	"github.com/sero-cash/go-sero/log"
	"github.com/sero-cash/go-sero/rlp"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/txtool/flight"
)

type LightNode struct {
	db   *serodb.LDBDatabase
	bcDB serodb.Database

	txPool *core.TxPool

	sri flight.SRI

	lastNumber uint64
}

var (
	teamReward = common.Hash{}
	powReward  = common.BytesToHash([]byte{1})
	posReward  = common.BytesToHash([]byte{2})
	posMiner   = common.BytesToHash([]byte{3})
)

var (
	pkrPrefix = []byte("PKr")
	nilPrefix = []byte("NIL")
)

func NewLightNode(dbPath string, txPool *core.TxPool, bcDB serodb.Database) (lightNode *LightNode) {

	db, err := serodb.NewLDBDatabase(dbPath, 1024, 1024)
	if err != nil {
		panic(err)
	}
	lightNode = &LightNode{
		txPool: txPool,
		sri:    flight.SRI_Inst,
		db:     db,
		bcDB:   bcDB,
	}
	current_light = lightNode

	AddJob("0/10 * * * * ?", lightNode.fetchBlockInfo)

	log.Info("Init NewLightNode success")
	return
}

var fetchCount = uint64(5000)

func (self *LightNode) getLastNumber() (num uint64) {

	if self.lastNumber == 0 {
		// light wallet start at block 1200000
		var initBlockNum = uint64(1280000)
		if seroparam.Is_Dev() {
			initBlockNum = uint64(0)
		}
		value, err := self.db.Get(numKey())
		if err != nil {
			self.db.Put(numKey(), uint64ToBytes(initBlockNum))
			return initBlockNum
		}
		self.lastNumber = bytesToUint64(value)
		if self.lastNumber == 0 {
			self.lastNumber = initBlockNum
			self.db.Put(numKey(), uint64ToBytes(initBlockNum))
		}
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
	start := self.getLastNumber()
	blocks, err := self.sri.GetBlocksInfo(start+1, fetchCount)
	if err != nil {
		log.Error("light GetBlocksInfo err:", err.Error())
	}
	if len(blocks) == 0 {
		return
	}
	var count uint64 = 0
	batch := self.db.NewBatch()
	for _, block := range blocks {
		// PKR -> Outs
		outs := block.Outs
		pkrMap := make(map[c_type.PKr][]BlockData)
		blockHash := common.Hash{}
		blockNum := uint64(block.Num)
		copy(blockHash[:], block.Hash[:])
		body := rawdb.ReadBody(self.bcDB, blockHash, blockNum)
		blockDB := rawdb.ReadBlock(self.bcDB, blockHash, blockNum)

		for _, out := range outs {
			txHash := common.Hash{}
			copy(txHash[:], out.State.TxHash[:])

			if (teamReward == txHash) {
				continue
			}

			var txInfo TxInfo
			if !(powReward == txHash || posReward == txHash || posMiner == txHash) {
				// fmt.Println("hex hash::", hexutil.Encode(txHash[:]), out.State.Num)
				txReceipt, _, _, _ := rawdb.ReadReceipt(self.bcDB, txHash)
				tx, _, _, _ := rawdb.ReadTransaction(self.bcDB, txHash)
				gasUsed := txReceipt.GasUsed
				txInfo = TxInfo{
					Num:       blockNum,
					TxHash:    out.State.TxHash,
					BlockHash: blockDB.Hash(),
					Gas:       tx.Gas(),
					GasUsed:   gasUsed,
					GasPrice:  *tx.GasPrice(),
					From:      tx.From(),
					// To:       *tx.To(),
					Time: *blockDB.Time(),
				}
			} else {
				txInfo = TxInfo{
					Num:       blockNum,
					TxHash:    out.State.TxHash,
					BlockHash: blockDB.Hash(),
					// To:       *tx.To(),
					Time: *blockDB.Time(),
				}
			}

			blockData := BlockData{
				TxInfo: txInfo,
				Out:    out,
			}

			pkr := *out.State.OS.ToPKr()
			if value, ok := pkrMap[pkr]; ok {
				v := value
				v = append(v, blockData)
				pkrMap[pkr] = v
			} else {
				pkrMap[pkr] = []BlockData{blockData}
			}
		}
		for pkr, v := range pkrMap {
			data, err := rlp.EncodeToBytes(v)
			if err != nil {
				return
			}
			batch.Put(pkrKey(pkr, uint64(block.Num)), data)
		}
		for _, tx := range body.Transactions {
			hash := tx.Hash()
			txHash := c_type.Uint256{}
			copy(txHash[:], hash[:])

			txReceipt, _, _, _ := rawdb.ReadReceipt(self.bcDB, tx.Hash())
			gasUsed := txReceipt.GasUsed

			// Index Tx Info
			txInfo := TxInfo{
				Num:       blockNum,
				TxHash:    txHash,
				BlockHash: blockDB.Hash(),
				Gas:       tx.Gas(),
				GasUsed:   gasUsed,
				GasPrice:  *tx.GasPrice(),
				From:      tx.From(),
				// To:       *tx.To(),
				Time: *blockDB.Time(),
			}

			nilValue := NilValue{
				Num:    blockNum,
				TxHash: txHash,
				TxInfo: txInfo,
			}
			if nilValue, err := rlp.EncodeToBytes(nilValue); err != nil {
				return
			} else {
				if tx.Stxt().Tx0() != nil {
					for _, in := range tx.Stxt().Tx0().Desc_O.Ins {
						batch.Put(nilKey(in.Nil), nilValue)
						batch.Put(nilKey(in.Root), nilValue)
					}
					for _, in := range tx.Stxt().Tx0().Desc_Z.Ins {
						batch.Put(nilKey(in.Trace), nilValue)
						batch.Put(nilKey(in.Nil), nilValue)
					}
				}
				if tx.Stxt().Tx1.Ins_C != nil {
					for _, in := range tx.Stxt().Tx1.Ins_C {
						batch.Put(nilKey(in.Nil), nilValue)
					}
				}
				if tx.Stxt().Tx1.Ins_P != nil {
					for _, in := range tx.Stxt().Tx1.Ins_P {
						batch.Put(nilKey(in.Nil), nilValue)
						batch.Put(nilKey(in.Root), nilValue)
					}
				}
				if tx.Stxt().Tx1.Ins_P0 != nil {
					for _, in := range tx.Stxt().Tx1.Ins_P0 {
						batch.Put(nilKey(in.Nil), nilValue)
						batch.Put(nilKey(in.Root), nilValue)
						batch.Put(nilKey(in.Trace), nilValue)
					}
				}
			}
		}
		// nils := block.Nils
		// if len(nils) > 0 {
		//	for _, Nil := range nils {
		//		batch.Put(nilKey(Nil, uint64(block.Num)), uint64ToBytes(1))
		//	}
		// }
		count++
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

type NilValue struct {
	Nil    c_type.Uint256
	Num    uint64
	TxHash c_type.Uint256
	TxInfo TxInfo
}

func uint64ToBytes(i uint64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func bytesToUint64(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

func nilKey(Nil c_type.Uint256) []byte {
	return append(nilPrefix, Nil[:]...)
}

func pkrKey(pkr c_type.PKr, num uint64) []byte {
	key := append(pkrPrefix, pkr[:]...)
	return append(key, uint64ToBytes(num)...)
}

func AddJob(spec string, run RunFunc) *cron.Cron {
	c := cron.New()
	c.AddJob(spec, &RunJob{run: run})
	c.Start()
	return c
}

type TxInfo struct {
	TxHash    c_type.Uint256
	Num       uint64
	BlockHash common.Hash
	Gas       uint64
	GasUsed   uint64
	GasPrice  big.Int
	From      common.Address
	To        common.Address
	Time      big.Int
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
