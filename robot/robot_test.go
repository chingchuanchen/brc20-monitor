package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMint(t *testing.T) {
	err := mint()
	assert.Nil(t, err)
}
