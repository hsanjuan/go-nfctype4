/***
    Copyright (c) 2020, Hector Sanjuan

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

package swtag

import (
	"testing"

	"github.com/hsanjuan/go-nfctype4/apdu"
)

type MockTag struct{}

func (t *MockTag) Command(capdu *apdu.CAPDU) *apdu.RAPDU {
	return apdu.NewRAPDU(apdu.RAPDUCommandCompleted)
}

func TestDriver(t *testing.T) {
	d := new(Driver)
	d.String()
	d.Tag = new(MockTag)
	d.String()
	d.Initialize()
	d.String()
	capdu := apdu.NewNDEFTagApplicationSelectAPDU()
	capduBytes, _ := capdu.Marshal()
	rx, _ := d.TransceiveBytes(capduBytes, 2)
	if len(rx) != 2 || rx[0] != 0x90 || rx[1] != 0x00 {
		t.Fail()
	}

	_, err := d.TransceiveBytes(capduBytes, 1)
	if err == nil {
		t.Error("Receiving more bytes than expected should fail")
	}
	d.Close()
}
