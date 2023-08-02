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

import "testing"

var data = map[string]string{
	"a": "A",
	"t": "T",
	"k": "кириллица",
	"n": "latinicca",
}

func testFormat(t *testing.T, format string, width int, expected string) {
	f := NewFormatter(format)
	actual := f.Format(data, width)

	if expected != actual {
		t.Errorf("Format error.\nExpected: '%s'\nActual: '%s'",
			expected, actual)
	}
}

func TestFormat(t *testing.T) {
	testFormat(t, "{foo bar}", 10,
		"foo bar   ")
	testFormat(t, "{foobar}", 3,
		"foo")
	testFormat(t, "{5:a}", 10,
		"    a     ")
	testFormat(t, "{%a} - {%t}", 5,
		"A - T")
	testFormat(t, "{10:%a}", 15,
		"         A     ")
	testFormat(t, "{-10:%a}", 10,
		"A         ")
	testFormat(t, "{-5:%a}", 10,
		"A         ")
	testFormat(t, "{-*%:%a}{*%:%t}", 10,
		"A        T")
	testFormat(t, "{-50%:%a}{50%:%a}", 10,
		"A        A")
	testFormat(t, "{-50%:%a}{*%:%a}", 10,
		"A        A")
	testFormat(t, "{%a %t %a}", 5,
		"A T A")
	testFormat(t, "{5:%a}", 5,
		"    A")
	testFormat(t, "{*%:%a}{*%:%a}{*%:%a}", 10,
		"  A  A   A")
	testFormat(t, "{*%:%a}{*%:%a}", 3,
		"A A")
	testFormat(t, "{-20%:%a}", 10,
		"A         ")
	testFormat(t, "{-50%:%k}{50%:%k}", 40,
		"кириллица                      кириллица")
	testFormat(t, "{-50%:%k}{50%:test}", 15,
		"кирилли    test")
}
