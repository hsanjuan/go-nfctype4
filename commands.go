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

// Select a file by ID
func Select(file_id []byte) error {
	if Driver == nil {
		return errors.New("Command driver not set")
	}
	c_apdu := SelectAPDU(file_id)
	c_apdu_bytes, err := c_apdu.Bytes()
	if err != nil {
		return err
	}
	le, _ := c_apdu.GetLe()
	max_rx_len := int(le) + 2 // For SW bytes
	response, err := Driver.TransceiveBytes(c_apdu_bytes, max_rx_len)
	if err != nil {
		return err
	}

	r_apdu := new(R_APDU)
	r_apdu.ParseBytes(response)

	if r_apdu.CommandCompleted() {
		return nil
	} else if r_apdu.FileNotFound() {
		return fmt.Errorf("Select: File %02x%02xh not found", file_id[0], file_id[1])
	} else {
		return fmt.Errorf("Select: Unknown error. SW1: %02xh. SW2: %02xh",
			r_apdu.SW1,
			r_apdu.SW2)
	}
}

func ReadBinary(offset uint16, length uint16) ([]byte, error) {
	if Driver == nil {
		return nil, errors.New("Command driver not set")
	}
	c_apdu := ReadBinaryAPDU(offset, length)
	c_apdu_bytes, err := c_apdu.Bytes()
	if err != nil {
		return nil, err
	}
	response, err := Driver.TransceiveBytes(c_apdu_bytes, int(length)+2)
	if err != nil {
		return nil, err
	}

	r_apdu := new(R_APDU)
	r_apdu.ParseBytes(response)
	if r_apdu.CommandCompleted() {
		return r_apdu.ResponseBody, nil
	} else {
		return nil, fmt.Errorf("ReadBinary: Unknown error. SW1: %02xh. SW2: %02xh",
			r_apdu.SW1,
			r_apdu.SW2)
	}
}

func NDEFApplicationSelect() error {
	if Driver == nil {
		return errors.New("Command driver not set")
	}
	c_apdu := NDEFTagApplicationSelectAPDU()
	c_apdu_bytes, err := c_apdu.Bytes()
	if err != nil {
		return err
	}
	le, _ := c_apdu.GetLe()
	max_rx_len := int(le) + 2 // For SW bytes
	response, err := Driver.TransceiveBytes(c_apdu_bytes, max_rx_len)
	if err != nil {
		return err
	}

	r_apdu := new(R_APDU)
	r_apdu.ParseBytes(response)

	if r_apdu.CommandCompleted() {
		return nil
	} else if r_apdu.FileNotFound() {
		return errors.New("NDEF Tag Application not found")
	} else {
		return fmt.Errorf("Unknown error. SW1: %02xh. SW2: %02xh",
			r_apdu.SW1,
			r_apdu.SW2)
	}
}

func CapabilityContainerSelect() error {
	bytes := Uint16ToBytes(CC_ID)
	return Select(bytes[:])
}

func CapabilityContainerRead() ([]byte, error) {
	// offset: 0. Length: 15
	return ReadBinary(0, 15)
}

// func UpdateBinaryAPDU {

// }
