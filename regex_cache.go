package jid

import (
	"regexp"
	"strconv"
	"sync"
)

// reCache memoizes successful regexp compiles for patterns that are built
// from user input on every keystroke (e.g. "(?i)^" + keyword). Growth is
// bounded by the distinct patterns typed in one interactive session.
var reCache sync.Map // string -> *regexp.Regexp

// compileCached is a drop-in replacement for regexp.Compile in hot paths.
// Failed compiles are not cached, so error behavior is identical to
// regexp.Compile.
func compileCached(pattern string) (*regexp.Regexp, error) {
	if v, ok := reCache.Load(pattern); ok {
		return v.(*regexp.Regexp), nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	reCache.Store(pattern, re)
	return re, nil
}

// mustCompileCached is a drop-in replacement for regexp.MustCompile in hot
// paths: it panics on invalid patterns, matching MustCompile behavior.
func mustCompileCached(pattern string) *regexp.Regexp {
	re, err := compileCached(pattern)
	if err != nil {
		panic(`regexp: Compile(` + strconv.Quote(pattern) + `): ` + err.Error())
	}
	return re
}
