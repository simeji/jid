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

func TestScrollPageUpDown(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`{"named":"go","NameTest":[1,2,3]}`, "")

	cl := len(e.getContents())
	// Then DefaultY = 1
	e.scrollPageDown(cl, 3)
	assert.Equal(2, e.contentOffset)
	e.scrollPageDown(cl, 3)
	assert.Equal(4, e.contentOffset)

	e.scrollPageUp(3)
	assert.Equal(2, e.contentOffset)

	// term height changed
	e.scrollPageDown(cl, 5)
	assert.Equal(6, e.contentOffset)

	e.scrollPageDown(cl, 5)
	assert.Equal(7, e.contentOffset)

	e.scrollPageDown(cl, 5)
	assert.Equal(7, e.contentOffset)

	e.scrollPageUp(5)
	assert.Equal(3, e.contentOffset)

	e.scrollPageUp(5)
	assert.Equal(0, e.contentOffset)

	e.scrollPageUp(5)
	assert.Equal(0, e.contentOffset)

	// term height > content size + default Y (a filter query line)
	e.scrollPageDown(cl, 10)
	assert.Equal(7, e.contentOffset)
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
	assert.Equal(".NameTest[0]", e.query.StringGet())

	_, e.complete, _, _ = e.manager.GetPretty(e.query, true)
	e.candidatemode = false
	e.tabAction()
	assert.Equal(".NameTest[1]", e.query.StringGet())
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

func TestRemoveLastJMESPathSegment(t *testing.T) {
	var assert = assert.New(t)

	assert.Equal("", removeLastJMESPathSegment(""))
	assert.Equal("to_array(@)[0]", removeLastJMESPathSegment("to_array(@)[0].id"))
	assert.Equal("to_array(@)", removeLastJMESPathSegment("to_array(@)[0]"))
	assert.Equal("", removeLastJMESPathSegment("to_array(@)"))
	assert.Equal("foo", removeLastJMESPathSegment("foo.bar"))
	assert.Equal("", removeLastJMESPathSegment("foo"))
	assert.Equal("foo", removeLastJMESPathSegment("foo[0]"))
	assert.Equal("foo.bar[1]", removeLastJMESPathSegment("foo.bar[1].baz"))
	// &field) inside a function call is one deletion unit; & is kept
	assert.Equal("max_by(@, &", removeLastJMESPathSegment("max_by(@, &base_stat)"))
	assert.Equal("sort_by(@, &", removeLastJMESPathSegment("sort_by(@, &name)"))
	assert.Equal("sort_by(@, &", removeLastJMESPathSegment("sort_by(@, &)"))
}

func TestDeleteWordBackwardJMESPath(t *testing.T) {
	var assert = assert.New(t)
	e := getEngine(`[{"id":1,"title":"hello"},{"id":2,"title":"world"}]`, "")

	// step-by-step deletion: .[1] | to_array(@)[0].id
	e.query.StringSet(".[1] | to_array(@)[0].id")
	e.queryCursorIdx = e.query.Length()

	e.deleteWordBackward()
	assert.Equal(".[1] | to_array(@)[0]", e.query.StringGet())

	e.deleteWordBackward()
	assert.Equal(".[1] | to_array(@)", e.query.StringGet())

	e.deleteWordBackward()
	assert.Equal(".[1] | ", e.query.StringGet())

	e.deleteWordBackward()
	assert.Equal(".[1]", e.query.StringGet())

	// simple pipe deletion
	e.query.StringSet(". | keys(@)")
	e.deleteWordBackward()
	assert.Equal(". | ", e.query.StringGet())

	e.deleteWordBackward()
	assert.Equal(".", e.query.StringGet())

	// &field) argument deletion: & is preserved
	e.query.StringSet(".stats | max_by(@, &base_stat)")
	e.queryCursorIdx = e.query.Length()
	e.deleteWordBackward()
	assert.Equal(".stats | max_by(@, &", e.query.StringGet())
}

