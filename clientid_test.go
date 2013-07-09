package gotorrent

import (
	"fmt"
	"testing"
)

func TestLenConstructor(t *testing.T) {
	// Testing the length
	{
		clientId := NewClientId()
		expected := ClientIdLength
		value := len(clientId)

		if value != expected {
			t.Errorf("len(clientId1) == %v, want %v", value, expected)
		}
	}

	// Make sure they are unique
	{
		expected := 10
		clients := make(map[ClientId]bool)

		for i := 0; i < expected; i++ {
			clientId := NewClientId()
			clients[clientId] = true
		}

		value := len(clients)

		if value != expected {
			t.Errorf("len(clients) == %v, want %v", value, expected)
		}

		fmt.Println(clients)
	}
}
