package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func DeepSerial(src interface{}) (ret bytes.Buffer) {
	if err := gob.NewEncoder(&ret).Encode(src); err != nil {
		panic(fmt.Sprintf("deepCopy encode error for : %v", src))
	}
	return
}

func DeepUnserial(buf *bytes.Buffer, dst interface{}) {
	if err := gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst); err != nil {
		panic(fmt.Sprintf("deepCopy decode error for : %v", err))
	}
}

func DeepCopy(dst, src interface{}) {
	var buf bytes.Buffer
	buf = DeepSerial(src)
	DeepUnserial(&buf, dst)
}
