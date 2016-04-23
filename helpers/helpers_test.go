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
	"bytes"
	"errors"
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

func TestGetByte(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Panic expected but didn't happen")
		}
	}()

	buf := bytes.NewBuffer([]byte{1})
	a := GetByte(buf)
	if a != 1 {
		t.Error("a should be 1")
	}

	// This should panic
	GetByte(buf)
}

func TestGetBytes(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Panic expected but didn't happen")
		}
	}()

	buf := bytes.NewBuffer([]byte{1, 2, 3})
	a := GetBytes(buf, 2)
	if a[0] != 1 || a[1] != 2 {
		t.Error("a should be [1,2]")
	}

	b := GetBytes(buf, 0)
	if len(b) != 0 {
		t.Error("b should be empty")
	}

	// This should panic
	GetBytes(buf, 5)
}

func TestHandleErrorPanic(t *testing.T) {
	err := errors.New("hmm")
	a := func() error {
		defer HandleErrorPanic(&err, "Test")
		panic(errors.New("Ops"))
	}
	a()
	if err.Error() != "Test: Ops" {
		t.Error("Expected an Ops error")
	}
}
