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

// CCID is the Capability container ID.
const CCID = uint16(0xE103)

// NDEFAPPLICATION is the name for the NDEF Application.
const NDEFAPPLICATION = uint64(0xD2760000850101)

// CAPDU represents a Command APDU
// (https://en.wikipedia.org/wiki/Smart_card_application_protocol_data_unit)
// which is used to send instructions and data to the NFC devices.
type CAPDU struct {
	//
	CLA  byte   //Class byte
	INS  byte   //Instruction byte
	P1   byte   //Param byte 1
	P2   byte   //Param byte 2
	Lc   []byte //Data field length // 0, 1 or 3 bytes
	Data []byte //Data field
	Le   []byte //Expected response length // 0, 1, 2, or 3 bytes
}

// GetLc computes the actual Lc value from the Lc bytes. Lc
// indicates the length of the data sent with a Command
// APDU and goes from 0 to 2^16-1.
func (apdu *CAPDU) GetLc() (uint16, error) {
	switch len(apdu.Lc) {
	case 0:
		return uint16(0), nil
	case 1:
		return uint16(apdu.Lc[0]), nil
	case 3:
		return BytesToUint16([2]byte{apdu.Lc[1], apdu.Lc[2]}), nil
	default:
		return 0, errors.New("CAPDU wrong Lc")
	}
}

// SetLc allows to easily set the value of the Lc bytes making sure
// they comply to the specification.
func (apdu *CAPDU) SetLc(n uint16) {
	if n == 0 {
		apdu.Lc = []byte{}
	} else if 1 <= n && n <= 255 { // 1-255
		apdu.Lc = []byte{byte(n)}
	} else {
		nBytes := Uint16ToBytes(n)
		apdu.Lc = []byte{0x00, nBytes[0], nBytes[1]}
	}
}

// BUG(hector): APDU's Le field could theoretically be 65536 (2^16), but
// this overflows uint16 so it's unsupported by SetLe and GetLe.
// It only happens in the case when Le has two bytes and both are 0 and in this
// case GetLe returns 2^16 -1.

// GetLe computes the actual Le value from the Le bytes. Le
// indicates the maximum length of the data to be received Command
// APDU and goes from 0 to 2^16.
func (apdu *CAPDU) GetLe() (uint16, error) {
	switch len(apdu.Le) {
	case 0:
		return uint16(0), nil
	case 1:
		n := apdu.Le[0]
		if n == 0 {
			return uint16(256), nil
		}
		return uint16(n), nil
	case 2:
		n0 := apdu.Le[0]
		n1 := apdu.Le[1]
		if n0 == 0 && n1 == 0 {
			//return uint16(65536) // Overflow! FIXME!
			return uint16(65535), nil
		}
		return BytesToUint16([2]byte{n0, n1}), nil
	case 3:
		return BytesToUint16([2]byte{apdu.Le[1], apdu.Le[2]}), nil
	default:
		return 0, errors.New("CAPDU wrong Le")
	}
}

// SetLe allows to easily set the value of the Le bytes making sure
// they comply to the specification.
func (apdu *CAPDU) SetLe(n uint16) {
	if n == 0 {
		apdu.Le = []byte{}
	} else if 1 <= n && n <= 255 {
		apdu.Le = []byte{byte(n)}
	} else if n == 256 {
		apdu.Le = []byte{byte(0)}
	} else {
		nBytes := Uint16ToBytes(n)
		if len(apdu.Lc) > 0 { // Make it 2 bytes
			apdu.Le = []byte{nBytes[0], nBytes[1]}
		} else { // 3 bytes then
			apdu.Le = []byte{0, nBytes[0], nBytes[1]}
		}
	}
}

// Test ensures that a CAPDU struct fields are in-line with the
// specification.
// This mostly means checking that Lc, Data, Le fields look ok.
//
// It returns an error when something is clearly wrong with the CAPDU
func (apdu *CAPDU) Test() error {
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

	dataLen, _ := apdu.GetLc()
	if int(dataLen) != len(apdu.Data) {
		return errors.New("APDU Lc value is different from the actual data length")
	}
	return nil
}

