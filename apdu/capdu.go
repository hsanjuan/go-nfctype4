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

package apdu

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hsanjuan/nfctype4/helpers"
)

// CAPDU.INS relevant to the Type 4 Tag Specification
const (
	INSSelect = byte(0xA4)
	INSRead   = byte(0xB0)
	INSUpdate = byte(0xD6)
)

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

// Reset clears the fields of the CAPDU to their default values.
func (apdu *CAPDU) Reset() {
	apdu.CLA = 0
	apdu.INS = 0
	apdu.P1 = 0
	apdu.P2 = 0
	apdu.Lc = []byte{}
	apdu.Data = []byte{}
	apdu.Le = []byte{}
}

// String provides a readable representation of the CAPDU
func (apdu *CAPDU) String() string {
	str := ""
	str += fmt.Sprintf("CLA: %02x | INS: %02x | P1: %02x | P2: %02x",
		apdu.CLA, apdu.INS, apdu.P1, apdu.P2)
	str += " | Lc: "
	for _, b := range apdu.Lc {
		str += fmt.Sprintf("%02x", b)
	}
	str += " | Data: "
	for _, b := range apdu.Data {
		str += fmt.Sprintf("%02x", b)
	}
	str += " | Le: "
	for _, b := range apdu.Le {
		str += fmt.Sprintf("%02x", b)
	}
	return str
}

// GetLc computes the actual Lc value from the Lc bytes. Lc
// indicates the length of the data sent with a Command
// APDU and goes from 0 to 2^16-1.
// Note this method will return
// 0 if it cannot make sense of the Lc bytes.
func (apdu *CAPDU) GetLc() uint16 {
	switch len(apdu.Lc) {
	case 0:
		return uint16(0) // This goes against spec
	case 1:
		return uint16(apdu.Lc[0])
	case 3:
		return helpers.BytesToUint16([2]byte{apdu.Lc[1], apdu.Lc[2]})
	default:
		return 0
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
		nBytes := helpers.Uint16ToBytes(n)
		apdu.Lc = []byte{0x00, nBytes[0], nBytes[1]}
	}
}

// BUG(hector): APDU's Le field could theoretically be 65536 (2^16), but
// this overflows uint16 so it's unsupported by SetLe and GetLe.
// It only happens in the case when Le has two bytes and both are 0 and in this
// case GetLe returns 2^16 -1.

// GetLe computes the actual Le value from the Le bytes. Le
// indicates the maximum length of the data to be received Command
// APDU and goes from 0 to 2^16. Note this method will return
// 0 if it cannot make sense of the Le bytes.
func (apdu *CAPDU) GetLe() uint16 {
	switch len(apdu.Le) {
	case 0:
		return uint16(0)
	case 1:
		n := apdu.Le[0]
		if n == 0 {
			return uint16(256)
		}
		return uint16(n)
	case 2:
		n0 := apdu.Le[0]
		n1 := apdu.Le[1]
		if n0 == 0 && n1 == 0 {
			//return uint16(65536) // Overflow! FIXME!
			return uint16(65535)
		}
		return helpers.BytesToUint16([2]byte{n0, n1})
	case 3:
		return helpers.BytesToUint16([2]byte{apdu.Le[1], apdu.Le[2]})
	default:
		return 0
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
		nBytes := helpers.Uint16ToBytes(n)
		if len(apdu.Lc) > 0 { // Make it 2 bytes
			apdu.Le = []byte{nBytes[0], nBytes[1]}
		} else { // 3 bytes then
			apdu.Le = []byte{0, nBytes[0], nBytes[1]}
		}
	}
}

