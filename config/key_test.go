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

package config

import (
	"strings"
	"testing"
)

func TestParseKeys(t *testing.T) {
	r := strings.NewReader("" +
		"playlist.enter = 2\n" +
		"vfs.enter = 3, 4\n")
	keys, err := ParseKeys(r)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp([]int{1, 2}, keys[SectionGlobal][CommandQuit]) {
		t.Fatal()
	}
	if !cmp([]int{2}, keys[SectionPlaylist][CommandEnter]) {
		t.Fatal()
	}
	if !cmp([]int{3, 4}, keys[SectionVFS][CommandEnter]) {
		t.Fatal()
	}
}

func cmp(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
