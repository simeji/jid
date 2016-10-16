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

	e.query.StringSet(".name[1]")
	e.ctrlwAction()
	assert.Equal(".name", e.query.StringGet())

	e.query.StringSet(".name[")
	e.ctrlwAction()
	assert.Equal(".name", e.query.StringGet())
}

func TestCtrlKAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r, false, false)
	assert.Equal(0, e.contentOffset)
	e.ctrlkAction()
	assert.Equal(0, e.contentOffset)
	e.contentOffset = 5
	e.ctrlkAction()
	assert.Equal(4, e.contentOffset)
	e.ctrlkAction()
	assert.Equal(3, e.contentOffset)
}

func TestCtrlJAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r, false, false)
	e.ctrljAction()
	assert.Equal(1, e.contentOffset)
	e.ctrljAction()
	e.ctrljAction()
	assert.Equal(3, e.contentOffset)
}

func TestCtrllAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r, false, false)
	assert.False(e.keymode)
	e.ctrllAction()
	assert.True(e.keymode)
	e.ctrllAction()
	assert.False(e.keymode)
}

func TestTabAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r, false, false)
	e.query.StringSet(".namet")
	e.complete = []string{"est", "NameTest"}

	e.candidatemode = false
	e.tabAction()
	assert.Equal(".NameTest", e.query.StringGet())

	_, e.complete, _, _ = e.manager.GetPretty(e.query, true)
	e.candidatemode = false
	e.tabAction()
	assert.Equal(".NameTest[", e.query.StringGet())

	_, e.complete, _, _ = e.manager.GetPretty(e.query, true)
	e.candidatemode = false
	e.tabAction()
	assert.Equal(".NameTest[", e.query.StringGet())
}

func TestEscAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r, false, false)
	assert.False(e.candidatemode)
	e.escAction()
	assert.False(e.candidatemode)
	e.candidatemode = true
	e.escAction()
	assert.False(e.candidatemode)
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