// Bytes provides the byte-slice value for a CAPDU, so it can be sent
// to the NFC device.
// It returns an error when something goes wrong (uses Test()).
func (apdu *CAPDU) Bytes() ([]byte, error) {
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

// RAPDU represents a Response APDU, which is received as an answer to
// Command APDUs. Response APDUs may contain data, along with two trailer
// bytes indicating the status.
type RAPDU struct {
	ResponseBody []byte //Data bytes
	SW1          byte   //Status Word 1
	SW2          byte   //Status Word 2
}

// ParseBytes parses a byte slice and sets the RAPDU fields accordingly.
// It returns the number of bytes parsed or an error if something goes wrong.
func (apdu *RAPDU) ParseBytes(bytes []byte) (int, error) {
	length := len(bytes)
	apdu.SW1 = bytes[length-2]
	apdu.SW2 = bytes[length-1]
	if length >= 3 {
		apdu.ResponseBody = bytes[0 : length-2]
	}
	return length, nil
}

// CommandCompleted checks if the RAPDU indicates a successful
// completion of a command.
func (apdu *RAPDU) CommandCompleted() bool {
	return apdu.SW1 == 0x90 && apdu.SW2 == 0x00
}

// FileNotFound checks if the RAPDU indicates that a file
// was not found (usually in response to a Select operation).
func (apdu *RAPDU) FileNotFound() bool {
	return apdu.SW1 == 0x6A && apdu.SW2 == 0x82
}

// NDEFTagApplicationSelectAPDU returns a new CAPDU
// which performs a Select operation by name with the NDEF
// Application Name.
func NDEFTagApplicationSelectAPDU() *CAPDU {
	cApdu := &CAPDU{
		CLA: byte(0x00),
		INS: byte(0xA4),
		P1:  byte(0x04), // Select by name
		P2:  byte(0x00), // First or only occurrence
		Data: []byte{
			0xD2,
			0x76,
			0x00,
			0x00,
			0x85,
			0x01,
			0x01}, // NDEF app name
	}
	cApdu.SetLc(7)
	// This would set a single-byte Le to 0, meaning response data
	// field might be present(and be up to 256 bytes according to Wikipedia)
	cApdu.SetLe(256)
	return cApdu
}

// CapabilityContainerSelectAPDU returns a new Select CAPDU to
// Select the Capabaility Container.
func CapabilityContainerSelectAPDU() *CAPDU {
	bytes := Uint16ToBytes(CCID)
	return SelectAPDU(bytes[:])
}

// ReadBinaryAPDU returns a new CAPDU to perform a binary
// read with the indicated offset and length.
func ReadBinaryAPDU(offset uint16, length uint16) *CAPDU {
	offsetBytes := Uint16ToBytes(offset)
	cApdu := &CAPDU{
		CLA: byte(0x00),
		INS: byte(0xB0),
		P1:  offsetBytes[0],
		P2:  offsetBytes[1],
	}
	cApdu.SetLe(length)
	return cApdu
}

// SelectAPDU returns a new CAPDU to perform a select
// operation by ID with the provided fileID
func SelectAPDU(fileID []byte) *CAPDU {
	cApdu := &CAPDU{
		CLA:  byte(0x00),
		INS:  byte(0xA4),
		P1:   byte(0x00), // Select by Id
		P2:   byte(0x0C), // First or only occurrence
		Data: fileID,
	}
	cApdu.SetLc(uint16(len(fileID))) //File ID length should be 2
	return cApdu
}

// BUG(hector): Capability Containers with more than 15 bytes (because
// they include optional TLV fields), will fail, as we only read
// 15 bytes and the CCLEN will not match the parsed data size.

// CapabilityContainerReadAPDU returns a new CAPDU to
// perform a binary read of 15 bytes with 0 offset (the
// regular size of a standard capability container).
func CapabilityContainerReadAPDU() *CAPDU {
	return ReadBinaryAPDU(0, 15)
}
