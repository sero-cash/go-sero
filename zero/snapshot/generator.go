package snapshot

import (
	"fmt"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/rawdb"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/trie"
	"github.com/sero-cash/go-sero/zero/consensus"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/stake"
	"log"
	"os"
)

type SnapshotGen struct {
	src_db *serodb.LDBDatabase
	src_state_db state.Database
	src_head_block_hash common.Hash
	src_head_num int64
	target_db *serodb.LDBDatabase

	target_head_block_hash common.Hash
	target_head_num int64

	blockStop chan bool
	stateStop chan bool
}

func NewSnapshotGen(src string,target string) (ret *SnapshotGen,err error) {
	sg:=SnapshotGen{}
	if sg.src_db,err=serodb.NewLDBDatabaseEx(src,1024*8,1024,true);err!=nil {
		return nil,err
	}
	sg.src_state_db=state.NewDatabase(sg.src_db)

	{
		hash := rawdb.ReadHeadBlockHash(sg.src_db)
		num := *rawdb.ReadHeaderNumber(sg.src_db, hash)
		header := rawdb.ReadHeader(sg.src_db, hash, num)
		for {
			if _, err := state.New(sg.src_state_db, header); err == nil {
				sg.src_head_num = header.Number.Int64()
				sg.src_head_block_hash = header.Hash()
				break
			}
			header = rawdb.ReadHeader(sg.src_db, header.ParentHash, header.Number.Uint64()-1)
		}
	}


	if err=sg.VerifyDB();err!=nil {
		return nil,err
	}

	if err=os.MkdirAll(target,os.ModePerm);err!=nil {
		return nil,err
	}
	if sg.target_db,err=serodb.NewLDBDatabase(target,1024*8,1024);err!=nil {
		return nil,err
	}
	sg.target_head_block_hash=rawdb.ReadHeadBlockHash(sg.target_db)
	if num:=rawdb.ReadHeaderNumber(sg.target_db,sg.target_head_block_hash);num!=nil {
		sg.target_head_num=int64(*num)
	} else {
		sg.target_head_num=-1
	}

	sg.blockStop=make(chan bool)
	sg.stateStop=make(chan bool)
	return &sg,nil
}

func (self *SnapshotGen) Close() {
	self.src_db.Close();
	self.target_db.Close();
}

func (self *SnapshotGen) RunBlock() {
	for {
		if self.ProcessBlock(10000)==true {
			break
		}
		log.Print("Process Block:",self.target_head_num)
	}
	self.blockStop<-true
}

func (self *SnapshotGen) RunState() {
	num:=uint64(self.src_head_num)
	hblock := rawdb.ReadCanonicalHash(self.src_db, num)
	header := rawdb.ReadHeader(self.src_db, hblock, num)
	if ok,count:=self.ProcessState(header.Root);ok {
		log.Print("End Process State:", num,"count=",count)
	}
	self.stateStop<-true
}

func (self *SnapshotGen) Run() {
	go self.RunBlock()
	go self.RunState()

	<-self.blockStop
	<-self.stateStop
}

func (self *SnapshotGen) VerifyDB() error {
	var num uint64 =1310000;
	hash:=rawdb.ReadCanonicalHash(self.src_db,num)
	consKeys:=consensus.GetConsKeys(self.src_db,num,hash)
	if len(consKeys)==0 {
		return fmt.Errorf("The src db is not SIP10 Snapshot")
	} else {
		return nil
	}
}

