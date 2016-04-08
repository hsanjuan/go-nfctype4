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

package helpers

import (
	"testing"
)

func TestConversions(t *testing.T) {
	testcases := []struct {
		Bytes [2]byte
		N     uint16
	}{
		{[2]byte{0, 1}, 1},
		{[2]byte{0xff, 0xff}, 65535},
		{[2]byte{0, 0}, 0},
	}
	for _, c := range testcases {
		n := BytesToUint16(c.Bytes)
		tmpBytes := Uint16ToBytes(n)
		n2 := BytesToUint16(tmpBytes)
		if n2 != c.N {
			t.Fail()
		}
	}
}
