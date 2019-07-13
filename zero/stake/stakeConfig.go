package stake

import (
	"github.com/sero-cash/go-czero-import/seroparam"
	"math/big"
)

var (
	poolValueThreshold, _ = new(big.Int).SetString("200000000000000000000000", 10)
	lockingBlockNum       = uint64(1088640)

	basePrice = big.NewInt(2000000000000000000)
	addition  = big.NewInt(759357240838722)

	baseReware, _ = new(big.Int).SetString("10500000000000000000", 10)
	rewareStep    = big.NewInt(76854301391338)

	maxReware, _ = new(big.Int).SetString("35600000000000000000", 10)

	//outOfDateWindow      = uint64(544320)
	//missVotedWindow      = uint64(725760)
	//payWindow            = uint64(42336)
	//statisticsMissWindow = uint64(6048)

	//test
	outOfDateWindow      = uint64(100)
	missVotedWindow      = uint64(120)
	payWindow            = uint64(5)
	statisticsMissWindow = uint64(10)
)

func GetPoolValueThreshold() *big.Int {
	if seroparam.Is_Dev() {
		return big.NewInt(1000000000000000000)
	}
	return poolValueThreshold
}

func GetLockingBlockNum() uint64 {
	if seroparam.Is_Dev() {
		return 10
	}
	return lockingBlockNum
}

func getStatisticsMissWindow() uint64 {
	if seroparam.Is_Dev() {
		return 10
	}
	return statisticsMissWindow
}

func getOutOfDateWindow() uint64 {
	if seroparam.Is_Dev() {
		return 100
	}
	return outOfDateWindow
}

func getMissVotedWindow() uint64 {
	if seroparam.Is_Dev() {
		return 105
	}
	return missVotedWindow
}

func getPayPeriod() uint64 {
	if seroparam.Is_Dev() {
		return 5
	}
	return payWindow
}

const (
	SOLO_RATE        = 3
	TOTAL_RATE       = 4
	minSharePoolSize = 20
	minMissRate      = 0.4
	MaxVoteCount     = 3
	ValidVoteCount   = 2
)
