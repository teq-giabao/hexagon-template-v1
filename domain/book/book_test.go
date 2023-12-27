package book_test

import (
	"github.com/stretchr/testify/assert"
	"hexagon/domain/book"
	"testing"
)

func TestNewBook(t *testing.T) {
	b := book.NewBook("9781804617007", "Microservices with Go")
	assert.Equal(t, b.ISBN, "9781804617007")
	assert.Equal(t, b.Name, "Microservices with Go")
}
