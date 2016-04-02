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
	"bytes"
	"errors"
)

const CC_ID = uint16(0xE103) // Capability container ID
const NDEF_APPLICATION_NAME = uint64(0xD2760000850101)

type C_APDU struct {
	// https://en.wikipedia.org/wiki/Smart_card_application_protocol_data_unit
	CLA  byte   //Class byte
	INS  byte   //Instruction byte
	P1   byte   //Param byte 1
	P2   byte   //Param byte 2
	Lc   []byte //Data field length // 0, 1 or 3 bytes
	Data []byte //Data field
	Le   []byte //Expected response length // 0, 1, 2, or 3 bytes
}

type R_APDU struct {
	ResponseBody []byte //Data bytes
	SW1          byte   //Status Word 1
	SW2          byte   //Status Word 2
}

func (apdu *C_APDU) GetLc() (uint16, error) {
	switch len(apdu.Lc) {
	case 0:
		return uint16(0), nil
	case 1:
		return uint16(apdu.Lc[0]), nil
	case 3:
		return BytesToUint16([2]byte{apdu.Lc[1], apdu.Lc[2]}), nil
	default:
		return 0, errors.New("C_APDU wrong Lc")
	}
}

func (apdu *C_APDU) SetLc(n uint16) {
	if n == 0 {
		apdu.Lc = []byte{}
	} else if 1 <= n && n <= 255 { // 1-255
		apdu.Lc = []byte{byte(n)}
	} else {
		n_bytes := Uint16ToBytes(n)
		apdu.Lc = []byte{0x00, n_bytes[0], n_bytes[1]}
	}
}

func (apdu *C_APDU) GetLe() (uint16, error) {
	switch len(apdu.Le) {
	case 0:
		return uint16(0), nil
	case 1:
		if n := apdu.Le[0]; n == 0 {
			return uint16(256), nil
		} else {
			return uint16(n), nil
		}
	case 2:
		n0 := apdu.Le[0]
		n1 := apdu.Le[1]
		if n0 == 0 && n1 == 0 {
			//return uint16(65536) // Overflow! FIXME!
			return uint16(65535), nil
		} else {
			return BytesToUint16([2]byte{n0, n1}), nil
		}
	case 3:
		return BytesToUint16([2]byte{apdu.Le[1], apdu.Le[2]}), nil
	default:
		return 0, errors.New("C_APDU wrong Le")
	}
}

// FIXME: We dont support Le = 65536 (overflows uint16)
func (apdu *C_APDU) SetLe(n uint16) {
	if n == 0 {
		apdu.Le = []byte{}
	} else if 1 <= n && n <= 255 {
		apdu.Le = []byte{byte(n)}
	} else if n == 256 {
		apdu.Le = []byte{byte(0)}
	} else {
		n_bytes := Uint16ToBytes(n)
		if len(apdu.Lc) > 0 { // Make it 2 bytes
			apdu.Le = []byte{n_bytes[0], n_bytes[1]}
		} else { // 3 bytes then
			apdu.Le = []byte{0, n_bytes[0], n_bytes[1]}
		}
	}
}

// Do some testing, mainly around the Lc, Data, Le fields
func (apdu *C_APDU) Test() error {
	lc := apdu.Lc
	le := apdu.Le
	// Test Lc
	switch {
	case len(lc) == 1 && lc[0] == 0:
		return errors.New("APDU Lc with 1 byte cannot be 0")
	case len(lc) == 2:
		return errors.New("APDU Lc cannot have 2 bytes")
	case len(lc) == 3:
		if lc[0] != 0 {
			return errors.New("APDU 3-byte-Lc's first byte must be 0")
		} else if lc[1] == 0 && lc[2] == 0 {
			return errors.New("APDU 3-byte-Lc cannot be all 0s")
		}
	case len(lc) > 3:
		return errors.New("APDU Lc cannot have more than 3")
	}

	// Test Le
	switch {
	case len(le) == 2 && len(lc) == 0:
		return errors.New("APDU 2-byte-Le needs Lc present")
	case len(le) == 3:
		if len(lc) != 0 {
			return errors.New("APDU 3-byte-Le is only compatible with empty Lc")
		} else if le[0] != 0 {
			return errors.New("APDU 3-byte-Le's first byte must be 0")
		}
	case len(le) > 3:
		return errors.New("APDU Le cannot have more 3 bytes")
	}

	data_length, _ := apdu.GetLc()
	if int(data_length) != len(apdu.Data) {
		return errors.New("APDU Lc value is different from the actual data length")
	}
	return nil
}

func (apdu *C_APDU) Bytes() ([]byte, error) {
	if err := apdu.Test(); err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.WriteByte(apdu.CLA)
	buffer.WriteByte(apdu.INS)
	buffer.WriteByte(apdu.P1)
	buffer.WriteByte(apdu.P2)
	buffer.Write(apdu.Lc)
	buffer.Write(apdu.Data)
	buffer.Write(apdu.Le)
	return buffer.Bytes(), nil
}

// Returns the number of bytes parsed or an error
func (apdu *R_APDU) ParseBytes(bytes []byte) (int, error) {
	length := len(bytes)
	apdu.SW1 = bytes[length-2]
	apdu.SW2 = bytes[length-1]
	if length >= 3 {
		apdu.ResponseBody = bytes[0 : length-2]
	}
	return length, nil
}

func (apdu *R_APDU) CommandCompleted() bool {
	return apdu.SW1 == 0x90 && apdu.SW2 == 0x00
}

func (apdu *R_APDU) FileNotFound() bool {
	return apdu.SW1 == 0x6A && apdu.SW2 == 0x82
}

func NDEFTagApplicationSelectAPDU() *C_APDU {
	c_apdu := &C_APDU{
		CLA:  byte(0x00),
		INS:  byte(0xA4),
		P1:   byte(0x04),                                       // Select by name
		P2:   byte(0x00),                                       // First or only occurrence
		Data: []byte{0xD2, 0x76, 0x00, 0x00, 0x85, 0x01, 0x01}, // NDEF app name
	}
	c_apdu.SetLc(7)
	// This would set a single-byte Le to 0, meaning response data field might be present
	// (and up to 256 bytes according to Wikipedia)
	c_apdu.SetLe(256)
	return c_apdu
}

func CapabilityContainerSelectAPDU() *C_APDU {
	bytes := Uint16ToBytes(CC_ID)
	return SelectAPDU(bytes[:])
}

func ReadBinaryAPDU(offset uint16, length uint16) *C_APDU {
	offset_bytes := Uint16ToBytes(offset)
	c_apdu := &C_APDU{
		CLA: byte(0x00),
		INS: byte(0xB0),
		P1:  offset_bytes[0],
		P2:  offset_bytes[1],
	}
	c_apdu.SetLe(length)
	return c_apdu
}

func SelectAPDU(file_id []byte) *C_APDU {
	c_apdu := &C_APDU{
		CLA:  byte(0x00),
		INS:  byte(0xA4),
		P1:   byte(0x00), // Select by Id
		P2:   byte(0x0C), // First or only occurrence
		Data: file_id,
	}
	c_apdu.SetLc(uint16(len(file_id))) //File ID length should be 2
	return c_apdu
}

// Only after selecting it!
func CapabilityContainerReadAPDU() *C_APDU {
	return ReadBinaryAPDU(0, 15)
}
