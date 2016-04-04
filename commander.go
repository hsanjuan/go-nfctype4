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

// Commander are capable of performing the NDEF Type 4 Tag Command Set,
// operations, by using a CommandDriver
type Commander struct {
	// Driver is the CommandDriver in charge of communicating with the
	// NFC device.
	Driver CommandDriver
}

// Select perfoms a select operation by file ID
// It returns an error if something fails, like cases when the
// response does not indicate success.
func (cmder *Commander) Select(fileID []byte) error {
	if cmder.Driver == nil {
		return errors.New("command driver not set")
	}
	cApdu := NewSelectAPDU(fileID)
	cApduBytes, err := cApdu.Bytes()
	if err != nil {
		return err
	}
	le, _ := cApdu.GetLe()
	maxRXLen := int(le) + 2 // For SW bytes
	response, err := cmder.Driver.TransceiveBytes(cApduBytes, maxRXLen)
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
func (cmder *Commander) ReadBinary(offset uint16, length uint16) ([]byte, error) {
	if cmder.Driver == nil {
		return nil, errors.New("Command driver not set")
	}
	cApdu := NewReadBinaryAPDU(offset, length)
	cApduBytes, err := cApdu.Bytes()
	if err != nil {
		return nil, err
	}
	response, err := cmder.Driver.TransceiveBytes(cApduBytes, int(length)+2)
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

// UpdateBinary performs an update operation on the tag, which
// allows to erase and write to a file.
// func (cmder *Commander) UpdateBinaryAPDU() {
// Unimplemented
// }

// NDEFApplicationSelect performs a Select operation on the NDEF
// application (which is basically the first step to use a NDEF Application).
// It returns an error if something goes wrong.
func (cmder *Commander) NDEFApplicationSelect() error {
	if cmder.Driver == nil {
		return errors.New("Commander.NDEFApplicationSelect: " +
			"Driver not set")
	}
	cApdu := NewNDEFTagApplicationSelectAPDU()
	cApduBytes, err := cApdu.Bytes()
	if err != nil {
		return err
	}
	le, _ := cApdu.GetLe()
	maxRXLen := int(le) + 2 // For SW bytes
	response, err := cmder.Driver.TransceiveBytes(cApduBytes, maxRXLen)
	if err != nil {
		return err
	}

	rApdu := new(RAPDU)
	rApdu.ParseBytes(response)

	if rApdu.CommandCompleted() {
		return nil
	} else if rApdu.FileNotFound() {
		return errors.New("Commander.NDEFApplicationSelect: " +
			"NDEF Tag Application not found")
	} else {
		return fmt.Errorf("Commander.NDEFApplicationSelect: "+
			"unknown error. SW1: %02xh. SW2: %02xh",
			rApdu.SW1,
			rApdu.SW2)
	}
}

// CapabilityContainerSelect performs a Select operation on the
// Capability Container File, which is necessary before reading its
// contents. It returns an error if the operation fails.
func (cmder *Commander) CapabilityContainerSelect() error {
	bytes := uint16ToBytes(CCID)
	return cmder.Select(bytes[:])
}

// CapabilityContainerRead performs a read binary operation on the
// capability container. It returns an error if the operation fails.
func (cmder *Commander) CapabilityContainerRead() ([]byte, error) {
	// offset: 0. Length: 15
	return cmder.ReadBinary(0, 15)
}
