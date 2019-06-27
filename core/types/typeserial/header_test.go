package typeserial

import (
	"testing"
)

type VersionPtr struct {
	VP Version
}

func TestRLP(t *testing.T) {
	/*buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)
	test := VersionPtr{nil}
	test_ptr := &test
	test_ptr = nil
	e := rlp.Encode(w, test_ptr)
	e = rlp.Encode(w, test)
	fmt.Println(e)
	w.Flush()

	dtest := Version{1}
	stream := rlp.NewStream(&buf, uint64(buf.Len()))
	_, size, _ := stream.Kind()
	fmt.Println(size)
	e = stream.Decode(&dtest)
	fmt.Println(e)
	e = stream.Decode(&dtest)

	fmt.Println(dtest)
	*/
}
