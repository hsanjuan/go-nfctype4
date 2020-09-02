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

// Package swtag provides a CommandDriver implementation which
// acts as a binary interface to software-based NFC Type 4 Tags which
// implement the tags.Tag interface.
package swtag

import (
	"errors"

	"github.com/hsanjuan/go-nfctype4/apdu"
	"github.com/hsanjuan/go-nfctype4/tags"
)

// Driver implements a CommandDriver to interface with a software tag
// (something that implements the Tag interface from the tags module).
//
// This means that this driver provides the binary channel to those tags.
//
// Among its applications, is the easily test that a software Tag
// conforms to the NFC Type 4 Tag specification, by attaching this
// driver to a `nfctype4.Device` with, and performing the Device
// operations on the Tag.
//
// The second application is to simulate a NFC Type 4 with a hardware
// NFC reader. Libnfc, for example, allows to initialize NFC Readers in
// Target mode, where the Libnfc device behaves like a tag rather than
// a reader. This driver makes it trivial to provide a libnfc device in
// Target mode with full-fledged Type 4 Tag behaviour.
//
// An example of how this is done can be seen visiting the following
// snippet: https://gitlab.com/snippets/18718
type Driver struct {
	Tag tags.Tag
}

// Initialize does nothing because software Tags don't need initialization.
func (driver *Driver) Initialize() error {
	return nil
}

// String returns information about this driver.
func (driver *Driver) String() string {
	str := "Software Tag Driver. "
	if driver.Tag != nil {
		str += "Driver.Tag is not defined."
	} else {
		str += "Ready."
	}
	return str
}

// TransceiveBytes parses the tx bytes to a Command APDU and uses the Tag
// to process the APDU and provide a Response APDU, which is in turn
// serialized and returned.
//
// It returns an error if the Tag field has not been set, if the APDUs
// cannot be serialized or deserialized, or if the response size
// is bigger than the expected size.
func (driver *Driver) TransceiveBytes(tx []byte, rxLen int) ([]byte, error) {
	if driver.Tag == nil {
		return nil, errors.New("Driver.TransceiveBytes: " +
			"Driver.Tag is not set.")
	}

	capdu := new(apdu.CAPDU)
	if _, err := capdu.Unmarshal(tx); err != nil {
		return nil, err
	}
	// fmt.Println(capdu)
	rapdu := driver.Tag.Command(capdu)
	// fmt.Println(rapdu)
	rxBuf, err := rapdu.Marshal()
	if err != nil {
		return nil, err
	}

	if len(rxBuf) > rxLen {
		return rxBuf, errors.New("Driver.TransceiveBytes: " +
			"The length of the response is larger than expected")
	}
	return rxBuf, nil
}

// Close does nothing.
func (driver *Driver) Close() {
	return
}
