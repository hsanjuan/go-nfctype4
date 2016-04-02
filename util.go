/***
    Copyright (c) 2016, Hector Sanjuan

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Lesser General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Lesser General Public License for more details.

    You should have received a copy of the GNU Lesser General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
***/

package nfctype4

import ()

func BytesToUint16(field [2]byte) uint16 {
	return uint16(field[0])<<8 | uint16(field[1])
}

func Uint16ToBytes(value uint16) [2]byte {
	byte0 := byte(value >> 8)
	byte1 := byte(0x00ff & value) //Probably the casting would suffice
	return [2]byte{byte0, byte1}
}
