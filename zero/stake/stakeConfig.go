package stake

import (
	"github.com/sero-cash/go-czero-import/seroparam"
	"math/big"
)

var (
	basePrice = big.NewInt(2000000000000000000)
	addition  = big.NewInt(368891382302157)

	baseReware = big.NewInt(2330000000000000000)
	rewareStep = big.NewInt(11022927689594)

	maxPrice = big.NewInt(5930000000000000000)

	//outOfDateWindow      = uint64(544320)
	//missVotedWindow      = uint64(725760)
	//payWindow            = uint64(42336)
	//statisticsMissWindow = uint64(6048)

	//test
	outOfDateWindow      = uint64(200)
	missVotedWindow      = uint64(300)
	payWindow            = uint64(5)
	statisticsMissWindow = uint64(1000)
)


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
	SOLO_RATE  = 2
	TOTAL_RATE = 3
	minSharePoolSize =200
	minMissRate      =0.4
	MaxVoteCount = 3
	ValidVoteCount = 2
)


