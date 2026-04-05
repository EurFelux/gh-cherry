package jqutil_test

import (
	"bytes"
	"testing"

	"github.com/EurFelux/gh-cherry/internal/jqutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sample struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestEncode_Passthrough(t *testing.T) {
	var buf bytes.Buffer
	err := jqutil.Encode(&buf, sample{Name: "a", Count: 1}, "")
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"a","count":1}`, buf.String())
}

func TestEncode_FieldExtraction(t *testing.T) {
	var buf bytes.Buffer
	err := jqutil.Encode(&buf, sample{Name: "hello", Count: 42}, ".name")
	require.NoError(t, err)
	assert.Equal(t, "\"hello\"\n", buf.String())
}

func TestEncode_ArrayIteration(t *testing.T) {
	input := []sample{
		{Name: "a", Count: 1},
		{Name: "b", Count: 2},
	}
	var buf bytes.Buffer
	err := jqutil.Encode(&buf, input, ".[].name")
	require.NoError(t, err)
	assert.Equal(t, "\"a\"\n\"b\"\n", buf.String())
}

func TestEncode_InvalidExpression(t *testing.T) {
	var buf bytes.Buffer
	err := jqutil.Encode(&buf, sample{}, ".[invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid jq expression")
}

func TestEncode_NullResult(t *testing.T) {
	var buf bytes.Buffer
	err := jqutil.Encode(&buf, sample{Name: "a"}, ".missing")
	require.NoError(t, err)
	assert.Equal(t, "null\n", buf.String())
}
