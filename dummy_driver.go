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
	"errors"
)

/*
 * Implements a CommandDriver which does nothing
 *
 */

type DummyCommandDriver struct {
	ReceiveBytes    [][]byte // On every Transceive it will return one of these
	ReceiveBytesPos int
}

func (driver *DummyCommandDriver) Initialize() error {
	return nil
}

func (driver *DummyCommandDriver) String() string {
	str := "Dummy driver :)"
	return str
}

func (driver *DummyCommandDriver) TransceiveBytes(tx []byte, rx_len int) ([]byte, error) {
	if driver.ReceiveBytesPos >= len(driver.ReceiveBytes) {
		return nil, errors.New("Dummy Driver: no data to return")
	}
	response := driver.ReceiveBytes[driver.ReceiveBytesPos]
	driver.ReceiveBytesPos = driver.ReceiveBytesPos + 1
	return response, nil
}

func (driver *DummyCommandDriver) Close() {
	return
}
