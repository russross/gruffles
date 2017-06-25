package main

import (
	"bytes"
	"fmt"
	"log"
)

const bitWordSize = 32

type BitSet struct {
	bits []uint32
}

func (set *BitSet) IsSet(i uint) bool {
	index, offset := i/bitWordSize, i%bitWordSize
	if index >= uint(len(set.bits)) {
		return false
	}
	return set.bits[index]&(1<<offset) != 0
}

func (set *BitSet) Clear(i uint) {
	index, offset := i/bitWordSize, i%bitWordSize
	if index >= uint(len(set.bits)) {
		return
	}
	set.bits[index] &^= 1 << offset
}

func (set *BitSet) Set(i uint) {
	index, offset := i/bitWordSize, i%bitWordSize
	for n := uint(len(set.bits)); n <= index; n++ {
		log.Printf("extending set to %d words for index %d", len(set.bits)+1, i)
		set.bits = append(set.bits, 0)
	}
	set.bits[index] |= 1 << offset
}

func (set *BitSet) String() string {
	buf := new(bytes.Buffer)
	for i := len(set.bits) - 1; i >= 0; i-- {
		fmt.Fprintf(buf, "%0*x", bitWordSize/4, set.bits[i])
	}
	raw := bytes.TrimLeft(buf.Bytes(), "0")
	if len(raw) == 0 {
		return "0"
	}
	return string(raw)
}
