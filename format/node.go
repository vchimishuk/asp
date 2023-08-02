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
	"fmt"
	"strconv"
	"unicode/utf8"
)

// node is a AST like node in which format expression is parsed.
type node interface {
	// format formats part of the format expression represented by
	// this node.
	format(data map[string]string) string
	// repr is used only for testing. It should not be used in any code
	// instead of *_test.go files.
	repr() string
}

// textNode represents AST text node. Since it is a simple string
// it evaluates to self.
type textNode string

func newTextNode(s string) node {
	return textNode(s)
}

func (n textNode) format(data map[string]string) string {
	return string(n)
}

func (n textNode) repr() string {
	return n.format(nil)
}

// substNode represents substitution AST node.
type substNode struct {
	width     int
	percent   bool
	alignLeft bool
	fmt       string
	keys      []string
}

func newSubstNode(width int, percent bool, alignLeft bool,
	fmt string, keys []string) node {

	return &substNode{width: width,
		percent:   percent,
		alignLeft: alignLeft,
		fmt:       fmt,
		keys:      keys}
}

func (n *substNode) format(data map[string]string) string {
	var f string

	if n.width != -1 {
		var sign string

		if n.alignLeft {
			sign = "-"
		} else {
			sign = ""
		}
		f = fmt.Sprintf("%%%s%ds", sign, n.width)
	} else {
		f = "%s"
	}

	params := make([]interface{}, 0, len(n.keys))
	for _, k := range n.keys {
		params = append(params, data[k])
	}

	s := fmt.Sprintf(f, fmt.Sprintf(n.fmt, params...))
	if n.width != -1 && utf8.RuneCountInString(s) > n.width {
		return string([]rune(s)[:n.width])
	} else {
		return s
	}
}

func (n *substNode) repr() string {
	repr := "{"

	if n.width >= 0 {
		if n.alignLeft {
			repr += "-"
		}
		repr += strconv.Itoa(n.width)
		if n.percent {
			repr += "%"
		}
	} else if n.percent {
		if n.alignLeft {
			repr += "-"
		}
		repr += "*%"
	}

	repr += ":" + n.fmt + "}"

	return repr
}
