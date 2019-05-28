package state1

/*
func CurrentLState() *State1 {
	if r, ok := lstate.CurrentLState().(*State1); !ok {
		return nil
	} else {
		return r
	}
}

func state1_file_name(num uint64, hash *common.Hash) (ret string) {
	ret = fmt.Sprintf("%010d.%s", num, hexutil.Encode(hash[:])[3:])
	return
}

var STATE1_LAST_NUM_KEY = []byte("LSTATE$STATE1$LAST$NUM$KEY")
var STATE1_LAST_HASH_KEY = []byte("LSTATE$STATE1$LAST$HASH$KEY")

func GetLastNum(getter serodb.Getter) (ret uint64) {
	if bs, err := getter.Get(STATE1_LAST_NUM_KEY); err != nil {
		ret = 0
		return
	} else {
		input := big.NewInt(0)
		if err := input.GobDecode(bs); err != nil {
			panic(err)
			return
		} else {
			ret = input.Uint64()
			return
		}
	}
}

func SetLastNum(putter serodb.Putter, num uint64) {
	if bs, err := big.NewInt(0).SetUint64(num).GobEncode(); err != nil {
		panic(err)
		return
	} else {
		if err := putter.Put(STATE1_LAST_NUM_KEY, bs); err != nil {
			panic(err)
			return
		} else {
			return
		}
	}
}

const delay_block_count = 6

func (self *State1) Parse(last_chose uint64) (chose uint64) {
	self.MakesureEnv()

	bc := light_ref.Ref_inst.Bc
	tks := bc.GetTks()
	next_num := GetLastNum(&self.db)

	last_num, last_file_name := zconfig.Get_State1_last_num_and_hash()
	next_num = uint64(last_num + 1)

	if next_num == 0 {
		current_header := bc.GetCurrenHeader()
		for {
			current_hash := current_header.Hash()
			current_num := current_header.Number.Uint64()
			if need_parse, err := self.needParse(current_num, &current_hash); err != nil {
				time.Sleep(1000 * 1000 * 1000 * 10)
				return
			} else {
				if need_parse {
					next_num = current_header.Number.Uint64()
					parent_hash := current_header.ParentHash
					current_header = bc.GetHeader(&parent_hash)
					if current_header == nil {
						break
					} else {
						hash := current_header.Hash()
						if hash != parent_hash {
							log.Error(
								"current.hash not equal the pre.parent_hash : ",
								"current.h", hexutil.Encode(hash[:])[:8],
								"pre.p_h",
								hexutil.Encode(parent_hash[:])[:8],
							)
							panic("parse block error")
						}
						continue
					}
				} else {
					break
				}
			}
		}
	}

	chose = bc.CashChose().Load().(uint64)
	current_header := bc.GetCurrenHeader()
	current_num := current_header.Number.Uint64()

	if chose == 0 {
		self.begin(last_file_name, nil, tks)
		return current_num
	} else {
		chose = light_ref.Ref_inst.GetDelayedNum(delay_block_count)
		chose_header := bc.GetHeaderByNumber(chose)
		chose_hash := chose_header.Hash()
		if next_num > chose {
			self.begin(last_file_name, &chose_hash, tks)
			return chose
		} else {
			hash := chose_header.Hash()
			self.begin(last_file_name, &hash, tks)
		}

	}

	parse_count := 0
	for parse_count < 2000 && next_num <= chose {
		header := bc.GetHeaderByNumber(next_num)
		hash := header.Hash()

		block := localdb.GetBlock(bc.GetDB(), next_num, hash.HashToUint256())
		if block == nil {
			temp_state := bc.NewState(&hash)
			if temp_state == nil {
				panic(fmt.Sprintf("new zstate error: %v:%v !", next_num, hash))
			} else {
				log.Debug("STATE1_PARSE GO BACK TO STATE: ", "num", next_num, "hash", hash)
			}
			block = &localdb.Block{}
			block.Pkgs = temp_state.Pkgs.GetPkgHashes()
			block.Roots = temp_state.State.GetBlockRoots()
			block.Dels = temp_state.State.GetBlockDels()
		}

		self.update(&header.ParentHash, next_num, &hash, block)

		next_num++
		parse_count++
	}

	if parse_count > 0 {
		self.save()
		SetLastNum(&self.db, next_num)
	}
	if parse_count > 1 {
		log.Info("STATE1 PARSE", "t", chose, "c", next_num-1)
	}

	return chose

}
*/
