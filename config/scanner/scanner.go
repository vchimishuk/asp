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
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Scanner struct {
	s    *bufio.Scanner
	line int
	err  error
	key  string
	val  string
}

func New(r io.Reader) *Scanner {
	return &Scanner{s: bufio.NewScanner(r)}
}

func (s *Scanner) Scan() bool {
	if s.err != nil {
		return false
	}

	for s.s.Scan() {
		s.line++
		l := trim(s.s.Text())

		if l != "" {
			pts := strings.SplitN(l, "=", 2)
			if len(pts) != 2 {
				f := "invalid key=value format at line %d"
				s.err = fmt.Errorf(f, s.line)
				return false
			}
			s.key = trim(pts[0])
			s.val = trim(pts[1])

			return true
		}
	}
	s.err = s.s.Err()

	return false
}

func (s *Scanner) Key() string {
	return s.key
}

func (s *Scanner) Val() string {
	return s.val
}

func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) Line() int {
	return s.line
}

func trim(s string) string {
	i := strings.Index(s, "#")
	if i != -1 {
		s = s[0:i]
	}

	return strings.Trim(s, " ")
}
