package bitarray

import (
	"testing"
)

func TestLenConstructor(t *testing.T) {
	ba := New(5)

	{
		expected := 5
		value := ba.Len()
		if value != expected {
			t.Errorf("ba.Len() == %v, want %v", value, expected)
		}
	}

	// A new bitarray should all be set to false
	{
		expected := false

		for i := 0; i < ba.Len(); i++ {
			value := ba.Get(0)
			if value != expected {
				t.Errorf("ba.Get(%v) == %v, want %v", i, value, expected)
			}
		}
	}
}

func TestGetSet(t *testing.T) {
	ba := New(25)

	{
		indexes := []int{3, 2, 7, 9, 11}
		expected := true

		for i := range indexes {
			ba.Set(i, expected)
			value := ba.Get(i)
			if value != expected {
				t.Errorf("ba.Get(%v) == %v, want %v", i, value, expected)
			}
		}
	}
}

func TestByteConstructor(t *testing.T) {
	bytes := []byte{
		0xaa, // 10101010
		0x55, // 01010101
		0xaa, // 10101010
		0x55, // 01010101
		0x80, // 10000000
	}

	ba := NewFromBytes(bytes, len(bytes)*8)

	{
		expected := len(bytes) * 8
		value := ba.Len()
		if value != expected {
			t.Errorf("ba.Len() == %v, want %v", value, expected)
		}
	}

	{
		values := map[int]bool{
			0:  true,
			1:  false,
			6:  true,
			7:  false,
			14: false,
			15: true,
			32: true,
			39: false,
		}

		for i, expected := range values {
			value := ba.Get(i)
			if value != expected {
				t.Errorf("ba.Get(%v) == %v, want %v", i, value, expected)
			}
		}
	}
}

func TestByteConstructorPartialBytes(t *testing.T) {
	bytes := []byte{
		0xaa, // 10101010
		0x55, // 01010101
		0x00, // 00000000
		0x00, // 00000000
		0x00, // 00000000
		0x00, // 00000000
		0x00, // 00000000
		0x00, // 00000000
	}

	ba := NewFromBytes(bytes, 12)

	{
		expected := 12
		value := ba.Len()
		if value != expected {
			t.Errorf("ba.Len() == %v, want %v", value, expected)
		}
	}

	{
		values := map[int]bool{
			0:  true,
			1:  false,
			6:  true,
			7:  false,
			11: true,
		}

		for i, expected := range values {
			value := ba.Get(i)
			if value != expected {
				t.Errorf("ba.Get(%v) == %v, want %v", i, value, expected)
			}
		}
	}
}

func TestCardinality(t *testing.T) {
	ba := NewFromBytes([]byte{0xaa, 0x55}, 16)
	expected := 8
	value := ba.Cardinality()

	if value != expected {
		t.Errorf("ba.Cardinality() == %v, want %v", value, expected)
	}
}

func TestSetIndices(t *testing.T) {
	ba := NewFromBytes([]byte{0xaa, 0x55}, 16)
	expected := []int{0, 2, 4, 6, 9, 11, 13, 15}
	value := ba.SetIndices()

	if !Equal(value, expected) {
		t.Errorf("ba.SetIndices() == %v, want %v", value, expected)
	}
}

func TestUnsetIndices(t *testing.T) {
	ba := NewFromBytes([]byte{0xaa, 0x55}, 16)
	expected := []int{1, 3, 5, 7, 8, 10, 12, 14}
	value := ba.UnsetIndices()

	if !Equal(value, expected) {
		t.Errorf("ba.UnsetIndices() == %v, want %v", value, expected)
	}
}

func Equal(a, b []int) bool {
	if a == nil || b == nil {
		panic("Argument nil")
	}

	if len(a) != len(b) {
		return false
	}

	for i, c := range a {
		if c != b[i] {
			return false
		}
	}

	return true
}

func TestString(t *testing.T) {
	ba := New(10)
	for i := 0; i < ba.Len(); i++ {
		if i%2 == 0 {
			ba.Set(i, true)
		}
	}

	expected := "1010101010"
	value := ba.String()
	if value != expected {
		t.Errorf("ba.String() == %v, want %v", value, expected)
	}
}
