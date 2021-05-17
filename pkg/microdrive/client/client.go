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

package client

//
type Client int

const (
	IF1 Client = iota
	QL
)

//
func (c Client) String() string {

	switch c {

	case IF1:
		return "Interface 1"

	case QL:
		return "QL"

	default:
		return "<unknown>"
	}
}

//
func (c Client) DefaultFormat() string {

	switch c {

	case IF1:
		return "mdr"

	case QL:
		return "mdv"

	default:
		return ""
	}
}
