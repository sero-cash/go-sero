package snapshot

import (
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/rawdb"
	"github.com/sero-cash/go-sero/core/state"
	"github.com/sero-cash/go-sero/serodb"
	"github.com/sero-cash/go-sero/trie"
	"github.com/sero-cash/go-sero/zero/consensus"
	"github.com/sero-cash/go-sero/zero/localdb"
	"github.com/sero-cash/go-sero/zero/stake"
	"os"
	"testing"
)


/*
func TestHeadStorage(t *testing.T) {
	db := serodb.NewMemDatabase()

	blockHead := types.NewBlockWithHeader(&types.Header{Extra: []byte("test block header")})
	blockFull := types.NewBlockWithHeader(&types.Header{Extra: []byte("test block full")})
	blockFast := types.NewBlockWithHeader(&types.Header{Extra: []byte("test block fast")})

	// Check that no head entries are in a pristine database
	if entry := ReadHeadHeaderHash(db); entry != (common.Hash{}) {
		t.Fatalf("Non head header entry returned: %v", entry)
	}
	if entry := ReadHeadBlockHash(db); entry != (common.Hash{}) {
		t.Fatalf("Non head block entry returned: %v", entry)
	}
	if entry := ReadHeadFastBlockHash(db); entry != (common.Hash{}) {
		t.Fatalf("Non fast head block entry returned: %v", entry)
	}
	// Assign separate entries for the head header and block
	WriteHeadHeaderHash(db, blockHead.Hash())
	WriteHeadBlockHash(db, blockFull.Hash())
	WriteHeadFastBlockHash(db, blockFast.Hash())

	// Check that both heads are present, and different (i.e. two heads maintained)
	if entry := ReadHeadHeaderHash(db); entry != blockHead.Hash() {
		t.Fatalf("Head header hash mismatch: have %v, want %v", entry, blockHead.Hash())
	}
	if entry := ReadHeadBlockHash(db); entry != blockFull.Hash() {
		t.Fatalf("Head block hash mismatch: have %v, want %v", entry, blockFull.Hash())
	}
	if entry := ReadHeadFastBlockHash(db); entry != blockFast.Hash() {
		t.Fatalf("Fast head block hash mismatch: have %v, want %v", entry, blockFast.Hash())
	}
}
*/

type SnapshotGen struct {
	src_db *serodb.LDBDatabase
	src_state_db state.Database
	src_head_block_hash common.Hash
	src_head_num int64
	target_db *serodb.LDBDatabase
	target_head_block_hash common.Hash
	target_head_num int64
}

func NewSnapshotGen(src string,target string) (ret *SnapshotGen,err error) {
	sg:=SnapshotGen{}
	if sg.src_db,err=serodb.NewLDBDatabase(src,1024,1024);err!=nil {
		return nil,err
	}
	sg.src_head_block_hash = rawdb.ReadHeadBlockHash(sg.src_db)
	sg.src_head_num=int64(*rawdb.ReadHeaderNumber(sg.src_db,sg.src_head_block_hash))
	sg.src_state_db=state.NewDatabase(sg.src_db)

	os.MkdirAll(target,os.ModePerm)
	if sg.target_db,err=serodb.NewLDBDatabase(target,1024,1024);err!=nil {
		return nil,err
	}
	sg.target_head_block_hash=rawdb.ReadHeadBlockHash(sg.target_db)
	if num:=rawdb.ReadHeaderNumber(sg.target_db,sg.target_head_block_hash);num!=nil {
		sg.target_head_num=int64(*num)
	} else {
		sg.target_head_num=-1
	}
	return &sg,err
}

func (self *SnapshotGen) Close() {
	self.src_db.Close();
	self.target_db.Close();
}

func (self *SnapshotGen) Process(step int) (bool) {
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

		header:=rawdb.ReadHeader(self.src_db,hblock,num)
		rawdb.WriteHeader(batch,header)

		body:=rawdb.ReadBody(self.src_db,hblock,num)
		rawdb.WriteBody(batch,hblock,num,body)

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
			localdb.PutRoot(batch,&root,r)
		}

		consKeys:=consensus.GetConsKeys(self.src_db,num,hblock)
		for _,key:=range consKeys {
			if v,err:=self.src_db.Get(key);err!=nil {
				panic(err)
			} else {
				batch.Put(key,v)
			}
		}

		bsnkey:=stake.BlockShareNumKey(hblock)
		if v,err:=self.src_db.Get(bsnkey);err!=nil {
		} else {
			batch.Put(bsnkey,v)
		}

		bvkey:=stake.BlockVotesKey(hblock)
		if v,err:=self.src_db.Get(bvkey);err!=nil {
		} else {
			batch.Put(bvkey,v)
		}


		self.ProcessState(header.Root,batch);


	}
	rawdb.WriteHeadBlockHash(batch,self.target_head_block_hash)
	rawdb.WriteHeadHeaderHash(batch,self.target_head_block_hash)
	rawdb.WriteHeadFastBlockHash(batch,self.target_head_block_hash)
	batch.Write()
	return false
}

func (self *SnapshotGen) ProcessState(root common.Hash,batch serodb.Batch) {
	sched := state.NewStateSync(root,self.target_db)
	queue := append([]common.Hash{}, sched.Missing(1024)...)
	for len(queue) > 0 {
		results := make([]trie.SyncResult, len(queue))
		for i, hash := range queue {
			data, err := self.src_state_db.TrieDB().Node(hash)
			if err != nil {
				if len(queue)==1 {
					return
				} else {
					panic(err)
				}
			}
			results[i] = trie.SyncResult{Hash: hash, Data: data}
		}
		if _, _, err := sched.Process(results); err != nil {
			panic("sched process error")
		}
		queue = append(queue[:0], sched.Missing(1024)...)
	}
	if _, err := sched.Commit(batch); err != nil {
		panic("sched commit error")
	}

}

func TestSnapshot(t *testing.T) {
	sg,_:=NewSnapshotGen(
		"/Users/tangzhige/Documents/gero/test/datadir/gero/chaindata_kk",
		"/Users/tangzhige/Documents/gero/test/datadir/gero/chaindata",
	)
	sg.Process(int(sg.src_head_num+1))
}
/*
func testIterativeStateSync(t *testing.T, batch int) {
	// Create a random state to copy
	srcDb, srcRoot, srcAccounts := makeTestState()

	// Create a destination state and sync with the scheduler
	dstDb := serodb.NewMemDatabase()
	sched := NewStateSync(srcRoot, dstDb)

	queue := append([]common.Hash{}, sched.Missing(batch)...)
	for len(queue) > 0 {
		results := make([]trie.SyncResult, len(queue))
		for i, hash := range queue {
			data, err := srcDb.TrieDB().Node(hash)
			if err != nil {
				t.Fatalf("failed to retrieve node data for %x", hash)
			}
			results[i] = trie.SyncResult{Hash: hash, Data: data}
		}
		if _, index, err := sched.Process(results); err != nil {
			t.Fatalf("failed to process result #%d: %v", index, err)
		}
		if index, err := sched.Commit(dstDb); err != nil {
			t.Fatalf("failed to commit data #%d: %v", index, err)
		}
		queue = append(queue[:0], sched.Missing(batch)...)
	}
	// Cross check that the two states are in sync
	checkStateAccounts(t, dstDb, srcRoot, srcAccounts)
}
*/