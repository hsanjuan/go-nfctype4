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

import (
	"fmt"
)

/*
 * Implements a CommandDriver which does nothing
 *
 */

// DummyCommandDriver implements a CommandDriver which does nothing
// and returns pre-programmed responses when it calls TransceiveBytes.
// It is used for testing, but also can be a simple example of how a
// CommandDriver is implemented.
type DummyCommandDriver struct {
	ReceiveBytes    [][]byte // Responses for every TransceiveBytes call
	ReceiveBytesPos int
}

// Initialize does nothing because it is a DummyDriver.
func (driver *DummyCommandDriver) Initialize() error {
	return nil
}

// String returns information about this driver.
func (driver *DummyCommandDriver) String() string {
	str := "Dummy driver :)"
	return str
}

// TransceiveBytes ignores the data sent, returns one of the elements
// in the ReceiveBytes array, and updates the ReceiveBytesPos to return
// the next one on the next call.
//
// It returns an error if we have already returned all the elements in
// ReceiveBytes at some point.
func (driver *DummyCommandDriver) TransceiveBytes(tx []byte, rxLen int) ([]byte, error) {
	if driver.ReceiveBytesPos >= len(driver.ReceiveBytes) {
		return nil, fmt.Errorf("DummyCommandDriver.TransceiveBytes: "+
			"no data to return (index %d)", driver.ReceiveBytesPos)
	}
	response := driver.ReceiveBytes[driver.ReceiveBytesPos]
	driver.ReceiveBytesPos = driver.ReceiveBytesPos + 1
	return response, nil
}

// Close does nothing because this is a DummyDriver.
func (driver *DummyCommandDriver) Close() {
	return
}
