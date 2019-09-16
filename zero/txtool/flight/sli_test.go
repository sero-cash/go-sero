package flight

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sero-cash/go-sero/zero/txtool"
)

func TestDecOut(t *testing.T) {
	outs_str := `[{"Root":"0xf3ead30acaf8576362413db319f2abf543c5db555dc0f91f7dd7d0bae28c3425","State":{"OS":{"index":1983489,"Out_O":{"Addr":"","Asset":{"Tkn":{"Currency":"0x0000000000000000000000000000000000000000000000000000000000000000","Value":0}}},"Out_Z":{"PKr":"0xeb4a555d1132357bbeb184132e163f9bd7cde4790231e802d783005f0b43b8a20e87f2e41c1e0221e3374c341331ec51c3acda0195da6c130a 166a3d0676b38fa5ea369d5cc177ff9d1d8d6e97ae117758de391e0426986370edbcecfe4f4322"},"Out_CM":null,"RootCM":"0x0d48e240051ac932 46aa7e46c024945a2493c02b19982d4bd9fddbf51ff1b611"},"TxHash":"0x75be54e70f74cd5c7ba799c73db80569634274e9f1f599411c8c62de79b6 3a09","Num":1538009}}]`
	var outs []txtool.Out
	outs = append(outs, txtool.Out{})
	bs, err := json.Marshal(outs)
	err := json.Unmarshal([]byte(outs_str), &outs)
	fmt.Println(err)
}
