// Copyright 2015 Viacheslav Chimishuk <vchimishuk@yandex.ru>
//
// This file is part of asp.
//
// Chub is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Chub is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Chub. If not, see <http://www.gnu.org/licenses/>.

package format

import (
	"reflect"
	"strings"
	"testing"
)

func reprSlice(nodes []node) string {
	snodes := make([]string, len(nodes))

	for i, n := range nodes {
		snodes[i] = n.repr()
	}

	return strings.Join(snodes, ", ")
}

func testParse(t *testing.T, s string, expected []node) {
	actual, err := parse(s, 10)
	if err != nil {
		t.Errorf("Error parsing string \"%s\". %s", s, err)
	} else if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Error parsing string \"%s\".\n"+
			"Expected: %s\nActual: %s",
			s, reprSlice(expected), reprSlice(actual))
	}
}

func TestParse(t *testing.T) {
	testParse(t, "",
		[]node{})
	testParse(t, " ",
		[]node{textNode(" ")})
	testParse(t, " foo ",
		[]node{textNode(" foo ")})
	testParse(t, "foo {{bar}} baz",
		[]node{textNode("foo {bar} baz")})
	testParse(t, "{%a}",
		[]node{newSubstNode(-1, false, false, "%s", []string{"a"})})
	testParse(t, "{10:%a}",
		[]node{newSubstNode(10, false, false, "%s", []string{"a"})})
	testParse(t, "{-10:%a}",
		[]node{newSubstNode(10, false, true, "%s", []string{"a"})})
	testParse(t, "{-20%:%a}",
		[]node{newSubstNode(10, true, true, "%s", []string{"a"})})
	testParse(t, "foo{30%:%a - %t}{-70%:%a}bar",
		[]node{newTextNode("foo"),
			newSubstNode(1, true, false, "%s - %s", []string{"a", "t"}),
			newSubstNode(3, true, true, "%s", []string{"a"}),
			newTextNode("bar")})
	testParse(t, "{10%:%a}{-90%:%t}",
		[]node{newSubstNode(1, true, false, "%s", []string{"a"}),
			newSubstNode(9, true, true, "%s", []string{"t"})})
	testParse(t, "{*%:%a}",
		[]node{newSubstNode(10, true, false, "%s", []string{"a"})})
	testParse(t, "{-*%:%a}",
		[]node{newSubstNode(10, true, true, "%s", []string{"a"})})
	testParse(t, "{*%:%a}{*%:%a}{*%:%a}",
		[]node{newSubstNode(3, true, false, "%s", []string{"a"}),
			newSubstNode(3, true, false, "%s", []string{"a"}),
			newSubstNode(4, true, false, "%s", []string{"a"})})
	testParse(t, "{50%:%a}{*%:%a}{*%:%a}",
		[]node{newSubstNode(5, true, false, "%s", []string{"a"}),
			newSubstNode(2, true, false, "%s", []string{"a"}),
			newSubstNode(3, true, false, "%s", []string{"a"})})
	testParse(t, "{-*%:%a}{50%:%a}{*%:%a}",
		[]node{newSubstNode(2, true, true, "%s", []string{"a"}),
			newSubstNode(5, true, false, "%s", []string{"a"}),
			newSubstNode(3, true, false, "%s", []string{"a"})})
}