// Check ensures that a CAPDU struct fields are in-line with the
// specification.
// This mostly means checking that Lc, Data, Le fields look ok.
//
// It returns an error when something is clearly wrong with the CAPDU
func (apdu *CAPDU) check() error {
	lc := apdu.Lc
	le := apdu.Le
	// Test Lc
	switch {
	case len(lc) == 1 && lc[0] == 0:
		return errors.New(
			"CAPDU.Check: APDU Lc with 1 byte cannot be 0")
	case len(lc) == 2:
		return errors.New("CAPDU.Check: APDU Lc cannot have 2 bytes")
	case len(lc) == 3:
		if lc[0] != 0 {
			return errors.New("CAPDU.Check: " +
				"APDU 3-byte-Lc's first byte must be 0")
		} else if lc[1] == 0 && lc[2] == 0 {
			return errors.New("CAPDU.Check: " +
				"APDU 3-byte-Lc cannot be all 0s")
		}
	case len(lc) > 3:
		return errors.New(
			"CAPDU.Check: APDU Lc cannot have more than 3")
	}

	// Test Le
	switch {
	case len(le) == 2 && len(lc) == 0:
		return errors.New(
			"CAPDU.Check: APDU 2-byte-Le needs Lc present")
	case len(le) == 3:
		if len(lc) != 0 {
			return errors.New("CAPDU.Check: " +
				"APDU 3-byte-Le is only " +
				"compatible with empty Lc")
		} else if le[0] != 0 {
			return errors.New("CAPDU.Check: " +
				"APDU 3-byte-Le's first byte must be 0")
		}
	case len(le) > 3:
		return errors.New("CAPDU.Check: " +
			"APDU Le cannot have more 3 bytes")
	}

	if int(apdu.GetLc()) != len(apdu.Data) {
		return errors.New("CAPDU.Check: " +
			"APDU Lc value is differs from the actual data length")
	}
	return nil
}

// Unmarshal parses a byte slice and sets the CAPDU fields accordingly.
// It resets the CAPDU structure before parsing.
// It returns the number of bytes parsed or an error if something goes wrong.
func (apdu *CAPDU) Unmarshal(buf []byte) (int, error) {
	if len(buf) < 4 {
		return 0, errors.New("CAPDU.Unmarshal: not enough bytes")
	}
	apdu.Reset()
	i := 0
	apdu.CLA = buf[i]
	i++
	apdu.INS = buf[i]
	i++
	apdu.P1 = buf[i]
	i++
	apdu.P2 = buf[i]
	i++

	// See table 5 here about what's going on
	// http://www.cardwerk.com/smartcards/smartcard_standard_ISO7816-4_5_basic_organizations.aspx
	// I have copied the comments for each case, but in some there are
	// clear typos. Also, contradicts the Wikipedia as how 3-byte values are
	// coded ("0x00->1 and 0xFF->65536" instead of
	// "0x00->65536 and 0x01->1...""

	// We chose to follow Wikipedia on this.
	bodyLen := len(buf) - i
	b1 := byte(0)
	b2 := byte(0)
	b3 := byte(0)
	if bodyLen > 0 {
		b1 = buf[i]
	}
	if bodyLen > 1 {
		b2 = buf[i+1]
	}
	if bodyLen > 2 {
		b3 = buf[i+2]
	}
	switch {
	case bodyLen == 0:
		// Case 1 - L=0 : the body is empty.
		// No byte is used for Lc valued to 0
		// No data byte is present.
		// No byte is used for Le valued to 0.
		// Nothing to do here
	case bodyLen == 1:
		// Case 2S - L=1
		// No byte is used for Lc valued to 0
		// No data byte is present.
		// B1 codes Le valued from 1 to 256
		apdu.Le = []byte{b1}
	case bodyLen == (1+int(b1)) && b1 != 0:
		// Case 3S - L=1 + (B1) and (B1) != 0
		// B1 codes Lc (=0) valued from 1 to 255
		// B2 to Bl are the Lc bytes of the data field
		// No byte is used for Le valued to 0.
		apdu.Lc = []byte{b1}
		i++
		apdu.Data = buf[i : i+int(b1)]
		i += int(b1)
	case bodyLen == (2+int(b1)) && b1 != 0:
		// Case 4S - L=2 + (B1) and (B1) != 0
		// B1 codes Lc (!=0) valued from 1 to 255
		// B2 to Bl-1 are the Lc bytes of the data field
		// Bl codes Le from 1 to 256
		apdu.Lc = []byte{b1}
		i++
		apdu.Data = buf[i : i+int(b1)]
		i += int(b1)
		apdu.Le = []byte{buf[i]}
		i++
	case bodyLen == 3 && b1 == 0:
		// Case 2E - L=3 and (B1)=0
		// No byte is used for Lc valued to 0
		// No data bytes is present
		// The Le field consists of the 3 bytes where B2 and
		// B3 code Le valued from 1 to 65536
		apdu.Le = []byte{b1, b2, b3}
		i += 3
	case bodyLen == (3+int(helpers.BytesToUint16([2]byte{b2, b3}))) && b1 == 0 && (b2|b3) != 0:
		// Case 3E - L=3 + (B2||B3). (B1)=0 and (B2||B3)=0
		// It should say: L=3 + (B2||B3). (B1)=0 and (B2||B3)!=0
		// The Lc field consists of the first 3 bytes where B2 and B3 code Lc (!=0) valued from 1 to 65536
		// B4 and B2 are the Lc bytes of the data field
		// No byte is used for Le valued to 0
		apdu.Lc = []byte{b1, b2, b3}
		i += 3
		apdu.Data = buf[i : i+int(apdu.GetLc())]
		i += int(apdu.GetLc())
	case bodyLen == (5+int(helpers.BytesToUint16([2]byte{b2, b3}))) && b1 == 0 && (b2|b3) != 0:
		//Case 4E - L= 5 + (B2||B3),(B1)=0 and (B2||B3)=0
		// The Lc field consists of the first 3 bytes where B2 and B3 code Lc (!=0) valued from 1 to 65535
		//B4 to Bl-2 are the Lc bytes of the data field
		//The Le field consists of the last 2 bytes Bl-1 and Bl which code Le valued from 1 to 65536
		apdu.Lc = []byte{b1, b2, b3}
		i += 3
		apdu.Data = buf[i : i+int(apdu.GetLc())]
		i += int(apdu.GetLc())
		apdu.Le = []byte{buf[i], buf[i+1]}
		i += 2
	}

	if err := apdu.check(); err != nil {
		return i, err
	}
	return i, nil
}

