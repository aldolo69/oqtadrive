/*
   OqtaDrive - Sinclair Microdrive emulator
   Copyright (c) 2021, Alexander Vollschwitz

   This file is part of OqtaDrive.

   OqtaDrive is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   OqtaDrive is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with OqtaDrive. If not, see <http://www.gnu.org/licenses/>.
*/

package if1

import (
	"strings"
)

//
var keywords = []string{
	"\xa5", "RND",
	"\xa6", "INKEY$",
	"\xa7", "PI",
	"\xa8", "FN",
	"\xa9", "POINT",
	"\xaa", "SCREEN$",
	"\xab", "ATTR",
	"\xac", "AT",
	"\xad", "TAB",
	"\xae", "VAL$",
	"\xaf", "CODE",
	"\xb0", "VAL",
	"\xb1", "LEN",
	"\xb2", "SIN",
	"\xb3", "COS",
	"\xb4", "TAN",
	"\xb5", "ASN",
	"\xb6", "ACS",
	"\xb7", "ATN",
	"\xb8", "LN",
	"\xb9", "EXP",
	"\xba", "INT",
	"\xbb", "SQR",
	"\xbc", "SGN",
	"\xbd", "ABS",
	"\xbe", "PEEK",
	"\xbf", "IN",
	"\xc0", "USR",
	"\xc1", "STR$",
	"\xc2", "CHR$",
	"\xc3", "NOT",
	"\xc4", "BIN",
	"\xc5", "OR",
	"\xc6", "AND",
	"\xc7", "<=",
	"\xc8", ">=",
	"\xc9", "<>",
	"\xca", "LINE",
	"\xcb", "THEN",
	"\xcc", "TO",
	"\xcd", "STEP",
	"\xce", "DEF FN",
	"\xcf", "CAT",
	"\xd0", "FORMAT",
	"\xd1", "MOVE",
	"\xd2", "ERASE",
	"\xd3", "OPEN #",
	"\xd4", "CLOSE #",
	"\xd5", "MERGE",
	"\xd6", "VERIFY",
	"\xd7", "BEEP",
	"\xd8", "CIRCLE",
	"\xd9", "INK",
	"\xda", "PAPER",
	"\xdb", "FLASH",
	"\xdc", "BRIGHT",
	"\xdd", "INVERSE",
	"\xde", "OVER",
	"\xdf", "OUT",
	"\xe0", "LPRINT",
	"\xe1", "LLIST",
	"\xe2", "STOP",
	"\xe3", "READ",
	"\xe4", "DATA",
	"\xe5", "RESTORE",
	"\xe6", "NEW",
	"\xe7", "BORDER",
	"\xe8", "CONTINUE",
	"\xe9", "DIM",
	"\xea", "REM",
	"\xeb", "FOR",
	"\xec", "GO TO",
	"\xed", "GO SUB",
	"\xee", "INPUT",
	"\xef", "LOAD",
	"\xf0", "LIST",
	"\xf1", "LET",
	"\xf2", "PAUSE",
	"\xf3", "NEXT",
	"\xf4", "POKE",
	"\xf5", "PRINT",
	"\xf6", "PLOT",
	"\xf7", "RUN",
	"\xf8", "SAVE",
	"\xf9", "RANDOMIZE",
	"\xfa", "IF",
	"\xfb", "CLS",
	"\xfc", "DRAW",
	"\xfd", "CLEAR",
	"\xfe", "RETURN",
	"\xff", "COPY",
}

//
var keywordReplace = strings.NewReplacer(keywords...)

//
func translate(s string) string {
	if strings.HasPrefix(s, "\x00") {
		return ""
	}
	return strings.Map(spectrumToASCII, keywordReplace.Replace(s))
}

//
func spectrumToASCII(r rune) rune {
	switch {
	case r < '\x20' || r > '\x7f':
		return '-'
	}
	return r
}
