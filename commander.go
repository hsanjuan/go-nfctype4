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

package nfctype4

import (
	"errors"
	"fmt"

	"github.com/hsanjuan/go-nfctype4/apdu"
)

// Commander can be used to perform the NDEF Type 4 Tag Command Set
// operations: Select, ReadBinary and UpdateBinary
//
// A Commander produces the right Command APDUs, serializes them and
// sends them to a CommandDriver.TransceiveBytes. The response is
// de-serialized into a Response APDU and processed.
type Commander struct {
	// Driver is the CommandDriver in charge of communicating with the tags.
	Driver CommandDriver
}

// Select perfoms a select operation by file ID
// It returns an error if something fails, like cases when the
// response does not indicate success.
func (cmder *Commander) Select(fileID uint16) error {
	if cmder.Driver == nil {
		return errors.New("command driver not set")
	}
	cApdu := apdu.NewSelectAPDU(fileID)
	cApduBytes, err := cApdu.Marshal()
	if err != nil {
		return err
	}
	maxRXLen := cApdu.GetLe() + 2 // For SW bytes
	response, err := cmder.Driver.TransceiveBytes(cApduBytes, int(maxRXLen))
	if err != nil {
		return err
	}

	rApdu := new(apdu.RAPDU)
	if _, err = rApdu.Unmarshal(response); err != nil {
		return err
	}

	if rApdu.CommandCompleted() {
		return nil
	} else if rApdu.FileNotFound() {
		return fmt.Errorf("Commander.Select: "+
			"File %02xh not found", fileID)
	} else {
		return fmt.Errorf("Select: "+
			"Unknown error. SW1: %02xh. SW2: %02xh",
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
	cApdu := apdu.NewReadBinaryAPDU(offset, length)
	cApduBytes, err := cApdu.Marshal()
	if err != nil {
		return nil, err
	}
	response, err := cmder.Driver.TransceiveBytes(cApduBytes, int(length)+2)
	if err != nil {
		return nil, err
	}

	rApdu := new(apdu.RAPDU)
	if _, err = rApdu.Unmarshal(response); err != nil {
		return nil, err
	}
	if rApdu.CommandCompleted() {
		return rApdu.ResponseBody, nil
	}

	return nil, fmt.Errorf("Commander.ReadBinary: "+
		"Error. SW1: %02xh. SW2: %02xh",
		rApdu.SW1,
		rApdu.SW2)
}

// UpdateBinary performs an update operation, which
// allows to erase and write the NDEF file.
func (cmder *Commander) UpdateBinary(buf []byte, offset uint16) error {
	if cmder.Driver == nil {
		return errors.New("Command driver not set")
	}
	cApdu := apdu.NewUpdateBinaryAPDU(buf, offset)
	cApduBytes, err := cApdu.Marshal()
	if err != nil {
		return err
	}
	response, err := cmder.Driver.TransceiveBytes(cApduBytes, 2) // SW bytes
	if err != nil {
		return err
	}

	rApdu := new(apdu.RAPDU)
	if _, err = rApdu.Unmarshal(response); err != nil {
		return err
	}
	if rApdu.CommandCompleted() {
		return nil
	}

	return fmt.Errorf("Commander.UpdateBinary: "+
		"Error. SW1: %02xh. SW2: %02xh",
		rApdu.SW1,
		rApdu.SW2)
}

// NDEFApplicationSelect performs a Select operation on the NDEF
// application (which is basically the first step to use a NDEF Application).
// It returns an error if something goes wrong.
func (cmder *Commander) NDEFApplicationSelect() error {
	if cmder.Driver == nil {
		return errors.New("Commander.NDEFApplicationSelect: " +
			"Driver not set")
	}
	cApdu := apdu.NewNDEFTagApplicationSelectAPDU()
	cApduBytes, err := cApdu.Marshal()
	if err != nil {
		return err
	}
	maxRXLen := cApdu.GetLe() + 2 // For SW bytes
	response, err := cmder.Driver.TransceiveBytes(cApduBytes, int(maxRXLen))
	if err != nil {
		return err
	}

	rApdu := new(apdu.RAPDU)
	if _, err = rApdu.Unmarshal(response); err != nil {
		return err
	}

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
