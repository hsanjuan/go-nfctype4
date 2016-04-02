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

// Defined TVL blocks
const (
	NDEF_FILE_CONTROL_TLV       = byte(0x04)
	PROPIETARY_FILE_CONTROL_TLV = byte(0x05)
)

type TLV struct {
	T byte    // Type of the TLV block. 00h-FEh. 00h-03h and 06h-FFh RFU.
	L [3]byte // Size of the value field. Two last bytes may be unused
	V []byte  // Value field
}

// Parse some bytes into a TLV struct. Returns the number of bytes parsed or an error
func (tlv *TLV) ParseBytes(bytes []byte) (int, error) {
	if len(bytes) == 0 {
		return 0, errors.New("Need at least 1 byte to parse a TLV")
	}
	tlv.T = bytes[0]
	if len(bytes) == 1 {
		// No length field. pff
		tlv.L = [3]byte{0, 0, 0}
		tlv.V = []byte{}
		return 1, nil
	}
	tlv.L[0] = bytes[1]
	if len(bytes) < 2+int(tlv.L[0]) { // At least
		return 0, errors.New("TLV Size malformed")
	}
	v_size := uint16(0)
	var parsed int
	if tlv.L[0] == 0xFF { // 3 byte format
		tlv.L[1] = bytes[2]
		tlv.L[2] = bytes[3]
		v_size = BytesToUint16([2]byte{tlv.L[1], tlv.L[2]})
		if len(bytes) < 4+int(v_size) {
			return 0, errors.New("Not enough bytes to parse!")
		}
		tlv.V = bytes[4 : 4+int(v_size)]
		parsed = 4 + len(tlv.V)
	} else {
		v_size = uint16(tlv.L[0])
		if len(bytes) < 2+int(v_size) {
			return 0, errors.New("Not enough bytes to parse!")
		}
		tlv.V = bytes[2 : 2+int(v_size)]
		parsed = 2 + len(tlv.V)
	}

	// Test just in case
	if err := tlv.Test(); err != nil {
		return 0, err
	}

	return parsed, nil
}

// Convert a TLV to []byte. Returns the result, or an error
func (tlv *TLV) Bytes() ([]byte, error) {
	if err := tlv.Test(); err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	buffer.WriteByte(tlv.T)
	if tlv.L[0] == 0xFF { // 3 byte format
		buffer.Write(tlv.L[:])
	} else { // 1 byte only
		buffer.WriteByte(tlv.L[0])
	}
	buffer.Write(tlv.V)
	return buffer.Bytes(), nil
}

// Test that it follows the standard
func (tlv *TLV) Test() error {
	if tlv.T != 0x04 && tlv.T != 0x05 {
		return errors.New("TLV T[ype] is RFU")
	}

	var v_size uint16
	if tlv.L[0] == 0xFF {
		v_size = BytesToUint16([2]byte{tlv.L[1], tlv.L[2]})
		if v_size < 0xFF {
			return errors.New("TLV 3-byte Length's last 2 bytes value should > 0xFF")
		}
		if v_size == 0xFFFF {
			return errors.New("TLV 3-byte Length's last 2 bytes value 0xFFFF is RFU")
		}
	} else {
		v_size = uint16(tlv.L[0])
	}
	if int(v_size) != len(tlv.V) {
		return errors.New("TLV L[ength] does not match the V[alue] length")
	}
	return nil
}

/////////////////////////////////////////////////////////////////////////

type ControlTLV struct {
	T                        byte    // Should always be 04h
	L                        byte    // Size of the value field. Should always be 06h
	FileID                   [2]byte // A valid File ID: 0001h-E101h, E104h-3EFFh, 3F01h-3FFEh, 4000h-FFFEh.
	MaximumFileSize          [2]byte
	FileReadAccessCondition  byte
	FileWriteAccessCondition byte
}

type NDEFFileControlTLV ControlTLV
type PropietaryFileControlTLV ControlTLV

