package jid

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEngine(t *testing.T) {
	var assert = assert.New(t)

	f, _ := os.Create("/dev/null")
	e, err := NewEngine(f, &EngineAttribute{
		DefaultQuery: "",
		Monochrome:   false,
	})
	assert.Nil(e)
	assert.NotNil(err)

	ee := getEngine(`{"name":"go"}`, "")
	assert.NotNil(ee)
	assert.Equal("", ee.query.StringGet())
	assert.Equal(0, ee.queryCursorIdx)
}

func TestNewEngineWithQuery(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"name":"go"}`, ".nam")
	assert.Equal(".nam", e.query.StringGet())
	assert.Equal(4, e.queryCursorIdx)

	e = getEngine(`{"name":"go"}`, "nam")
	assert.Equal("", e.query.StringGet())
	assert.Equal(0, e.queryCursorIdx)

	e = getEngine(`{"name":"go"}`, ".nam..")
	assert.Equal("", e.query.StringGet())
	assert.Equal(0, e.queryCursorIdx)
}

func TestDeleteChar(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"name":"go"}`, "")
	e.query.StringSet(".name")
	e.queryCursorIdx = e.query.Length()

	e.deleteChar()
	assert.Equal(".nam", e.query.StringGet())
	assert.Equal(4, e.queryCursorIdx)
}

func TestDeleteWordBackward(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"name":"go"}`, "")
	e.query.StringSet(".name")

	e.deleteWordBackward()
	assert.Equal(".", e.query.StringGet())
	assert.Equal(1, e.queryCursorIdx)

	e.query.StringSet(".name[1]")
	e.deleteWordBackward()
	assert.Equal(".name", e.query.StringGet())
	assert.Equal(5, e.queryCursorIdx)

	e.query.StringSet(".name[")
	e.deleteWordBackward()
	assert.Equal(".name", e.query.StringGet())
	assert.Equal(5, e.queryCursorIdx)
}

func TestDeleteLineQuery(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"name":"go"}`, "")

	e.query.StringSet(".name")
	e.deleteLineQuery()
	assert.Equal("", e.query.StringGet())
	assert.Equal(0, e.queryCursorIdx)
}

func TestScrollToAbove(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"named":"go","NameTest":[1,2,3]}`, "")
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
	e := getEngine(`{"named":"go","NameTest":[1,2,3]}`, "")
	e.scrollToBelow()
	assert.Equal(1, e.contentOffset)
	e.scrollToBelow()
	e.scrollToBelow()
	assert.Equal(3, e.contentOffset)
}
func TestScrollToBottomAndTop(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"named":"go","NameTest":[1,2,3]}`, "")

	e.scrollToBottom(5)
	assert.Equal(4, e.contentOffset)

	e.scrollToTop()
	assert.Equal(0, e.contentOffset)
}

func TestGetContents(t *testing.T) {
	var assert = assert.New(t)

	e := getEngine(`{"name":"go"}`, "")
	c := e.getContents()
	assert.Equal([]string{`{`, `  "name": "go"`, "}"}, c)
	assert.Equal([]string{}, e.candidates)
	assert.Equal([]string{"", ""}, e.complete)

	e = getEngine(`{"name":"go", "naming":"simeji", "foo":"bar"}`, "")
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
	e := getEngine(`{"name":"go"}`, "")

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
	e := getEngine(`{"name":"go","NameTest":[1,2,3]}`, "")
	e.query.StringSet(".")
	e.queryConfirm = false
	e.candidates = []string{"test", "testing", "foo"}

	e.candidateidx = 0
	e.confirmCandidate()
	assert.Equal(".test", e.query.StringGet())
	assert.True(e.queryConfirm)
	assert.Equal(5, e.queryCursorIdx)

	e.candidateidx = 2
	e.confirmCandidate()
	assert.Equal(".foo", e.query.StringGet())

	assert.True(e.queryConfirm)
	assert.Equal(4, e.queryCursorIdx)

	e = getEngine(`{"name":"go"}`, "")
	e.query.StringSet(".name.hoge")
	e.candidates = []string{"aaa", "bbb", "ccc"}
	e.candidateidx = 1
	e.confirmCandidate()

	assert.True(e.queryConfirm)
	assert.Equal(9, e.queryCursorIdx)
	assert.Equal(".name.bbb", e.query.StringGet())
}

func TestCtrllAction(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"name":"go","NameTest":[1,2,3]}`, "")
	assert.False(e.keymode)
	e.toggleKeymode()
	assert.True(e.keymode)
	e.toggleKeymode()
	assert.False(e.keymode)
}

func TestTabAction(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"name":"go","NameTest":[1,2,3]}`, "")
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
	e := getEngine(`{"name":"go","NameTest":[1,2,3]}`, "")
	assert.False(e.candidatemode)
	e.escapeCandidateMode()
	assert.False(e.candidatemode)
	e.candidatemode = true
	e.escapeCandidateMode()
	assert.False(e.candidatemode)
}

func TestInputChar(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"name":"go"}`, "")
	e.query.StringSet(".name")
	e.queryCursorIdx = e.query.Length()
	assert.Equal(5, e.queryCursorIdx)

	e.inputChar('n')
	assert.Equal(".namen", e.query.StringGet())
	assert.Equal(6, e.queryCursorIdx)

	e.inputChar('.')
	assert.Equal(".namen.", e.query.StringGet())
	assert.Equal(7, e.queryCursorIdx)
}

func TestMoveCursorForwardAndBackward(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"name":"simeji"}`, "")
	e.query.StringSet(".ne")

	e.moveCursorForward()
	assert.Equal(1, e.queryCursorIdx)
	e.moveCursorForward()
	assert.Equal(2, e.queryCursorIdx)
	e.moveCursorForward()
	assert.Equal(3, e.queryCursorIdx)
	e.moveCursorForward()
	assert.Equal(3, e.queryCursorIdx)

	e.moveCursorBackward()
	assert.Equal(2, e.queryCursorIdx)
	e.moveCursorBackward()
	assert.Equal(1, e.queryCursorIdx)
	e.moveCursorBackward()
	assert.Equal(0, e.queryCursorIdx)
	e.moveCursorBackward()
	assert.Equal(0, e.queryCursorIdx)
}

func TestMoveCursorToTopAndEnd(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"name":"simeji"}`, "")
	e.query.StringSet(".ne")

	e.moveCursorToTop()
	assert.Zero(e.queryCursorIdx)

	e.moveCursorToEnd()
	assert.Equal(3, e.queryCursorIdx)
}

func getEngine(j string, qs string) *Engine {
	r := bytes.NewBufferString(j)
	e, _ := NewEngine(r, &EngineAttribute{
		DefaultQuery: qs,
		Monochrome:   false,
	})
	ee := e.(*Engine)
	return ee
}
