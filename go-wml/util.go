// This file is part of Go WML.
//
// Go WML is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Go WML is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Go WML.  If not, see <https://www.gnu.org/licenses/>.

package wml

import (
	"regexp"
	"strings"
)

func EscapeString(s string) string {
	return strings.ReplaceAll(s, "\"", "\"\"")
}

func IndentString(text string, indent uint) string {
	r, _ := regexp.Compile(`(?m)^([^$])`)
	return r.ReplaceAllString(text, strings.Repeat("\t", int(indent))+"$1")
}