func (self *SnapshotGen) ProcessBlock(step int) (bool) {
	if step==0 {
		return true
	}
	if self.target_head_num>=self.src_head_num {
		return true
	}
	if self.target_head_num+int64(step)>self.src_head_num {
		step=int(self.src_head_num-self.target_head_num)
	}
	batch:=self.target_db.NewBatch()
	for i:=0;i<step;i++ {
		self.target_head_num++
		num:=uint64(self.target_head_num);

		self.target_head_block_hash = rawdb.ReadCanonicalHash(self.src_db, num)
		hblock:=self.target_head_block_hash
		rawdb.WriteCanonicalHash(batch, hblock, num)

		if self.target_head_num==0 {
			config:=rawdb.ReadChainConfig(self.src_db,hblock)
			rawdb.WriteChainConfig(batch,hblock,config)
			ver:=rawdb.ReadDatabaseVersion(self.src_db)
			rawdb.WriteDatabaseVersion(batch,ver)
		}

		header:=rawdb.ReadHeader(self.src_db,hblock,num)
		rawdb.WriteHeader(batch,header)

		body:=rawdb.ReadBody(self.src_db,hblock,num)
		rawdb.WriteBody(batch,hblock,num,body)

		block:=types.NewBlockWithHeader(header).WithBody(body.Transactions)
		rawdb.WriteTxLookupEntries(batch,block)

		receipts:=rawdb.ReadReceipts(self.src_db,hblock,num)
		rawdb.WriteReceipts(batch,hblock,num,receipts)

		td:=rawdb.ReadTd(self.src_db,hblock,num)
		rawdb.WriteTd(batch,hblock,num,td)


		lcBlock:=localdb.GetBlock(self.src_db,num,hblock.HashToUint256())
		localdb.PutBlock(batch,num,hblock.HashToUint256(),lcBlock)
		for _,pkg:=range lcBlock.Pkgs {
			p:=localdb.GetPkg(self.src_db,&pkg)
			localdb.PutPkg(batch,&pkg,p)
		}
		for _,root:=range lcBlock.Roots {
			r:=localdb.GetRoot(self.src_db,&root)
			if r==nil {
				if num>=seroparam.SIP2() {
					panic("root state err")
				}
			} else {
				localdb.PutRoot(batch,&root,r)
			}
		}

		consKeys := consensus.GetConsKeys(self.src_db, num, hblock)
		if len(consKeys) > 0 {
			consensus.PutConsKeys(batch, num, hblock, consKeys)
			for _, key := range consKeys {
				if v, err := self.src_db.Get(key); err != nil {
					panic(err)
				} else {
					batch.Put(key, v)
				}
			}
		}

		bsnkey := stake.BlockShareNumKey(hblock)
		if v, err := self.src_db.Get(bsnkey); err != nil {
		} else {
			batch.Put(bsnkey, v)
		}

		bvkey := stake.BlockVotesKey(hblock)
		if v, err := self.src_db.Get(bvkey); err != nil {
		} else {
			batch.Put(bvkey, v)
		}

	}
	rawdb.WriteHeadBlockHash(batch,self.target_head_block_hash)
	rawdb.WriteHeadHeaderHash(batch,self.target_head_block_hash)
	rawdb.WriteHeadFastBlockHash(batch,self.target_head_block_hash)
	batch.Write()
	return false
}

func (self *SnapshotGen) ProcessState(root common.Hash) (bool,int) {
	const batch_num int=1024*10;
	sched := state.NewStateSync(root,self.target_db)
	queue := append([]common.Hash{}, sched.Missing(batch_num)...)
	count:=0
	for len(queue) > 0 {
		results := make([]trie.SyncResult, len(queue))
		const c int=1024
		input:=make(chan int)
		end:=make(chan bool)
		output:=make(chan bool)
		for i:=0;i<c;i++ {
			go func() {
				DONE:
				for {
					select {
					case index := <-input:
						hash := queue[index]
						data, err := self.src_state_db.TrieDB().Node(hash)
						if err != nil {
							panic("tri get node error")
						}
						results[index] = trie.SyncResult{Hash: hash, Data: data}
					case <-end:
						break DONE
					}
				}
				output <- true
			}()
		}
		for i:=0;i<len(queue);i++ {
			input<-i
		}
		for i:=0;i<c;i++ {
			end<-true
		}
		for i:=0;i<c;i++ {
			<-output
		}
		if _, _, err := sched.Process(results); err != nil {
			panic("sched process error")
		}
		if ct, err := sched.Commit(self.target_db); err != nil {
			panic("sched commit error")
		} else {
			log.Print("trie:",ct)
			count+=ct
		}
		queue = append(queue[:0], sched.Missing(batch_num)...)
	}
	return true,count
}
