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
}

func TestSpaceAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r, false, false)
	e.query.StringSet(".name")

	e.spaceAction()
	assert.Equal(".name ", e.query.StringGet())
}

func TestBackSpaceAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r, false, false)
	e.query.StringSet(".name")

	e.backspaceAction()
	assert.Equal(".nam", e.query.StringGet())
}

func TestCtrlwAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r, false, false)
	e.query.StringSet(".name")

	e.ctrlwAction()
	assert.Equal(".", e.query.StringGet())
}

func TestTabAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r, false, false)
	e.query.StringSet(".namet")
	e.complete = []string{"est", "NameTest"}

	e.tabAction()
	assert.Equal(".NameTest", e.query.StringGet())

	_, e.complete, _ = e.manager.GetPretty(e.query)
	e.tabAction()
	assert.Equal(".NameTest[", e.query.StringGet())
}

func TestInputAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r, false, false)
	e.query.StringSet(".name")

	e.inputAction('n')
	assert.Equal(".namen", e.query.StringGet())

	e.inputAction('.')
	assert.Equal(".namen.", e.query.StringGet())
}
