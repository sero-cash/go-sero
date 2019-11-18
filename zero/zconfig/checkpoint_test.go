package zconfig

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestCheckPoint(t *testing.T) {
	fmt.Println("checkpoint: ", hex.EncodeToString(CheckPoints.points[10000]))
	fmt.Println("maxNum: ", CheckPoints.MaxNum())
}