func TestConfirmCandidateJMESPath(t *testing.T) {
	var assert = assert.New(t)

	// function candidate: strips partial pipe suffix, inserts "| funcname(args)"
	e := getEngine(`{"users":[{"name":"alice"}]}`, "")
	e.query.StringSet(".users | len")
	e.candidates = []string{"length("}
	e.candidateidx = 0
	e.confirmCandidate()
	assert.Equal(".users | length(@)", e.query.StringGet())
	assert.True(e.queryConfirm)

	// wildcard field candidate: appends .fieldname
	e.query.StringSet(".users[*]")
	e.candidates = []string{"age", "name"}
	e.candidateidx = 1
	e.confirmCandidate()
	assert.Equal(".users[*].name", e.query.StringGet())

	// wildcard trailing dot: no double-dot when query ends with "."
	e.query.StringSet(".users[*].")
	e.candidates = []string{"age", "name"}
	e.candidateidx = 1
	e.confirmCandidate()
	assert.Equal(".users[*].name", e.query.StringGet())

	// wildcard mid-path (e.g. after index): Contains("[*]") appends .field without PopKeyword
	e.query.StringSet(".users[*].addr[0]")
	e.candidates = []string{"city", "zip"}
	e.candidateidx = 0
	e.confirmCandidate()
	assert.Equal(".users[*].addr[0].city", e.query.StringGet())

	// post-expression object field: suffix has "(" → append to full query
	e.query.StringSet(".[0] | to_array(@)[0]")
	e.candidates = []string{"id", "name"}
	e.candidateidx = 0
	e.confirmCandidate()
	assert.Equal(".[0] | to_array(@)[0].id", e.query.StringGet())

	// post-expression with trailing dot: no double-dot
	e.query.StringSet(".[0] | to_array(@)[0].")
	e.candidates = []string{"id", "name"}
	e.candidateidx = 0
	e.confirmCandidate()
	assert.Equal(".[0] | to_array(@)[0].id", e.query.StringGet())
}

func TestTabActionJMESPath(t *testing.T) {
	var assert = assert.New(t)

	// [* → Tab → [*]
	e := getEngine(`{"items":[{"id":1},{"id":2}]}`, "")
	e.query.StringSet(".items[*")
	e.candidatemode = false
	e.tabAction()
	assert.Equal(".items[*]", e.query.StringGet())

	// single-element array: complete[1]="[0]" → Tab appends [0]
	e2 := getEngine(`{"x":42}`, "")
	e2.query.StringSet(". | to_array(@)")
	e2.complete = []string{"[0]", "[0]"}
	e2.candidatemode = false
	e2.tabAction()
	assert.Equal(". | to_array(@)[0]", e2.query.StringGet())

	// wildcard sub-projection: [*].field + Tab → appends [0] (eval fixes transparently)
	e3 := getEngine(`{"items":[{"id":1},{"id":2}]}`, "")
	e3.query.StringSet(".items[*].id")
	e3.complete = []string{"[", "["}
	e3.candidatemode = false
	e3.tabAction()
	assert.Equal(".items[*].id[0]", e3.query.StringGet())
}

// TestGetContentsCompleteSingleMatch validates the preconditions for the
// "highlight while typing" feature: when a partial query narrows candidates to
// one, complete[1] holds the full key name and candidatemode is false.
func TestGetContentsCompleteSingleMatch(t *testing.T) {
	e := getEngine(`{"name":"alice","foo":"bar"}`, "")
	e.query.StringSet(".na")
	e.queryCursorIdx = e.query.Length()
	e.getContents()
	// complete[0] = suffix to display in green ("me"), complete[1] = full key ("name")
	assert.Equal(t, "me", e.complete[0])
	assert.Equal(t, "name", e.complete[1])
	// candidatemode must be false so the "typing hint" branch activates
	e.setCandidateData()
	assert.False(t, e.candidatemode)
}

// TestGetContentsCompleteMultiMatch confirms that with multiple matches
// complete[1] is the common-prefix partial (not a real key), so
// findKeyLineInContents will return -1 and no highlighting is triggered.
func TestGetContentsCompleteMultiMatch(t *testing.T) {
	e := getEngine(`{"name":"alice","naming":"bob","foo":"bar"}`, "")
	e.query.StringSet(".na")
	e.queryCursorIdx = e.query.Length()
	e.getContents()
	// common suffix "m"; complete[1]="nam" is a partial prefix, not a real key
	assert.Equal(t, "m", e.complete[0])
	assert.Equal(t, "nam", e.complete[1])
	assert.Equal(t, []string{"name", "naming"}, e.candidates)
	// "nam" is not a real JSON key → no highlight line found
	contents := []string{`{`, `  "name": "alice",`, `  "naming": "bob"`, `}`}
	line, _ := findKeyLineInContents(contents, e.complete[1])
	assert.Equal(t, -1, line)
}

