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

func TestDeleteChar(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r, false, false)
	e.query.StringSet(".name")
	e.cursorOffsetX = 5

	e.deleteChar()
	assert.Equal(".nam", e.query.StringGet())
	assert.Equal(4, e.cursorOffsetX)
}

func TestDeleteWordBackward(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r, false, false)
	e.query.StringSet(".name")

	e.deleteWordBackward()
	assert.Equal(".", e.query.StringGet())
	assert.Equal(1, e.cursorOffsetX)

	e.query.StringSet(".name[1]")
	e.deleteWordBackward()
	assert.Equal(".name", e.query.StringGet())
	assert.Equal(5, e.cursorOffsetX)

	e.query.StringSet(".name[")
	e.deleteWordBackward()
	assert.Equal(".name", e.query.StringGet())
	assert.Equal(5, e.cursorOffsetX)
}

func Test(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r, false, false)
	assert.Equal(0, e.contentOffset)
	e.scrollToAbove()
	assert.Equal(0, e.contentOffset)
	e.contentOffset = 5
	e.scrollToAbove()
	assert.Equal(4, e.contentOffset)
	e.scrollToAbove()
	assert.Equal(3, e.contentOffset)
}

func TestScrollToBelow(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r, false, false)
	e.scrollToBelow()
	assert.Equal(1, e.contentOffset)
	e.scrollToBelow()
	e.scrollToBelow()
	assert.Equal(3, e.contentOffset)
}

func TestCtrllAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r, false, false)
	assert.False(e.keymode)
	e.toggleKeymode()
	assert.True(e.keymode)
	e.toggleKeymode()
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
	e.escapeCandidateMode()
	assert.False(e.candidatemode)
	e.candidatemode = true
	e.escapeCandidateMode()
	assert.False(e.candidatemode)
}

func TestConfirmCandidate(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r, false, false)
	e.query.StringSet(".name.hoge")
	e.candidateidx = 1
	e.confirmCandidate([]string{"aaa", "bbb", "ccc"})

	assert.True(e.queryConfirm)
	assert.Equal(9, e.cursorOffsetX)
	assert.Equal(".name.bbb", e.query.StringGet())

}
func TestInputChar(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r, false, false)
	e.query.StringSet(".name")
	e.cursorOffsetX = len(e.query.Get())
	assert.Equal(5, e.cursorOffsetX)

	e.inputChar('n')
	assert.Equal(".namen", e.query.StringGet())
	assert.Equal(6, e.cursorOffsetX)

	e.inputChar('.')
	assert.Equal(".namen.", e.query.StringGet())
	assert.Equal(7, e.cursorOffsetX)
}
