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

// Values allowed for the T fields of TLV Blocks.
const (
	TypeNDEFFileControlTLV       = byte(0x04)
	TypePropietaryFileControlTLV = byte(0x05)
)

// TLV represents a plain TLV block which is just a container for some data.
//
// TLV Blocks have a L field which indicates the length of the V field. This
// field can be of 1 or 3 bytes. For the shorter version, the last 2 bytes of
// the array are left unused.
type TLV struct {
	T byte    // Type of the TLV block. 00h-FEh. 00h-03h and 06h-FFh RFU.
	L [3]byte // Size of the value field. Two last bytes may be unused
	V []byte  // Value field
}

// ParseBytes parses a byte slice and sets the TLV struct fields accordingly.
// It returns the number of bytes parsed or an error if the result does
// not look correct (it uses Test()).
func (tlv *TLV) ParseBytes(bytes []byte) (int, error) {
	if len(bytes) == 0 {
		return 0, errors.New(
			"TLV.ParseBytes: need at least 1 byte to parse a TLV")
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
		return 0, errors.New("TLV.ParseBytes: TLV.L field is malformed")
	}
	vLen := uint16(0)
	var parsed int
	if tlv.L[0] == 0xFF { // 3 byte format
		tlv.L[1] = bytes[2]
		tlv.L[2] = bytes[3]
		vLen = bytesToUint16([2]byte{tlv.L[1], tlv.L[2]})
		if len(bytes) < 4+int(vLen) {
			return 0, errors.New(
				"TLV.ParseBytes: not enough bytes to parse")
		}
		tlv.V = bytes[4 : 4+int(vLen)]
		parsed = 4 + len(tlv.V)
	} else {
		vLen = uint16(tlv.L[0])
		if len(bytes) < 2+int(vLen) {
			return 0, errors.New(
				"TLV.ParseBytes: not enough bytes to parse")
		}
		tlv.V = bytes[2 : 2+int(vLen)]
		parsed = 2 + len(tlv.V)
	}

	// Test just in case
	if err := tlv.Test(); err != nil {
		return 0, err
	}

	return parsed, nil
}

// Bytes returns the byte slice representation of a TLV.
// It returns an error if the TLV breaks the spec (uses Test()).
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

// Test performs some tests on a TLV to ensure that it follows the
// specification. These are mostly related to the L bytes being used
// correctly.
// It returns an error when something does not look right.
func (tlv *TLV) Test() error {
	if tlv.T != 0x04 && tlv.T != 0x05 {
		return errors.New("TLV T[ype] is RFU")
	}

	var vLen uint16
	if tlv.L[0] == 0xFF {
		vLen = bytesToUint16([2]byte{tlv.L[1], tlv.L[2]})
		if vLen < 0xFF {
			return errors.New(
				"TLV.Test: 3-byte Length's last 2 bytes " +
					"value should > 0xFF")
		}
		if vLen == 0xFFFF {
			return errors.New("TLV.Test: 3-byte Length's last " +
				"2 bytes value 0xFFFF is RFU")
		}
	} else {
		vLen = uint16(tlv.L[0])
	}
	if int(vLen) != len(tlv.V) {
		return errors.New(
			"TLV.Test: L[ength] does not match the V[alue] length")
	}
	return nil
}

/////////////////////////////////////////////////////////////////////////

// ControlTLV is a specialized version of a TLV with a fixed size and a
// fixed format for the V field. The V field is used to indicate a
// fileID, its maximum size and read/write access flags.
type ControlTLV struct {
	T byte // Should always be 04h
	L byte // Size of the value field. Always 06h.
	// A valid File ID: 0001h-E101h, E104h-3EFFh, 3F01h-3FFEh, 4000h-FFFEh.
	FileID [2]byte
	// Size of the file containing the NDEF message
	MaximumFileSize          [2]byte
	FileReadAccessCondition  byte
	FileWriteAccessCondition byte
}

// NDEFFileControlTLV is a ControlTLV for a file containing a NDEF Message.
type NDEFFileControlTLV ControlTLV

// PropietaryFileControlTLV is a ControlTLV for a file containing some
// propietary format.
type PropietaryFileControlTLV ControlTLV

// ParseBytes parses a byte slice and sets the ControlTLV fields accordingly.
// It returns the number of bytes parsed or an error if the result does
// not look correct (it uses Test()).
func (cTLV *ControlTLV) ParseBytes(bytes []byte) (int, error) {
	// Parse it to a regular TLV
	tlv := new(TLV)
	parsed, err := tlv.ParseBytes(bytes)
	if err != nil {
		return 0, err
	}
	if parsed != 8 {
		return 0, errors.New("ControlTLV: Wrong size")
	}

	cTLV.T = tlv.T
	cTLV.L = tlv.L[0]
	cTLV.FileID[0] = tlv.V[0]
	cTLV.FileID[1] = tlv.V[1]
	cTLV.MaximumFileSize[0] = tlv.V[2]
	cTLV.MaximumFileSize[1] = tlv.V[3]
	cTLV.FileReadAccessCondition = tlv.V[4]
	cTLV.FileWriteAccessCondition = tlv.V[5]

	if err := cTLV.Test(); err != nil {
		return 0, err
	}

	// Return that we parsed 8 bytes
	return 8, nil
}

