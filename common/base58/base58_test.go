package base58

import "testing"

func TestIsBase58Str(t *testing.T) {
	tests :=[]struct{
		str string
		exp bool
	}{
		{"123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz",true},
		{"64t1MPxFp4yzxNJ64zp1NmrTXWsrLuw9DMiMZeujbD2HVAKhjR3zpKnuFVjjAXAp86G2PzSVSsdiMdwp5JPoqxtP", true},
		{"",false},
		{`""`,false},
		{"000000000",false},
		{"IIIIIII",false},
		{"OOOOOOOOOO",false},
		{"lllllll",false},
		{"0123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz",false},
		{"123456789ABCDEFGHJKLMNPOSTUVWXYZabcdefghijkmnopqrstuvwxyz",false},
	}
	for _,test := range tests{
		if result :=IsBase58Str(test.str); result != test.exp {
			t.Errorf("IsBase58Str(%s) == %v; expected %v",
				test.str, result, test.exp)
		}
	}
}