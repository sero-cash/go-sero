// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ethash

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/sero-cash/go-sero/common/math"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/params"
)

type diffTest struct {
	ParentTimestamp    uint64
	ParentDifficulty   *big.Int
	CurrentTimestamp   uint64
	CurrentBlocknumber *big.Int
	CurrentDifficulty  *big.Int
}

func (d *diffTest) UnmarshalJSON(b []byte) (err error) {
	var ext struct {
		ParentTimestamp    string
		ParentDifficulty   string
		CurrentTimestamp   string
		CurrentBlocknumber string
		CurrentDifficulty  string
	}
	if err := json.Unmarshal(b, &ext); err != nil {
		return err
	}

	d.ParentTimestamp = math.MustParseUint64(ext.ParentTimestamp)
	d.ParentDifficulty = math.MustParseBig256(ext.ParentDifficulty)
	d.CurrentTimestamp = math.MustParseUint64(ext.CurrentTimestamp)
	d.CurrentBlocknumber = math.MustParseBig256(ext.CurrentBlocknumber)
	d.CurrentDifficulty = math.MustParseBig256(ext.CurrentDifficulty)

	return nil
}

func TestCalcDifficulty(t *testing.T) {
	file, err := os.Open(filepath.Join("..", "..", "tests", "testdata", "BasicTests", "difficulty.json"))
	if err != nil {
		t.Skip(err)
	}
	defer file.Close()

	tests := make(map[string]diffTest)
	err = json.NewDecoder(file).Decode(&tests)
	if err != nil {
		t.Fatal(err)
	}

	config := &params.ChainConfig{AutumnTwilightBlock: big.NewInt(0)}

	for name, test := range tests {
		number := new(big.Int).Sub(test.CurrentBlocknumber, big.NewInt(1))
		diff := CalcDifficulty(config, test.CurrentTimestamp, &types.Header{
			Number:     number,
			Time:       new(big.Int).SetUint64(test.ParentTimestamp),
			Difficulty: test.ParentDifficulty,
		})
		if diff.Cmp(test.CurrentDifficulty) != 0 {
			t.Error(name, "failed. Expected", test.CurrentDifficulty, "and calculated", diff)
		}
	}
}

func TestCalcDifficultyV2(t *testing.T) {
	print(130000, 330000000)
	print(130000, 340000000)
	print(130000, 340000001)
	print(130000, 1651524619)

	print(130001, 330000000)
	print(130001, 340000000)
	print(130001, 340000001)
	print(130001, 1699999999)
	print(130001, 1700000000)
	print(130001, 1700000001)
	print(130001, 3999999999)
	print(130001, 4000000000)
	print(130001, 4000000001)
	print(130001, 16999999999)
	print(130001, 17000000000)
	print(130001, 17000000001)

	print(3057600, 330000000)
	print(3057600, 339999999)
	print(3057600, 340000000)
	print(3057600, 340000001)
	print(3057600, 1699999999)
	print(3057600, 1700000000)
	print(3057600, 1700000001)
	print(3057600, 3999999999)
	print(3057600, 4000000000)
	print(3057600, 4000000001)
	print(3057600, 16999999999)
	print(3057600, 17000000000)
	print(3057600, 17000000001)

	print(3057601, 330000000)
	print(3057601, 339999999)
	print(3057601, 340000000)
	print(3057601, 340000001)
	print(3057601, 1699999999)
	print(3057601, 1700000000)
	print(3057601, 1700000001)
	print(3057601, 3999999999)
	print(3057601, 4000000000)
	print(3057601, 4000000001)
	print(3057601, 16999999999)
	print(3057601, 17000000000)
	print(3057601, 17000000001)

	print(3057601+8294400, 330000000)
	print(3057601+8294400, 339999999)
	print(3057601+8294400, 340000000)
	print(3057601+8294400, 340000001)
	print(3057601+8294400, 1699999999)
	print(3057601+8294400, 1700000000)
	print(3057601+8294400, 1700000001)
	print(3057601+8294400, 3999999999)
	print(3057601+8294400, 4000000000)
	print(3057601+8294400, 4000000001)
	print(3057601+8294400, 16999999999)
	print(3057601+8294400, 17000000000)
	print(3057601+8294400, 17000000001)

	print(3057601+8294400*2, 330000000)
	print(3057601+8294400*2, 339999999)
	print(3057601+8294400*2, 340000000)
	print(3057601+8294400*2, 340000001)
	print(3057601+8294400*2, 1699999999)
	print(3057601+8294400*2, 1700000000)
	print(3057601+8294400*2, 1700000001)
	print(3057601+8294400*2, 3999999999)
	print(3057601+8294400*2, 4000000000)
	print(3057601+8294400*2, 4000000001)
	print(3057601+8294400*2, 16999999999)
	print(3057601+8294400*2, 17000000000)
	print(3057601+8294400*2, 17000000001)


	print(120000, 2075498290)
}

func print(number int64, difficulty int64) {
	header := &types.Header{
		Number:     big.NewInt(number),
		Difficulty: big.NewInt(difficulty),
	}
	v2 := accumulateRewardsV2(nil, header)
	fmt.Println(number, difficulty)
	fmt.Println(v2)
	fmt.Println(new(big.Float).Quo(new(big.Float).SetInt(v2), big.NewFloat(1e+18)))
	fmt.Println("-------------------------")
}
