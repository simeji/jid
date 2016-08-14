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
	assert.Equal(e.json.Get("name").MustString(), "go", "they should be equal")
}

func TestParse(t *testing.T) {
	var assert = assert.New(t)

	r := bytes.NewBufferString("{\"name\":\"go\"}")

	_, e := parse(r)
	assert.True(e)

	r2 := bytes.NewBufferString("{\"name\":\"go\"")
	_, e2 := parse(r2)
	assert.False(e2)
}
