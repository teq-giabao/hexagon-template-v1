package hashing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBcryptHasher_HashAndCompare(t *testing.T) {
	hasher := NewBcryptHasher()

	hash, err := hasher.Hash("Password123!")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = hasher.Compare(hash, "Password123!")
	assert.NoError(t, err)

	err = hasher.Compare(hash, "wrong")
	assert.Error(t, err)
}
