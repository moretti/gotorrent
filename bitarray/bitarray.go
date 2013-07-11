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

func NewFromString(bits string) *BitArray {
	bit := map[rune]bool{
		'0': false,
		'1': true,
	}

	bitArray := New(len(bits))
	for i, c := range bits {
		value := bit[c]
		bitArray.bits[i] = value
	}
	return bitArray
}

func (bitArray *BitArray) Len() int {
	return len(bitArray.bits)
}

func (bitArray *BitArray) Get(index int) bool {
	if index < 0 || index >= bitArray.Len() {
		panic("Out of range")
	}
	return bitArray.bits[index]
}

func (bitArray *BitArray) Set(index int, value bool) {
	if index < 0 || index >= bitArray.Len() {
		panic("Out of range")
	}

	bitArray.bits[index] = value
}

func (bitArray *BitArray) Cardinality() int {
	count := 0
	length := bitArray.Len()
	for i := 0; i < length; i++ {
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
	length := bitArray.Len()
	set := make([]int, length)
	count := 0
	for i := 0; i < length; i++ {
		if predicate(bitArray.bits[i]) {
			set[count] = i
			count++
		}
	}
	return set[:count]
}

func (bitArray *BitArray) And(other *BitArray) *BitArray {
	if other == nil {
		panic("Argument null: other")
	}

	length := bitArray.Len()
	if length != other.Len() {
		panic("Lenghts must be equal")
	}

	result := New(length)
	for i := 0; i < length; i++ {
		result.bits[i] = bitArray.bits[i] && other.bits[i]
	}
	return result
}

func (bitArray *BitArray) Xor(other *BitArray) *BitArray {
	if other == nil {
		panic("Argument null: other")
	}

	length := bitArray.Len()
	if length != other.Len() {
		panic("Lenghts must be equal")
	}

	result := New(length)
	for i := 0; i < length; i++ {
		result.bits[i] = bitArray.bits[i] != other.bits[i]
	}
	return result
}

func (bitArray *BitArray) Or(other *BitArray) *BitArray {
	if other == nil {
		panic("Argument null: other")
	}

	length := bitArray.Len()
	if length != other.Len() {
		panic("Lenghts must be equal")
	}

	result := New(length)
	for i := 0; i < length; i++ {
		result.bits[i] = bitArray.bits[i] || other.bits[i]
	}
	return result
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
