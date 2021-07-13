package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldDetectAlphaNumericString(t *testing.T) {
	assert.True(t, IsStringAlphaNumeric("abc"))
	assert.True(t, IsStringAlphaNumeric("abc123"))
	assert.False(t, IsStringAlphaNumeric("abc123@"))
}

func TestShouldSplitIntoEvenStringsOfFour(t *testing.T) {
	input := testStringInput

	arrayOfStrings := SliceString(input, 4)

	assert.Equal(t, len(arrayOfStrings), 3)
	assert.Equal(t, "abcd", arrayOfStrings[0])
	assert.Equal(t, "efgh", arrayOfStrings[1])
	assert.Equal(t, "ijkl", arrayOfStrings[2])
}

func TestShouldSplitIntoEvenStringsOfOne(t *testing.T) {
	input := testStringInput

	arrayOfStrings := SliceString(input, 1)

	assert.Equal(t, 12, len(arrayOfStrings))
	assert.Equal(t, "a", arrayOfStrings[0])
	assert.Equal(t, "b", arrayOfStrings[1])
	assert.Equal(t, "c", arrayOfStrings[2])
	assert.Equal(t, "d", arrayOfStrings[3])
	assert.Equal(t, "l", arrayOfStrings[11])
}

func TestShouldSplitIntoUnevenStringsOfFour(t *testing.T) {
	input := testStringInput + "m"

	arrayOfStrings := SliceString(input, 4)

	assert.Equal(t, len(arrayOfStrings), 4)
	assert.Equal(t, "abcd", arrayOfStrings[0])
	assert.Equal(t, "efgh", arrayOfStrings[1])
	assert.Equal(t, "ijkl", arrayOfStrings[2])
	assert.Equal(t, "m", arrayOfStrings[3])
}

func TestShouldFindSliceDifferencesDelta(t *testing.T) {
	before := []string{"abc", "onetwothree"}
	after := []string{"abc", "xyz"}

	added, removed := StringSlicesDelta(before, after)

	require.Len(t, added, 1)
	require.Len(t, removed, 1)
	assert.Equal(t, "onetwothree", removed[0])
	assert.Equal(t, "xyz", added[0])
}

func TestShouldNotFindSliceDifferencesDelta(t *testing.T) {
	before := []string{"abc", "onetwothree"}
	after := []string{"abc", "onetwothree"}

	added, removed := StringSlicesDelta(before, after)

	require.Len(t, added, 0)
	require.Len(t, removed, 0)
}

func TestShouldFindSliceDifferences(t *testing.T) {
	a := []string{"abc", "onetwothree"}
	b := []string{"abc", "xyz"}

	assert.True(t, IsStringSlicesDifferent(a, b))
	assert.True(t, IsStringSlicesDifferentFold(a, b))

	c := []string{"Abc", "xyz"}

	assert.True(t, IsStringSlicesDifferent(b, c))
	assert.False(t, IsStringSlicesDifferentFold(b, c))
}

func TestShouldNotFindSliceDifferences(t *testing.T) {
	a := []string{"abc", "onetwothree"}
	b := []string{"abc", "onetwothree"}

	assert.False(t, IsStringSlicesDifferent(a, b))
	assert.False(t, IsStringSlicesDifferentFold(a, b))
}

func TestShouldFindSliceDifferenceWhenDifferentLength(t *testing.T) {
	a := []string{"abc", "onetwothree"}
	b := []string{"abc", "onetwothree", "more"}

	assert.True(t, IsStringSlicesDifferent(a, b))
	assert.True(t, IsStringSlicesDifferentFold(a, b))
}

func TestShouldFindStringInSliceContains(t *testing.T) {
	a := "abc"
	slice := []string{"abc", "onetwothree"}

	assert.True(t, IsStringInSliceContains(a, slice))
}

func TestShouldNotFindStringInSliceContains(t *testing.T) {
	a := "xyz"
	slice := []string{"abc", "onetwothree"}

	assert.False(t, IsStringInSliceContains(a, slice))
}

func TestShouldFindStringInSliceFold(t *testing.T) {
	a := "xYz"
	b := "AbC"
	slice := []string{"XYz", "abc"}

	assert.True(t, IsStringInSliceFold(a, slice))
	assert.True(t, IsStringInSliceFold(b, slice))
}

func TestShouldNotFindStringInSliceFold(t *testing.T) {
	a := "xyZ"
	b := "ABc"
	slice := []string{"cba", "zyx"}

	assert.False(t, IsStringInSliceFold(a, slice))
	assert.False(t, IsStringInSliceFold(b, slice))
}

func TestIsStringInSliceSuffix(t *testing.T) {
	suffixes := []string{"apple", "banana"}

	assert.True(t, IsStringInSliceSuffix("apple.banana", suffixes))
	assert.True(t, IsStringInSliceSuffix("a.banana", suffixes))
	assert.True(t, IsStringInSliceSuffix("a_banana", suffixes))
	assert.True(t, IsStringInSliceSuffix("an.apple", suffixes))
	assert.False(t, IsStringInSliceSuffix("an.orange", suffixes))
	assert.False(t, IsStringInSliceSuffix("an.apple.orange", suffixes))
}