// Bytes returns the byte slice representation of a ControlTLV.
// It returns an error if the ControlTLV does not look correct (it uses
// Test()).
func (cTLV *ControlTLV) Bytes() ([]byte, error) {
	// Test that this cTLV looks good
	if err := cTLV.Test(); err != nil {
		return nil, err
	}

	// Copy this to a regular TLV and leverage Bytes() from there
	tlv := new(TLV)
	tlv.T = cTLV.T
	tlv.L[0] = cTLV.L
	var v bytes.Buffer
	v.Write(cTLV.FileID[:])
	v.Write(cTLV.MaximumFileSize[:])
	v.WriteByte(cTLV.FileReadAccessCondition)
	v.WriteByte(cTLV.FileWriteAccessCondition)
	tlv.V = v.Bytes()
	return tlv.Bytes()
}

// Test makes sure that the ControlTLV is not breaking the specification
// by checking its fields' values are acceptable. If not, it returns an error.
//
// ControlTLV have a number of Rerserved values for FileIDs and
// access conditions which should not be used.
func (cTLV *ControlTLV) Test() error {
	fileID := bytesToUint16(cTLV.FileID)
	switch fileID {
	case 0x000, 0xe102, 0xe103, 0x3f00, 0x3fff:
		return errors.New(
			"ControlTLV.Test: File ID is reserved by ISO/IEC_7816-4")

	case 0xffff:
		return errors.New("ControlTLV.Test: File ID is invalid (RFU)")
	}

	maxLen := bytesToUint16(cTLV.MaximumFileSize)
	if 0x0000 <= maxLen && maxLen <= 0x0004 {
		return errors.New(
			"ControlTLV.Test: Maximum File Size value is RFU")
	}

	if 0x01 <= cTLV.FileReadAccessCondition && cTLV.FileReadAccessCondition <= 0x7f {
		return errors.New(
			"ControlTLV.Test: Read Access Condition has RFU value")
	}

	if 0x01 <= cTLV.FileWriteAccessCondition && cTLV.FileWriteAccessCondition <= 0x7f {
		return errors.New(
			"ControlTLV.Test: Write Access Condition has RFU value")
	}
	return nil
}

// ParseBytes parses a byte slice and sets the NDEFFileControlTLV fields
// accordingly.
// It returns the number of bytes parsed or an error if the result does
// not look correct (it uses ControlTLV's Test() and checks the value
// of the T field is good.
func (nfcTLV *NDEFFileControlTLV) ParseBytes(bytes []byte) (int, error) {
	// Reuse functions
	tlv := (*ControlTLV)(nfcTLV)
	parsed, err := tlv.ParseBytes(bytes)
	if err != nil {
		return parsed, err
	}

	if !tlv.IsNDEFFileControlTLV() {
		return parsed, errors.New("NDEFFileControlTLV.ParseBytes: " +
			"TLV is not a NDEF File Control TLV")
	}

	return parsed, nil
}

// Bytes returns the byte slice representation of a NDEFFileControlTLV.
// It returns an error if the underlying ControlTLV does not look correct
// (it uses Test()).
func (nfcTLV *NDEFFileControlTLV) Bytes() ([]byte, error) {
	tlv := (*ControlTLV)(nfcTLV)
	return tlv.Bytes()
}

// ParseBytes parses a byte slice and sets the PropietaryFileControlTLV fields
// accordingly.
// It returns the number of bytes parsed or an error if the result does
// not look correct (it uses ControlTLV's Test() and checks the value
// of the T field is good.
func (pfcTLV *PropietaryFileControlTLV) ParseBytes(bytes []byte) (int, error) {
	// Reuse functions
	tlv := (*ControlTLV)(pfcTLV)
	parsed, err := tlv.ParseBytes(bytes)
	if err != nil {
		return parsed, err
	}

	if !tlv.IsPropietaryFileControlTLV() {
		return parsed, errors.New(
			"PropietaryFileControlTLV.ParseBytes:" +
				"TLV is not a Propietary File Control TLV")
	}

	return parsed, nil
}

// Bytes returns the byte slice representation of a PropietaryFileControlTLV.
// It returns an error if the underlying ControlTLV does not look correct
// (it uses Test()).
func (pfcTLV *PropietaryFileControlTLV) Bytes() ([]byte, error) {
	tlv := (*ControlTLV)(pfcTLV)
	return tlv.Bytes()
}

// IsNDEFFileControlTLV returns true if the T field has the right value.
func (cTLV *ControlTLV) IsNDEFFileControlTLV() bool {
	return cTLV.T == TypeNDEFFileControlTLV
}

// IsPropietaryFileControlTLV returns true if the T field has the right value.
func (cTLV *ControlTLV) IsPropietaryFileControlTLV() bool {
	return cTLV.T == TypePropietaryFileControlTLV
}

// IsFileReadable returns true when the ReadAccessCondition field indicates
// that the ControlTLV file is readable.
func (cTLV *ControlTLV) IsFileReadable() bool {
	return cTLV.FileReadAccessCondition == 0x00
}

// IsFileWriteable returns true when the ReadAccessCondition field indicates
// that the ControlTLV file is writeable.
func (cTLV *ControlTLV) IsFileWriteable() bool {
	return cTLV.FileWriteAccessCondition == 0x00
}

// IsFileReadOnly returns true when the ReadAccessCondition field indicates
// that the ControlTLV file is read-only.
func (cTLV *ControlTLV) IsFileReadOnly() bool {
	return cTLV.FileWriteAccessCondition == 0xFF && cTLV.IsFileReadable()
}