func TestFindKeyLineInContents(t *testing.T) {
	contents := []string{
		`{`,
		`  "name": "alice",`,
		`  "age": 30,`,
		`  "url": "https://example.com/name"`,
		`}`,
	}
	line, indent := findKeyLineInContents(contents, "name")
	assert.Equal(t, 1, line)
	assert.Equal(t, 2, indent)

	line, indent = findKeyLineInContents(contents, "age")
	assert.Equal(t, 2, line)
	assert.Equal(t, 2, indent)

	// "name" in a value string is not a key
	line, _ = findKeyLineInContents(contents, "missing")
	assert.Equal(t, -1, line)
	line, _ = findKeyLineInContents(contents, "alice")
	assert.Equal(t, -1, line)
}

func TestFindKeyLineInContentsEmpty(t *testing.T) {
	line, _ := findKeyLineInContents([]string{}, "name")
	assert.Equal(t, -1, line)
	line, _ = findKeyLineInContents([]string{"{", "}"}, "name")
	assert.Equal(t, -1, line)
}

func TestFindKeyLineInContentsFirst(t *testing.T) {
	contents := []string{`{"id": 1}`}
	line, indent := findKeyLineInContents(contents, "id")
	assert.Equal(t, 0, line)
	assert.Equal(t, 0, indent)
}

func TestFindKeyLineInContentsShallowWins(t *testing.T) {
	// "name" appears nested (indent 8) before the shallow occurrence (indent 2)
	contents := []string{
		`{`,
		`  "abilities": [`,
		`    {`,
		`        "name": "overgrow"`,  // indent 8, nested
		`    }`,
		`  ],`,
		`  "name": "bulbasaur"`,  // indent 2, root
		`}`,
	}
	line, indent := findKeyLineInContents(contents, "name")
	assert.Equal(t, 6, line)   // root-level line
	assert.Equal(t, 2, indent) // shallowest indent
}

func TestAmpFieldCursorPos(t *testing.T) {
	e := getEngine(`[{"name":"alice","age":30}]`, "")

	// &partial) mode: cursor should land right after &
	e.query.StringSet(". | max_by(@, &)")
	pos := e.ampFieldCursorPos(e.query.StringGet())
	// ". | max_by(@, &" has 15 runes (& at index 14), so cursor = 15
	assert.Equal(t, 15, pos)

	// with partial typed: cursor stays right after & (index 15 as well)
	e.query.StringSet(". | max_by(@, &base_stat)")
	pos = e.ampFieldCursorPos(e.query.StringGet())
	assert.Equal(t, 15, pos)

	// no & in query: fall back to query.Length()
	e.query.StringSet(". | keys(@)")
	pos = e.ampFieldCursorPos(e.query.StringGet())
	assert.Equal(t, e.query.Length(), pos)

	// no pipe: fall back to query.Length()
	e.query.StringSet(".name")
	pos = e.ampFieldCursorPos(e.query.StringGet())
	assert.Equal(t, e.query.Length(), pos)
}

// TestSetCandidateDataAmpFieldCursorPosition verifies that when &field placeholder
// text is deleted from the query, the cursor is placed at placeholderStart (between
// '&' and ')'), NOT at the end of the query string.
func TestSetCandidateDataAmpFieldCursorPosition(t *testing.T) {
	// JSON: array of objects, query has "&field)" placeholder inside a pipe expression
	e := getEngine(`[{"name":"alice","age":30}]`, "")
	// Simulate the state after sort_by(@, &field) was auto-inserted:
	//   query = ". | sort_by(@, &field)"
	//   placeholderStart = index of 'f' in "field" = 16 (". | sort_by(@, &" is 16 runes)
	//   placeholderLen   = 5 ("field")
	e.query.StringSet(". | sort_by(@, &field)")
	// placeholderStart is the rune index of the first char of "field"
	// ". | sort_by(@, &" has 17 chars (indices 0..16), so "field" starts at 17
	phStart := len([]rune(". | sort_by(@, &"))
	e.placeholderStart = phStart
	e.placeholderLen = 5 // len("field")
	e.queryCursorIdx = phStart

	// Force the candidates that setCandidateData will see (simulates what
	// getContents() would produce for this query)
	e.candidates = []string{"age", "name"}
	e.complete = []string{"", "age"}

	e.setCandidateData()

	// After deletion: query becomes ". | sort_by(@, &)" (field removed)
	assert.Equal(t, ". | sort_by(@, &)", e.query.StringGet())
	// Cursor should be at placeholderStart (between '&' and ')'), not at the end
	assert.Equal(t, phStart, e.queryCursorIdx)
	// Placeholder must be cleared
	assert.Equal(t, -1, e.placeholderStart)
	assert.Equal(t, 0, e.placeholderLen)
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
