package digits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadUnexisting(t *testing.T) {
	store := NewMemoryStore()

	store.Load("dummy")

	assert.Nil(t, store.Load("dummy"))
}

func TestSaveAndLoad(t *testing.T) {
	store := NewMemoryStore()

	idt := &Identity{PhoneNumber: "+331234567890"}
	store.Save("dummy", idt)

	if idt != store.Load("dummy") {
		t.Error("loaded identity should point to the same memory address")
	}
	assert.Nil(t, store.Load("dummy2"))
}
