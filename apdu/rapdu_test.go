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

package apdu

import (
	"bytes"
	"testing"
)

func TestRAPDUMarshalUnmarshal(t *testing.T) {
	testcases := [][]byte{
		{0x00, 0x90, 0x00},
		{0x90, 0x00},
	}
	for _, c := range testcases {
		rapdu := &RAPDU{}
		rapdu.Unmarshal(c)
		rapduBytes, _ := rapdu.Marshal()
		t.Logf("Expected: % 02X", c)
		t.Logf("Got     : % 02X ", rapduBytes)
		if !bytes.Equal(c, rapduBytes) {
			t.Fail()
		}
	}

}

func TestRAPDUNew(t *testing.T) {
	testcases := []int{
		RAPDUCommandCompleted,
		RAPDUCommandNotAllowed,
		RAPDUFileNotFound,
		RAPDUInactiveState,
	}
	for _, c := range testcases {
		rapdu := NewRAPDU(c)
		if rapdu == nil {
			t.Fail()
		}
	}
}
