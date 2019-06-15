package ethapi

import (
	"math/big"

	"github.com/sero-cash/go-czero-import/seroparam"
)

var (
	big1                = big.NewInt(1)
	big2                = big.NewInt(2)
	big6                = big.NewInt(6)
	big9                = big.NewInt(9)
	bigMinus99          = big.NewInt(-99)
	bigOne     *big.Int = big.NewInt(1)
	big200W    *big.Int = big.NewInt(2000000)
	base                = big.NewInt(1e+17)
	big100              = big.NewInt(100)
	oneSero             = new(big.Int).Mul(big.NewInt(10), base)

	lReward = new(big.Int).Mul(big.NewInt(176), base)
	hReward = new(big.Int).Mul(big.NewInt(445), base)

	argA, _ = new(big.Int).SetString("985347985347985", 10)
	argB, _ = new(big.Int).SetString("16910256410256400000", 10)

	oriReward    = new(big.Int).Mul(big.NewInt(66773505743), big.NewInt(1000000000))
	interval     = big.NewInt(8294400)
	halveNimber  = big.NewInt(3057600)
	difficultyL1 = big.NewInt(340000000)
	difficultyL2 = big.NewInt(1700000000)
	difficultyL3 = big.NewInt(4000000000)
	difficultyL4 = big.NewInt(17000000000)
)

func accumulateRewardsV1(diff *big.Int, gasUsed uint64, gasLimit uint64) *big.Int {

	reward := new(big.Int).Mul(big.NewInt(350), base)

	difficulty := big.NewInt(1717986918)

	if diff.Cmp(difficulty) < 0 {
		ratio := new(big.Int).Div(new(big.Int).Mul(diff, big100), difficulty).Uint64()
		if ratio >= 80 {
			reward = reward.Mul(reward, big.NewInt(4)).Div(reward, big.NewInt(5))
		} else if ratio >= 60 {
			reward = reward.Mul(reward, big.NewInt(3)).Div(reward, big.NewInt(5))
		} else if ratio >= 40 {
			reward = reward.Mul(reward, big.NewInt(2)).Div(reward, big.NewInt(5))
		} else if ratio >= 20 {
			reward = reward.Mul(reward, big.NewInt(1)).Div(reward, big.NewInt(5))
		} else {
			reward = big.NewInt(0).Set(oneSero)
		}
	}

	ratio := new(big.Int).Div(new(big.Int).Mul(new(big.Int).SetUint64(gasUsed), big100), new(big.Int).SetUint64(gasLimit)).Uint64()
	if ratio >= 80 {
		reward = new(big.Int).Div(new(big.Int).Mul(reward, big6), big.NewInt(5))
	} else {
		reward = reward.Mul(reward, big.NewInt(4)).Div(reward, big.NewInt(5))
	}

	if reward.Cmp(oneSero) < 0 {
		reward = big.NewInt(0).Set(oneSero)
	}
	return reward
}

func accumulateRewardsV2(number, diff *big.Int) [3]*big.Int {
	var res [3]*big.Int
	rewardStd := new(big.Int).Set(oriReward)
	if number.Uint64() >= halveNimber.Uint64() {
		i := new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(number, halveNimber), interval), big1)
		rewardStd.Div(rewardStd, new(big.Int).Exp(big2, i, nil))
	}

	var reward *big.Int
	if diff.Cmp(difficultyL1) < 0 { //<3.4
		reward = new(big.Int).Mul(big.NewInt(10), base)
	} else if diff.Cmp(difficultyL2) < 0 { //<17
		ratio := new(big.Int).Add(new(big.Int).Mul(big.NewInt(56), base), new(big.Int).Mul(big.NewInt(16470000000), new(big.Int).Sub(diff, difficultyL1)))
		reward = new(big.Int).Div(new(big.Int).Mul(rewardStd, ratio), oriReward)
	} else if diff.Cmp(difficultyL3) < 0 { //<40
		ratio := new(big.Int).Add(new(big.Int).Mul(big.NewInt(280), base), new(big.Int).Mul(big.NewInt(2170000000), new(big.Int).Sub(diff, difficultyL2)))
		reward = new(big.Int).Div(new(big.Int).Mul(rewardStd, ratio), oriReward)
	} else if diff.Cmp(difficultyL4) < 0 { //<170
		ratio := new(big.Int).Add(new(big.Int).Mul(big.NewInt(330), base), new(big.Int).Mul(big.NewInt(2590000000), new(big.Int).Sub(diff, difficultyL3)))
		reward = new(big.Int).Div(new(big.Int).Mul(rewardStd, ratio), oriReward)
	} else {
		reward = rewardStd
	}
	res[0] = reward
	res[1] = new(big.Int).Div(rewardStd, big.NewInt(15))
	res[2] = new(big.Int).Div(new(big.Int).Mul(reward, big2), big.NewInt(15))
	return res
}

func accumulateRewardsV3(number, bdiff *big.Int) [3]*big.Int {
	var res [3]*big.Int
	diff := new(big.Int).Div(bdiff, big.NewInt(1000000000))
	reward := new(big.Int).Add(new(big.Int).Mul(argA, diff), argB)

	if reward.Cmp(lReward) < 0 {
		reward = new(big.Int).Set(lReward)
	} else if reward.Cmp(hReward) > 0 {
		reward = new(big.Int).Set(hReward)
	}

	i := new(big.Int).Add(new(big.Int).Div(new(big.Int).Sub(number, halveNimber), interval), big1)
	reward.Div(reward, new(big.Int).Exp(big2, i, nil))

	res[0] = reward
	res[1] = big.NewInt(0)
	res[2] = new(big.Int).Div(reward, big.NewInt(5))
	return res
}

/**
  [0] block reward
  [1] community reward
  [2] team reward
*/
func GetBlockReward(number, diff *big.Int, gasUsed, gasLimit uint64) [3]*big.Int {
	if number.Uint64() >= seroparam.SIP3() {
		return accumulateRewardsV3(number, diff)
	} else if number.Uint64() >= seroparam.SIP1() {
		return accumulateRewardsV2(number, diff)
	} else {
		var res [3]*big.Int
		res[0] = accumulateRewardsV1(diff, gasUsed, gasLimit)
		res[1] = big.NewInt(0)
		res[2] = big.NewInt(0)
		return res
	}
}
