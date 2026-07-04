package jid

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// genLargeJSON returns a deterministic JSON document with n array elements
// (~n * 200 bytes). Shape mirrors real jid usage: a top-level object with a
// large "users" array of nested objects, scalar fields, and a key containing
// a dot to exercise the escaped-key path.
func genLargeJSON(n int) []byte {
	var sb strings.Builder
	sb.Grow(n * 200)
	sb.WriteString(`{"a.b":"dotkey","meta":{"count":`)
	fmt.Fprintf(&sb, "%d", n)
	sb.WriteString(`,"source":"bench"},"users":[`)
	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb,
			`{"id":%d,"name":"user%d","age":%d,"active":%t,"tags":["alpha","beta"],"address":{"city":"city%d","zip":"%07d","country":"JP"}}`,
			i, i, 20+i%50, i%2 == 0, i, i)
	}
	sb.WriteString(`]}`)
	return []byte(sb.String())
}

func benchJsonManager(b *testing.B, n int) *JsonManager {
	b.Helper()
	jm, err := NewJsonManager(bytes.NewReader(genLargeJSON(n)))
	if err != nil {
		b.Fatal(err)
	}
	return jm
}

// benchPrettyRows returns the pretty-printed root document as display rows.
func benchPrettyRows(b *testing.B, n int) []string {
	b.Helper()
	jm := benchJsonManager(b, n)
	s, _, _, err := jm.GetPretty(NewQueryWithString("."), false)
	if err != nil {
		b.Fatal(err)
	}
	return strings.Split(s, "\n")
}

func BenchmarkGetPretty(b *testing.B) {
	for _, n := range []int{100, 5000} {
		for _, qs := range []string{".", ".users[0].address"} {
			b.Run(fmt.Sprintf("n=%d/q=%s", n, qs), func(b *testing.B) {
				jm := benchJsonManager(b, n)
				q := NewQueryWithString(qs)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if _, _, _, err := jm.GetPretty(q, false); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkGetFilteredDataLegacy(b *testing.B) {
	for _, n := range []int{100, 5000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			jm := benchJsonManager(b, n)
			q := NewQueryWithString(".users[0].na")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				jm.GetFilteredData(q, false)
			}
		})
	}
}

func BenchmarkGetFilteredDataJMESPath(b *testing.B) {
	for _, n := range []int{100, 5000} {
		for _, qs := range []string{".users[*].name", ". | keys(@)"} {
			b.Run(fmt.Sprintf("n=%d/q=%s", n, qs), func(b *testing.B) {
				jm := benchJsonManager(b, n)
				q := NewQueryWithString(qs)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					jm.GetFilteredData(q, false)
				}
			})
		}
	}
}

func BenchmarkEvalJMESPath(b *testing.B) {
	for _, n := range []int{100, 5000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			jm := benchJsonManager(b, n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := jm.evalJMESPath("users[*].name"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkValidate(b *testing.B) {
	r := []rune(".users[3].address.city")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !validate(r) {
			b.Fatal("expected valid query")
		}
	}
}

func BenchmarkGetKeywords(b *testing.B) {
	q := NewQueryWithString(".users[3].address.ci")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if kw := q.GetKeywords(); len(kw) == 0 {
			b.Fatal("expected keywords")
		}
	}
}

// benchWideObject returns a JsonManager whose root object has many keys,
// to exercise suggestion candidate generation.
func benchWideObjectManager(b *testing.B) *JsonManager {
	b.Helper()
	var sb strings.Builder
	sb.WriteByte('{')
	for i := 0; i < 50; i++ {
		if i > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, `"key%02d":%d`, i, i)
	}
	sb.WriteByte('}')
	jm, err := NewJsonManager(strings.NewReader(sb.String()))
	if err != nil {
		b.Fatal(err)
	}
	return jm
}

func BenchmarkSuggestionGet(b *testing.B) {
	jm := benchWideObjectManager(b)
	s := NewSuggestion()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Get(jm.origin, "key1")
	}
}

func BenchmarkGetCandidateKeys(b *testing.B) {
	jm := benchWideObjectManager(b)
	s := NewSuggestion()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if c := s.GetCandidateKeys(jm.origin, "key"); len(c) == 0 {
			b.Fatal("expected candidates")
		}
	}
}

func BenchmarkGetCurrentKeys(b *testing.B) {
	jm := benchWideObjectManager(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if k := getCurrentKeys(jm.origin); len(k) == 0 {
			b.Fatal("expected keys")
		}
	}
}

func BenchmarkRowsToCellsColor(b *testing.B) {
	rows := benchPrettyRows(b, 5000)
	term := NewTerminal(FilterPrompt, DefaultY, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := term.rowsToCells(rows, ""); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRowsToCellsMono(b *testing.B) {
	rows := benchPrettyRows(b, 5000)
	term := NewTerminal(FilterPrompt, DefaultY, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := term.rowsToCells(rows, ""); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHighlightCandidateKey(b *testing.B) {
	cells := makeCells(`      "name": "user42",`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		highlightCandidateKey(cells, "name", 6)
	}
}

func BenchmarkFindKeyLineInContents(b *testing.B) {
	rows := benchPrettyRows(b, 5000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if line, _ := findKeyLineInContents(rows, "name"); line < 0 {
			b.Fatal("expected key to be found")
		}
	}
}

// BenchmarkEngineGetContents measures the per-frame cost of the engine's
// content pipeline (filter + pretty-print + split) without termbox.
func BenchmarkEngineGetContents(b *testing.B) {
	for _, n := range []int{100, 5000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			jm := benchJsonManager(b, n)
			e := &Engine{
				manager:    jm,
				query:      NewQueryWithString(".users[0]"),
				complete:   []string{"", ""},
				candidates: []string{},
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if c := e.getContents(); len(c) == 0 {
					b.Fatal("expected contents")
				}
			}
		})
	}
}
