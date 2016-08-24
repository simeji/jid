package jig

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewEngine(t *testing.T) {
	var assert = assert.New(t)

	f, _ := os.Create("/dev/null")
	e := NewEngine(f, false, false)
	assert.Equal(e, &Engine{}, "they should be equal")

	r := bytes.NewBufferString("{\"name\":\"go\"}")
	e = NewEngine(r, false, false)
	assert.NotEqual(e, &Engine{}, "they should be not equal")
	//assert.Equal(e.json.Get("name").MustString(), "go", "they should be equal")
}
