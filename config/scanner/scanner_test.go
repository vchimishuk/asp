// Copyright 2017 Viacheslav Chimishuk <vchimishuk@yandex.ru>
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

package scanner

import (
	"strings"
	"testing"
)

func TestKeyVal(t *testing.T) {
	r := strings.NewReader("" +
		"key1 = val1 # A comment.\n" +
		"# Comments only line.\n" +
		" \n" +
		"key2 = val2 \n")
	s := New(r)

	if !s.Scan() {
		t.Fatal(s.Err())
	}
	assert(t, "key1", s.Key())
	assert(t, "val1", s.Val())
	if !s.Scan() {
		t.Fatal(s.Err())
	}
	assert(t, "key2", s.Key())
	assert(t, "val2", s.Val())
	if s.Scan() {
		t.Fatal(s.Err())
	}
}

func TestErr(t *testing.T) {
	s := New(strings.NewReader("invalid line"))

	if s.Scan() {
		t.Fatal(s.Err())
	}
	assert(t, "invalid key=value format at line 1", s.Err().Error())
}

func assert(t *testing.T, exp, act string) {
	if exp != act {
		t.Fatalf("%s != %s", exp, act)
	}
}
