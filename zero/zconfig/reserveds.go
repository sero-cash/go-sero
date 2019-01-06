package zconfig

type Grid struct {
	MaxNum uint64
	Files  []string
}
type Reserveds struct {
	Gs     [32]Grid
	Target uint64
}

func NewReserveds(target uint64) (ret Reserveds) {
	ret.Target = target
	return
}

func log2(num uint64) (log int) {
	for ; num != 0; log++ {
		num >>= 1
	}
	return
}

func (self *Reserveds) Insert(num uint64, file string) (files []string) {
	if num > self.Target {
		return
	} else {
		if self.Target-num < 64 {
			return
		} else {
			diff := log2(self.Target - 64 - num)
			if diff < 32 {
				if self.Gs[diff].MaxNum > num {
					return []string{file}
				} else {
					if self.Gs[diff].MaxNum < num {
						files = self.Gs[diff].Files
						self.Gs[diff].Files = []string{file}
						self.Gs[diff].MaxNum = num
						return
					} else {
						self.Gs[diff].Files = append(self.Gs[diff].Files, file)
						return
					}
				}
			} else {
				return []string{file}
			}
		}
	}
}