// Marshal provides the byte-slice value for a CAPDU, so it can be sent
// to the NFC device.
// It returns an error when something goes wrong (uses Test()).
func (apdu *CAPDU) Marshal() ([]byte, error) {
	if err := apdu.check(); err != nil {
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

// NewNDEFTagApplicationSelectAPDU returns a new CAPDU
// which performs a Select operation by name with the NDEF
// Application Name.
func NewNDEFTagApplicationSelectAPDU() *CAPDU {
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

// NewReadBinaryAPDU returns a new CAPDU to perform a binary
// read with the indicated offset and length.
func NewReadBinaryAPDU(offset uint16, length uint16) *CAPDU {
	offsetBytes := helpers.Uint16ToBytes(offset)
	cApdu := &CAPDU{
		CLA: byte(0x00),
		INS: byte(0xB0),
		P1:  offsetBytes[0],
		P2:  offsetBytes[1],
	}
	cApdu.SetLe(length)
	return cApdu
}

// NewSelectAPDU returns a new CAPDU to perform a select
// operation by ID with the provided fileID
func NewSelectAPDU(fileID uint16) *CAPDU {
	dataBuf := helpers.Uint16ToBytes(fileID)
	cApdu := &CAPDU{
		CLA:  byte(0x00),
		INS:  byte(0xA4),
		P1:   byte(0x00), // Select by Id
		P2:   byte(0x0C), // First or only occurrence
		Data: dataBuf[:],
	}
	cApdu.SetLc(2) //File ID length should be 2
	return cApdu
}

// BUG(hector): Capability Containers with more than 15 bytes (because
// they include optional TLV fields), will fail, as we only read
// 15 bytes and the CCLEN will not match the parsed data size.

// NewCapabilityContainerReadAPDU returns a new CAPDU to
// perform a binary read of 15 bytes with 0 offset (the
// regular size of a standard capability container).
func NewCapabilityContainerReadAPDU() *CAPDU {
	return NewReadBinaryAPDU(0, 15)
}
