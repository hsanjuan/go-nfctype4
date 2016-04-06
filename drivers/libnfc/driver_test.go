// +build !nolibnfc

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

package libnfc

import (
	"fmt"
	"github.com/hsanjuan/nfctype4"
)

func ExampleDevice_Read_libnfcCommandDriver() {
	// Before running, make sure that the NFC reader device
	// is detected by libnfc and that the tag is in contact
	// with the device as it will be read right away or fail.
	driver := new(Driver) // Set Driver to LibNFC
	device := new(nfctype4.Device)
	device.Setup(driver)
	message, err := device.Read() // Read the tag
	if err != nil {
		fmt.Println(err)
	} else { // See what the NDEF message has
		fmt.Println(message)
	}
}
