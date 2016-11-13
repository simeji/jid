package jid

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewEngine(t *testing.T) {
	var assert = assert.New(t)

	f, _ := os.Create("/dev/null")
	e := NewEngine(f)
	assert.Equal(e, &Engine{}, "they should be equal")

	r := bytes.NewBufferString("{\"name\":\"go\"}")
	e = NewEngine(r).(*Engine)
	assert.NotEqual(e, &Engine{}, "they should be not equal")
}

func TestDeleteChar(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r).(*Engine)
	e.query.StringSet(".name")
	e.cursorOffsetX = 5

	e.deleteChar()
	assert.Equal(".nam", e.query.StringGet())
	assert.Equal(4, e.cursorOffsetX)
}

func TestDeleteWordBackward(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r).(*Engine)
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

func TestScrollToAbove(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r).(*Engine)
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
	e := NewEngine(r).(*Engine)
	e.scrollToBelow()
	assert.Equal(1, e.contentOffset)
	e.scrollToBelow()
	e.scrollToBelow()
	assert.Equal(3, e.contentOffset)
}

func TestGetContents(t *testing.T) {
	var assert = assert.New(t)

	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r).(*Engine)
	c := e.getContents()
	assert.Equal([]string{`{`, `  "name": "go"`, "}"}, c)
	assert.Equal([]string{}, e.candidates)
	assert.Equal([]string{"", ""}, e.complete)

	r = bytes.NewBufferString(`{"name":"go", "naming":"simeji", "foo":"bar"}`)
	e = NewEngine(r).(*Engine)
	e.query.StringSet(".n")
	c = e.getContents()
	assert.Equal([]string{`{`, `  "foo": "bar",`, `  "name": "go",`, `  "naming": "simeji"`, "}"}, c)
	assert.Equal([]string{"name", "naming"}, e.candidates)
	assert.Equal([]string{"am", "nam"}, e.complete)

	e.keymode = true
	c = e.getContents()
	assert.Equal([]string{"name", "naming"}, c)
	assert.Equal([]string{"name", "naming"}, e.candidates)
	assert.Equal([]string{"am", "nam"}, e.complete)
}

func TestSetCandidateData(t *testing.T) {
	var assert = assert.New(t)
	e := NewEngine(bytes.NewBufferString(`{"name":"go"`)).(*Engine)

	// case 1
	e.candidates = []string{"test", "testing"}
	e.complete = []string{"est", "test"}
	e.candidatemode = true
	e.candidateidx = 1

	e.setCandidateData()
	assert.False(e.candidatemode)
	assert.Zero(e.candidateidx)
	assert.Equal([]string{}, e.candidates)

	// case 2
	e.candidates = []string{"test"}
	e.complete = []string{"", "test"}
	e.candidatemode = true
	e.candidateidx = 1

	e.setCandidateData()
	assert.False(e.candidatemode)
	assert.Zero(e.candidateidx)
	assert.Equal([]string{}, e.candidates)

	// case 3
	e.candidates = []string{"test", "testing"}
	e.complete = []string{"", "test"}
	e.candidatemode = true
	e.candidateidx = 2

	e.setCandidateData()
	assert.True(e.candidatemode)
	assert.Zero(e.candidateidx)
	assert.Equal([]string{"test", "testing"}, e.candidates)

	// case 4
	e.candidates = []string{"test", "testing"}
	e.complete = []string{"", "test"}
	e.candidatemode = true
	e.candidateidx = 1

	e.setCandidateData()
	assert.True(e.candidatemode)
	assert.Equal(1, e.candidateidx)
	assert.Equal([]string{"test", "testing"}, e.candidates)

	// case 4
	e.candidates = []string{"test", "testing"}
	e.complete = []string{"", "test"}
	e.candidatemode = false
	e.candidateidx = 1

	e.setCandidateData()
	assert.False(e.candidatemode)
	assert.Equal(0, e.candidateidx)
	assert.Equal([]string{}, e.candidates)

}

func TestConfirmCandidate(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r).(*Engine)
	e.query.StringSet(".")
	e.queryConfirm = false
	e.candidates = []string{"test", "testing", "foo"}

	e.candidateidx = 0
	e.confirmCandidate()
	assert.Equal(".test", e.query.StringGet())
	assert.True(e.queryConfirm)
	assert.Equal(5, e.cursorOffsetX)

	e.candidateidx = 2
	e.confirmCandidate()
	assert.Equal(".foo", e.query.StringGet())

	assert.True(e.queryConfirm)
	assert.Equal(4, e.cursorOffsetX)

	r = bytes.NewBufferString(`{"name":"go"}`)
	e = NewEngine(r).(*Engine)
	e.query.StringSet(".name.hoge")
	e.candidates = []string{"aaa", "bbb", "ccc"}
	e.candidateidx = 1
	e.confirmCandidate()

	assert.True(e.queryConfirm)
	assert.Equal(9, e.cursorOffsetX)
	assert.Equal(".name.bbb", e.query.StringGet())
}

func TestCtrllAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r).(*Engine)
	assert.False(e.keymode)
	e.toggleKeymode()
	assert.True(e.keymode)
	e.toggleKeymode()
	assert.False(e.keymode)
}

func TestTabAction(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","NameTest":[1,2,3]}`)
	e := NewEngine(r).(*Engine)
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
	e := NewEngine(r).(*Engine)
	assert.False(e.candidatemode)
	e.escapeCandidateMode()
	assert.False(e.candidatemode)
	e.candidatemode = true
	e.escapeCandidateMode()
	assert.False(e.candidatemode)
}

func TestInputChar(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	e := NewEngine(r).(*Engine)
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

func TestMoveCursorForwardAndBackward(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"name":"simeji"}`)
	e.query.StringSet(".ne")

	e.cursorOffsetX = 0
	e.moveCursorForward()
	assert.Equal(1, e.cursorOffsetX)
	e.moveCursorForward()
	assert.Equal(2, e.cursorOffsetX)
	e.moveCursorForward()
	assert.Equal(3, e.cursorOffsetX)
	e.moveCursorForward()
	assert.Equal(3, e.cursorOffsetX)

	e.moveCursorBackward()
	assert.Equal(2, e.cursorOffsetX)
	e.moveCursorBackward()
	assert.Equal(1, e.cursorOffsetX)
	e.moveCursorBackward()
	assert.Equal(0, e.cursorOffsetX)
	e.moveCursorBackward()
	assert.Equal(0, e.cursorOffsetX)
}

func TestMoveCursorToTopAndEnd(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"name":"simeji"}`)
	e.query.StringSet(".ne")

	e.cursorOffsetX = 2
	e.moveCursorToTop()
	assert.Zero(e.cursorOffsetX)

	e.moveCursorToEnd()
	assert.Equal(3, e.cursorOffsetX)
}

func getEngine(j string) *Engine {
	r := bytes.NewBufferString(j)
	e := NewEngine(r).(*Engine)
	return e
}
