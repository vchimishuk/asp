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

// TODO: Alternative if missed or empty.

import (
	"strings"
	"unicode/utf8"
)

// Formatter allows format structs into string.
// Formatting rules are described with special mini language which allows
// to align fields.
//
// Formatting language description. Format string is a regular string which
// may includes special substitution forms. Substitution forms are surrounded
// with curly braces. For instance: "foo {%a} bar". This is a simple format
// string with one trivial substitution ("{%a}") inside. Substitution has
// a following format: {[[-]width[%]]:name}. Where "width" is a number of chars
// which will be allocated for substitution's result text. Optional minus sign
// sets aligning to the left, instead of right which is the default. With % sign
// width can be set in percents instead of absolute value. Percent width values
// are calculated during formatter creation depends on the given parameters.
// "format" is a format string like in Printf is used. Format verbs should be
// prefixed with percent sign. Available format verb values depends on data
// map parameter passed to the Format method.
// Examples.
// If data is map[string]string{
//         "a": "A",
//         "b": "B",
// }
//
// "foo {%a} bar {%b}" formatted to "foo A bar B"
// "{{{10:%a-%b}}}" formatted to "{       A-B}"
// "{{{-10:%a-%b}}}" formatted to "{A-B       }"
type Formatter interface {
	Format(data map[string]string, width int) string
}

type formatter struct {
	// format is the formatting pattern.
	format string
	// To prevent format compilation every time we compile
	// it once per width and cache it.
	nodes []node
	width int
}

// Validate returns an error if given not valid formatter string.
// Validate usually should be called before NewFormatter to be sure that
// runtime error will not happend during formatting later.
func Validate(format string) error {
	_, err := parse(format, 100)

	return err
}

// NewFormatter returns new Formatter object for the given format pattern.
func NewFormatter(format string) Formatter {
	return &formatter{format: format}
}

func (f *formatter) Format(data map[string]string, width int) string {
	if f.width != width {
		f.nodes = nil
	}
	if f.nodes == nil {
		nodes, err := parse(f.format, width)
		if err != nil {
			panic(err)
		}
		f.nodes = nodes
		f.width = width
	}

	res := ""
	for _, n := range f.nodes {
		res += n.format(data)
	}

	// Resulting text should fit to the required width.
	l := utf8.RuneCountInString(res)
	if l < f.width {
		res += strings.Repeat(" ", f.width-l)
	} else {
		res = truncate(res, f.width)
	}

	return res
}

func truncate(s string, size int) string {
	bytes := 0
	runes := 0
	p := s
	for len(p) > 0 && runes < size {
		_, size := utf8.DecodeRuneInString(p)
		bytes += size
		runes += 1
		p = p[size:]
	}

	if bytes < len(s) {
		return s[:bytes]
	} else {
		return s
	}
}
