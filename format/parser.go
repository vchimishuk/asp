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
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	// All possible parser atomata states.
	stateText = iota
	stateSubst
)

func parse(s string, width int) (nodes []node, err error) {
	i := 0
	l := len(s)
	state := stateText
	nodes = make([]node, 0, 8)

	for i < l {
		var n node
		var read int

		switch state {
		case stateSubst:
			n, read, err = parseSubst(s, i)
			state = stateText
		case stateText:
			n, read, err = parseText(s, i)
			state = stateSubst
			// If text was started with empty text node.
			if read == 0 {
				continue
			}
		default:
			panic(nil)
		}

		if err != nil {
			return nil, err
		}

		nodes = append(nodes, node(n))
		i += read
		state = stateText

	}

	// Do some validation now: check if percent usage is correct;
	// and percentages sum is not more than 100%.
	sNodes := 0
	pNodes := 0
	starNodes := 0
	percentSum := 0
	staticPercWidth := 0
	staticTextWidth := 0

	for _, n := range nodes {
		if sn, ok := n.(*substNode); ok {
			sNodes += 1
			if sn.percent {
				pNodes += 1
				if sn.width == -1 {
					starNodes += 1
				} else {
					percentSum += sn.width
				}
			} else {
				staticPercWidth += sn.width
			}
		} else if tn, ok := n.(textNode); ok {
			staticTextWidth += len(tn)
		} else {
			panic(nil)
		}
	}
	if pNodes != 0 && pNodes != sNodes {
		return nil, fmt.Errorf("All subsitutions should have percent " +
			"width value or none of them.")
	}
	if percentSum > 100 {
		return nil, fmt.Errorf("Sum of percent widths can not " +
			"be more than 100%%.")
	}

	// Now every *% substitution width should be calculated. In percents for now.
	if starNodes > 0 {
		w := (100 - percentSum) / starNodes
		// 100% / 3 == 33%. So we have to correct one node (+1%) to get 100%
		corr := 100 - percentSum - w*starNodes

		for _, n := range nodes {
			if sn, ok := n.(*substNode); ok {
				if sn.percent && sn.width == -1 {
					sn.width = w + corr
					// Correct only the first node.
					corr = 0
				}
			}
		}
	}

	// Convert all percet widths to absolute values.
	// Space given for all percent fields.
	percWidth := width - staticPercWidth - staticTextWidth
	onePercWidth := float64(percWidth) / 100
	actPercWidthSum := 0
	var last *substNode

	for _, n := range nodes {
		if sn, ok := n.(*substNode); ok {
			if sn.percent {
				sn.width = round(float64(sn.width) * onePercWidth)
				actPercWidthSum += sn.width
				last = sn
			}
		}
	}

	// Our math may has some error, so just correct the last node to make
	// format to match the resulting width.
	if last != nil {
		last.width += percWidth - actPercWidthSum
	}

	return nodes, nil
}

func parseText(s string, pos int) (n node, read int, err error) {
	s = s[pos:]
	i := 0
	tok := ""

loop:
	for l := len(s); i < l; i++ {
		r := s[i]

		switch r {
		case '{':
			if i+1 < l && s[i+1] == '{' {
				i++
			} else {
				break loop
			}
		case '}':
			if i+1 < l && s[i+1] == '}' {
				i++
			} else {
				err := fmt.Errorf("{ expected but } found.")
				return nil, i, err
			}
		}

		tok += string(r)
	}

	return newTextNode(tok), i, nil
}

func parseSubst(s string, pos int) (n node, read int, err error) {
	s = s[pos:]
	if !strings.HasPrefix(s, "{") {
		panic(nil)
	}

	i := strings.IndexRune(s, '}')
	if i == -1 {
		return nil, 0, fmt.Errorf("} expected but end of text reached.")
	}

	read = i + 1
	tok := s[1:i]
	parts := strings.Split(tok, ":")

	switch len(parts) {
	case 1:
		format, keys, err := parseSubstFormat(parts[0])
		if err != nil {
			return nil, read, err
		}
		n = newSubstNode(-1, false, false, format, keys)
	case 2:
		width, percent, alignLeft, err := parseSubstStarWidth(parts[0])
		if err != nil {
			width, percent, alignLeft, err = parseSubstNumWidth(parts[0])
			if err != nil {
				return nil, read, err
			}
		}
		format, keys, err := parseSubstFormat(parts[1])
		if err != nil {
			return nil, read, err
		}
		n = newSubstNode(width, percent, alignLeft, format, keys)
	default:
		return nil, read, fmt.Errorf(
			"Illegal formatter substitution format. Column: %d.", pos)
	}

	return n, read, nil
}

func parseSubstFormat(s string) (format string, keys []string, err error) {
	format = s
	keys = make([]string, 0, 2)
	re := regexp.MustCompile("%[\\w%]")
	indexes := re.FindAllStringIndex(s, -1)

	for _, index := range indexes {
		beg := index[0]
		v := rune(s[beg+1])

		if v != '%' {
			keys = append(keys, string(v))
			format = format[:beg+1] + "s" + format[beg+2:]
		}
	}

	return format, keys, nil
}

func parseSubstStarWidth(s string) (width int, percent bool, alignLeft bool,
	err error) {

	if s == "*%" || s == "-*%" {
		alignLeft = strings.HasPrefix(s, "-")

		return -1, true, alignLeft, nil
	}

	return 0, false, false, fmt.Errorf("Failed to parse width.")
}

func parseSubstNumWidth(s string) (width int, percent bool, alignLeft bool,
	err error) {

	if strings.HasSuffix(s, "%") {
		percent = true
		s = s[:len(s)-1]
	}

	width, err = strconv.Atoi(s)
	if err != nil {
		return 0, false, false, err
	}

	if width < 0 {
		alignLeft = true
		width *= -1
	}

	return width, percent, alignLeft, nil
}

func round(f float64) int {
	frac := f - math.Floor(f)

	if frac <= 0.5 {
		return int(math.Floor(f))
	} else {
		return int(math.Ceil(f))
	}
}
