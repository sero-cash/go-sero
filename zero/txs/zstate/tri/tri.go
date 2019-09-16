// copyright 2018 The sero.cash Authors
// This file is part of the go-sero library.
//
// The go-sero library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sero library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sero library. If not, see <http://www.gnu.org/licenses/>.

package tri

import (
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/serodb"
)

type Tri interface {
	TryGet(key []byte) ([]byte, error)
	TryUpdate(key, value []byte) error
	SetState(obj *c_type.PKr, key *c_type.Uint256, value *c_type.Uint256)
	GetState(obj *c_type.PKr, key *c_type.Uint256) (ret c_type.Uint256)
	GlobalGetter() serodb.Getter
}

func slice2Uint256(s []byte) (r c_type.Uint256) {
	copy(r[:], s)
	return
}

type KEY_NAME string

func (name KEY_NAME) Bytes() []byte {
	return []byte(name)
}

func TryGetUint256s(tri Tri, key []byte, cb func([]byte, *c_type.Uint256)) (hashes []c_type.Uint256) {
	if v, err := tri.TryGet(key); err != nil {
		panic(err)
		return
	} else {
		if len(v) > 0 {
			for i := 0; i < len(v); i += 32 {
				b := slice2Uint256(v[i : i+32])
				hashes = append(hashes, b)
				if cb != nil {
					if o, err := tri.TryGet(b[:]); err != nil {
						panic(err)
						return
					} else {
						if len(o) > 0 {
							cb(o, &b)
						}
					}
				} else {
				}
			}
		}
	}
	return
}

func TryUpdateUint256s(tri Tri, key []byte, hashes []c_type.Uint256) {
	outs := []byte{}
	for _, v := range hashes {
		outs = append(outs, v[:]...)
	}
	if len(outs) > 0 {
		if err := tri.TryUpdate(key, outs); err != nil {
			panic(err)
			return
		}
	}
	return
}

type unserial interface {
	Unserial(v []byte) error
}

func GetObj(tri Tri, key []byte, obj unserial) {
	if v, err := tri.TryGet(key); err != nil {
		panic(err)
		return
	} else {
		if err := obj.Unserial(v); err != nil {
			panic(err)
			return
		} else {
		}
	}
	return
}

/*func GetGlobalObj(tri Tri, key []byte, obj unserial) {
	if v, err := tri.TryGlobalGet(key); err != nil {
		if err := obj.Unserial(nil); err != nil {
			panic(err)
			return
		} else {
		}
	} else {
		if err := obj.Unserial(v); err != nil {
			panic(err)
			return
		} else {
		}
	}
	return
}*/

func GetDBObj(db serodb.Getter, key []byte, obj unserial) {
	if v, err := db.Get(key); err != nil {
		if err := obj.Unserial(nil); err != nil {
			panic(err)
			return
		} else {
		}
	} else {
		if err := obj.Unserial(v); err != nil {
			panic(err)
			return
		} else {
		}
	}
	return
}

type serial interface {
	Serial() ([]byte, error)
}

func UpdateObj(tri Tri, key []byte, obj serial) {
	if s, err := obj.Serial(); err != nil {
		panic(err)
		return
	} else {
		if err := tri.TryUpdate(key, s); err != nil {
			panic(err)
			return
		} else {
		}
	}
	return
}

/*func UpdateGlobalObj(tri Tri, key []byte, obj serial) {
	if s, err := obj.Serial(); err != nil {
		panic(err)
		return
	} else {
		if err := tri.TryGlobalPut(key, s); err != nil {
			panic(err)
			return
		} else {
		}
	}
	return
}*/

func UpdateDBObj(database serodb.Putter, key []byte, obj serial) {
	if s, err := obj.Serial(); err != nil {
		panic(err)
		return
	} else {
		if err := database.Put(key, s); err != nil {
			panic(err)
			return
		} else {
		}
	}
	return
}
