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
	"fmt"
)

// Select perfoms a select operation by file ID
// It returns an error if something fails, like cases when the
// response does not indicate success.
func Select(fileID []byte) error {
	if Driver == nil {
		return errors.New("command driver not set")
	}
	cApdu := SelectAPDU(fileID)
	cApduBytes, err := cApdu.Bytes()
	if err != nil {
		return err
	}
	le, _ := cApdu.GetLe()
	maxRXLen := int(le) + 2 // For SW bytes
	response, err := Driver.TransceiveBytes(cApduBytes, maxRXLen)
	if err != nil {
		return err
	}

	rApdu := new(RAPDU)
	rApdu.ParseBytes(response)

	if rApdu.CommandCompleted() {
		return nil
	} else if rApdu.FileNotFound() {
		return fmt.Errorf("Select: File %02x%02xh not found", fileID[0], fileID[1])
	} else {
		return fmt.Errorf("Select: Unknown error. SW1: %02xh. SW2: %02xh",
			rApdu.SW1,
			rApdu.SW2)
	}
}

// ReadBinary performs a read binary operation with the given
// offset and length.
// It returns the Payload of the response (which may be shorter
// than the length provided), or an error if the operation is not
// successful.
func ReadBinary(offset uint16, length uint16) ([]byte, error) {
	if Driver == nil {
		return nil, errors.New("Command driver not set")
	}
	cApdu := ReadBinaryAPDU(offset, length)
	cApduBytes, err := cApdu.Bytes()
	if err != nil {
		return nil, err
	}
	response, err := Driver.TransceiveBytes(cApduBytes, int(length)+2)
	if err != nil {
		return nil, err
	}

	rApdu := new(RAPDU)
	rApdu.ParseBytes(response)
	if rApdu.CommandCompleted() {
		return rApdu.ResponseBody, nil
	}

	return nil, fmt.Errorf("ReadBinary: Error. SW1: %02xh. SW2: %02xh",
		rApdu.SW1,
		rApdu.SW2)
}

// NDEFApplicationSelect performs a Select operation on the NDEF
// application (which is basically the first step to use a NDEF Application).
// It returns an error if something goes wrong.
func NDEFApplicationSelect() error {
	if Driver == nil {
		return errors.New("Driver (CommandDriver) not set")
	}
	cApdu := NDEFTagApplicationSelectAPDU()
	cApduBytes, err := cApdu.Bytes()
	if err != nil {
		return err
	}
	le, _ := cApdu.GetLe()
	maxRXLen := int(le) + 2 // For SW bytes
	response, err := Driver.TransceiveBytes(cApduBytes, maxRXLen)
	if err != nil {
		return err
	}

	rApdu := new(RAPDU)
	rApdu.ParseBytes(response)

	if rApdu.CommandCompleted() {
		return nil
	} else if rApdu.FileNotFound() {
		return errors.New("NDEF Tag Application not found")
	} else {
		return fmt.Errorf("unknown error. SW1: %02xh. SW2: %02xh",
			rApdu.SW1,
			rApdu.SW2)
	}
}

// CapabilityContainerSelect performs a Select operation on the
// Capability Container File, which is necessary before reading its
// contents. It returns an error if the operation fails.
func CapabilityContainerSelect() error {
	bytes := Uint16ToBytes(CCID)
	return Select(bytes[:])
}

// CapabilityContainerRead performs a read binary operation on the
// capability container. It returns an error if the operation fails.
func CapabilityContainerRead() ([]byte, error) {
	// offset: 0. Length: 15
	return ReadBinary(0, 15)
}

// func UpdateBinaryAPDU {

// }