func (c_tlv *ControlTLV) ParseBytes(bytes []byte) (int, error) {
	// Parse it to a regular TLV
	tlv := new(TLV)
	parsed, err := tlv.ParseBytes(bytes)
	if err != nil {
		return 0, err
	}
	if parsed != 8 {
		return 0, errors.New("Wrong size for a Control TLV")
	}

	c_tlv.T = tlv.T
	c_tlv.L = tlv.L[0]
	c_tlv.FileID[0] = tlv.V[0]
	c_tlv.FileID[1] = tlv.V[1]
	c_tlv.MaximumFileSize[0] = tlv.V[2]
	c_tlv.MaximumFileSize[1] = tlv.V[3]
	c_tlv.FileReadAccessCondition = tlv.V[4]
	c_tlv.FileWriteAccessCondition = tlv.V[5]

	if err := c_tlv.Test(); err != nil {
		return 0, err
	}

	// Return that we parsed 8 bytes
	return 8, nil
}

func (c_tlv *ControlTLV) Bytes() ([]byte, error) {
	// Test that this c_tlv looks good
	if err := c_tlv.Test(); err != nil {
		return nil, err
	}

	// Copy this to a regular TLV and leverage Bytes() from there
	tlv := new(TLV)
	tlv.T = c_tlv.T
	tlv.L[0] = c_tlv.L
	var v bytes.Buffer
	v.Write(c_tlv.FileID[:])
	v.Write(c_tlv.MaximumFileSize[:])
	v.WriteByte(c_tlv.FileReadAccessCondition)
	v.WriteByte(c_tlv.FileWriteAccessCondition)
	tlv.V = v.Bytes()
	return tlv.Bytes()
}

func (c_tlv *ControlTLV) Test() error {
	file_id := BytesToUint16(c_tlv.FileID)
	switch file_id {
	case 0x000, 0xe102, 0xe103, 0x3f00, 0x3fff:
		return errors.New("TLV: File ID is reserved by ISO/IEC_7816-4")

	case 0xffff:
		return errors.New("TLV: File ID is invalid (RFU)")
	}

	max_size := BytesToUint16(c_tlv.MaximumFileSize)
	if 0x0000 <= max_size && max_size <= 0x0004 {
		return errors.New("TLV: Maximum File Size value is RFU")
	}

	if 0x01 <= c_tlv.FileReadAccessCondition && c_tlv.FileReadAccessCondition <= 0x7f {
		return errors.New("TLV: Read Access Condition has RFU value")
	}

	if 0x01 <= c_tlv.FileWriteAccessCondition && c_tlv.FileWriteAccessCondition <= 0x7f {
		return errors.New("TLV: Write Access Condition has RFU value")
	}
	return nil
}

func (nfc_tlv *NDEFFileControlTLV) ParseBytes(bytes []byte) (int, error) {
	// Reuse functions
	tlv := (*ControlTLV)(nfc_tlv)
	parsed, err := tlv.ParseBytes(bytes)
	if err != nil {
		return parsed, err
	}

	if !tlv.IsNDEFFileControlTLV() {
		return parsed, errors.New("TLV is not a NDEF File Control TLV")
	}

	return parsed, nil
}

func (nfc_tlv *NDEFFileControlTLV) Bytes() ([]byte, error) {
	tlv := (*ControlTLV)(nfc_tlv)
	return tlv.Bytes()
}

func (pfc_tlv *PropietaryFileControlTLV) ParseBytes(bytes []byte) (int, error) {
	// Reuse functions
	tlv := (*ControlTLV)(pfc_tlv)
	parsed, err := tlv.ParseBytes(bytes)
	if err != nil {
		return parsed, err
	}

	if !tlv.IsPropietaryFileControlTLV() {
		return parsed, errors.New("TLV is not a Propietary File Control TLV")
	}

	return parsed, nil
}

func (pfc_tlv *PropietaryFileControlTLV) Bytes() ([]byte, error) {
	tlv := (*ControlTLV)(pfc_tlv)
	return tlv.Bytes()
}

func (tlv *ControlTLV) IsNDEFFileControlTLV() bool {
	return tlv.T == NDEF_FILE_CONTROL_TLV
}

func (tlv *ControlTLV) IsPropietaryFileControlTLV() bool {
	return tlv.T == PROPIETARY_FILE_CONTROL_TLV
}

func (tlv *ControlTLV) IsFileReadable() bool {
	return tlv.FileReadAccessCondition == 0x00
}

func (tlv *ControlTLV) IsFileWriteable() bool {
	return tlv.FileWriteAccessCondition == 0x00
}

func (tlv *ControlTLV) IsFileReadOnly() bool {
	return tlv.FileWriteAccessCondition == 0xFF && tlv.IsFileReadable()
}
