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

// Package dummy provides a trivial CommandDriver implementation used for
// testing and examples.
package dummy

import (
	"fmt"
)

// Driver implements a CommandDriver which does nothing
// and returns pre-programmed responses when it calls TransceiveBytes.
// It is used for testing, but also can be a simple example of how a
// CommandDriver is implemented.
type Driver struct {
	ReceiveBytes    [][]byte // Responses for every TransceiveBytes call
	ReceiveBytesPos int
}

// Initialize does nothing because it is a DummyDriver.
func (driver *Driver) Initialize() error {
	return nil
}

// String returns information about this driver.
func (driver *Driver) String() string {
	str := "Dummy driver :)"
	return str
}

// TransceiveBytes ignores the data sent, returns one of the elements
// in the ReceiveBytes array, and updates the ReceiveBytesPos to return
// the next one on the next call.
//
// It returns an error if we have already returned all the elements in
// ReceiveBytes at some point.
func (driver *Driver) TransceiveBytes(tx []byte, rxLen int) ([]byte, error) {
	if driver.ReceiveBytesPos >= len(driver.ReceiveBytes) {
		return nil, fmt.Errorf("Driver.TransceiveBytes: "+
			"no data to return (index %d)", driver.ReceiveBytesPos)
	}
	response := driver.ReceiveBytes[driver.ReceiveBytesPos]
	driver.ReceiveBytesPos = driver.ReceiveBytesPos + 1
	return response, nil
}

// Close does nothing because this is a DummyDriver.
func (driver *Driver) Close() {
	return
}
