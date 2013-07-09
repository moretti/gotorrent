package bitarray

import (
	"bytes"
)

// BitArray represents a series of bits, i.e. 10001101.
// Bits are stored in order, left to right, for example:
//
// bits:  10001101
// index: 01234567
type BitArray struct {
	bits []bool
}

func New(length int) *BitArray {
	if length < 0 {
		panic("Invalid argument: length")
	}

	bitArray := BitArray{
		bits: make([]bool, length),
	}
	return &bitArray
}

func NewFromBytes(bytes []byte, length int) *BitArray {
	if bytes == nil {
		panic("Argument null: bytes")
	}
	if length < 0 || length > len(bytes)*8 {
		panic("Invalid argument: length")
	}

	bitArray := New(length)

	for i := 0; i < len(bytes); i++ {
		base := i * 8
		byteValue := bytes[i]
		for bit := 0; bit < 8; bit++ {
			if base+bit >= length {
				break
			}
			bitArray.bits[base+bit] = byteValue&(byte(1)<<byte(7-bit)) > 0
		}
	}

	return bitArray
}

func (bitArray *BitArray) Len() int {
	return len(bitArray.bits)
}

func (bitArray *BitArray) Get(index int) bool {
	if index < 0 || index >= len(bitArray.bits) {
		panic("Out of range")
	}
	return bitArray.bits[index]
}

func (bitArray *BitArray) Set(index int, value bool) {
	if index < 0 || index >= len(bitArray.bits) {
		panic("Out of range")
	}

	bitArray.bits[index] = value
}

func (bitArray *BitArray) Cardinality() int {
	count := 0
	for i := 0; i < len(bitArray.bits); i++ {
		if bitArray.bits[i] {
			count++
		}
	}
	return count
}

func (bitArray *BitArray) SetIndices() []int {
	return bitArray.indices(func(value bool) bool {
		return value
	})
}

func (bitArray *BitArray) UnsetIndices() []int {
	return bitArray.indices(func(value bool) bool {
		return !value
	})
}

func (bitArray *BitArray) indices(predicate func(value bool) bool) []int {
	length := len(bitArray.bits)
	set := make([]int, length)
	count := 0
	for i := 0; i < len(bitArray.bits); i++ {
		if predicate(bitArray.bits[i]) {
			set[count] = i
			count++
		}
	}
	return set[:count]
}

func (bitArray *BitArray) String() string {
	var buffer bytes.Buffer

	str := map[bool]string{
		false: "0",
		true:  "1",
	}

	for i := 0; i < bitArray.Len(); i++ {
		buffer.WriteString(str[bitArray.bits[i]])
	}

	return buffer.String()
}
