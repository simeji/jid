package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/simeji/jid"
	"github.com/stretchr/testify/assert"
)

var called int = 0

func TestMain(m *testing.M) {
	called = 0
	code := m.Run()
	defer os.Exit(code)
}

func TestJidRun(t *testing.T) {
	var assert = assert.New(t)

	e := &EngineMock{err: nil}
	result := run(e, false)
	assert.Zero(result)
	assert.Equal(2, called)

	result = run(e, true)
	assert.Equal(1, called)

	result = run(e, false)
	assert.Zero(result)
}

func TestJidRunWithError(t *testing.T) {
	called = 0
	var assert = assert.New(t)
	e := &EngineMock{err: fmt.Errorf("")}
	result := run(e, false)
	assert.Equal(2, result)
	assert.Equal(0, called)
}

type EngineMock struct{ err error }

func (e *EngineMock) Run() jid.EngineResultInterface {
	return &EngineResultMock{err: e.err}
}
func (e *EngineMock) GetQuery() jid.QueryInterface {
	return jid.NewQuery([]rune(""))
}

type EngineResultMock struct{ err error }

func (e *EngineResultMock) GetQueryString() string {
	called = 1
	return ".querystring"
}
func (e *EngineResultMock) GetContent() string {
	called = 2
	return `{"test":"result"}`
}
func (e *EngineResultMock) GetError() error {
	return e.err
}
