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

package capabilitycontainer

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/hsanjuan/go-nfctype4/helpers"
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
	T byte   // Type of the TLV block. 00h-FEh. 00h-03h and 06h-FFh RFU.
	L uint16 // Size of the value field.
	V []byte // Value field
}

// Reset sets the fields of the TLV to their default values.
func (tlv *TLV) Reset() {
	tlv.T = 0
	tlv.L = 0
	tlv.V = []byte{}
}

// Unmarshal parses a byte slice and sets the TLV struct fields accordingly.
// It always resets the TLV before parsing.
// It returns the number of bytes parsed or an error if the result does
// not look correct.
func (tlv *TLV) Unmarshal(buf []byte) (rLen int, err error) {
	defer helpers.HandleErrorPanic(&err, "TLV.Unmarshal")
	bytesBuf := bytes.NewBuffer(buf)
	tlv.Reset()

	tlv.T = helpers.GetByte(bytesBuf)
	if bytesBuf.Len() == 0 {
		// No length field. pff
		tlv.L = 0
		tlv.V = []byte{}
		return 1, nil
	}
	l0 := helpers.GetByte(bytesBuf)

	if l0 == 0xFF { // 3 byte format
		l1 := helpers.GetByte(bytesBuf)
		l2 := helpers.GetByte(bytesBuf)
		tlv.L = helpers.BytesToUint16([2]byte{l1, l2})
	} else {
		tlv.L = uint16(l0)
	}
	tlv.V = helpers.GetBytes(bytesBuf, int(tlv.L))

	rLen = len(buf) - bytesBuf.Len()
	if err := tlv.check(); err != nil {
		return rLen, err
	}
	if tlv.L < 0xFF && l0 == 0xFF {
		return rLen, errors.New("TLV.Unmarshal: " +
			"3-byte length used for a value < 0xFF")
	}

	return rLen, nil
}

// Marshal returns the byte slice representation of a TLV.
// It returns an error if the TLV breaks the spec.
func (tlv *TLV) Marshal() ([]byte, error) {
	if err := tlv.check(); err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	buffer.WriteByte(tlv.T)
	if tlv.L >= 0xFF { // 3 byte format
		buffer.WriteByte(0xFF)
		lBytes := helpers.Uint16ToBytes(tlv.L)
		buffer.Write(lBytes[:])
	} else { // 1 byte only
		buffer.WriteByte(byte(tlv.L))
	}
	buffer.Write(tlv.V)
	return buffer.Bytes(), nil
}

// Check performs some tests on a TLV to ensure that it follows the
// specification. These are mostly related to the L bytes being used
// correctly.
// It returns an error when something does not look right.
func (tlv *TLV) check() error {
	// This cannot be checked anymore with L as uint16
	// var vLen uint16
	// if tlv.L[0] == 0xFF {
	// 	vLen = helpers.BytesToUint16([2]byte{tlv.L[1], tlv.L[2]})
	// 	if vLen < 0xFF {
	// 		return errors.New(
	// 			"TLV.check: 3-byte Length's last 2 bytes " +
	// 				"value should > 0xFF")
	// 	}
	// 	if vLen == 0xFFFF {
	// 		return errors.New("TLV.check: 3-byte Length's last " +
	// 			"2 bytes value 0xFFFF is RFU")
	// 	}
	// } else {
	// 	vLen = uint16(tlv.L[0])
	// }

	if int(tlv.L) != len(tlv.V) {
		return errors.New(
			"TLV.check: L[ength] does not match the V[alue] length")
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
	FileID uint16
	// Size of the file containing the NDEF message
	MaximumFileSize          uint16
	FileReadAccessCondition  byte
	FileWriteAccessCondition byte
}

// NDEFFileControlTLV is a ControlTLV for a file containing a NDEF Message.
type NDEFFileControlTLV ControlTLV

// PropietaryFileControlTLV is a ControlTLV for a file containing some
// propietary format.
type PropietaryFileControlTLV ControlTLV

// Unmarshal parses a byte slice and sets the ControlTLV fields accordingly.
// It returns the number of bytes parsed or an error if the result does
// not look correct.
func (cTLV *ControlTLV) Unmarshal(buf []byte) (rLen int, err error) {
	// Parse it to a regular TLV
	tlv := new(TLV)
	rLen, err = tlv.Unmarshal(buf)
	if err != nil {
		return rLen, err
	}
	if rLen != 8 {
		return rLen, fmt.Errorf("ControlTLV: Wrong size %d", rLen)
	}

	cTLV.T = tlv.T
	cTLV.L = byte(tlv.L)
	cTLV.FileID = helpers.BytesToUint16([2]byte{tlv.V[0], tlv.V[1]})
	cTLV.MaximumFileSize = helpers.BytesToUint16([2]byte{tlv.V[2], tlv.V[3]})
	cTLV.FileReadAccessCondition = tlv.V[4]
	cTLV.FileWriteAccessCondition = tlv.V[5]

	if err := cTLV.check(); err != nil {
		return rLen, err
	}

	// Return that we parsed 8 bytes
	return rLen, nil
}

// Marshal returns the byte slice representation of a ControlTLV.
// It returns an error if the ControlTLV does not look correct..
func (cTLV *ControlTLV) Marshal() ([]byte, error) {
	// Test that this cTLV looks good
	if err := cTLV.check(); err != nil {
		return nil, err
	}

	// Copy this to a regular TLV and leverage Marshal() from there
	tlv := new(TLV)
	tlv.T = cTLV.T
	tlv.L = uint16(cTLV.L)
	var v bytes.Buffer
	fileID := helpers.Uint16ToBytes(cTLV.FileID)
	v.Write(fileID[:])
	mfs := helpers.Uint16ToBytes(cTLV.MaximumFileSize)
	v.Write(mfs[:])
	v.WriteByte(cTLV.FileReadAccessCondition)
	v.WriteByte(cTLV.FileWriteAccessCondition)
	tlv.V = v.Bytes()
	return tlv.Marshal()
}

// Check makes sure that the ControlTLV is not breaking the specification
// by checking its fields' values are acceptable. If not, it returns an error.
//
// ControlTLV have a number of Rerserved values for FileIDs and
// access conditions which should not be used.
func (cTLV *ControlTLV) check() error {
	switch cTLV.FileID {
	case 0x000, 0xe102, 0xe103, 0x3f00, 0x3fff:
		return errors.New(
			"ControlTLV.check: File ID is reserved by ISO/IEC_7816-4")

	case 0xffff:
		return errors.New("ControlTLV.check: File ID is invalid (RFU)")
	}

	if 0x0000 <= cTLV.MaximumFileSize && cTLV.MaximumFileSize <= 0x0004 {
		return errors.New(
			"ControlTLV.check: Maximum File Size value is RFU")
	}

	if 0x01 <= cTLV.FileReadAccessCondition && cTLV.FileReadAccessCondition <= 0x7f {
		return errors.New(
			"ControlTLV.check: Read Access Condition has RFU value")
	}

	if 0x01 <= cTLV.FileWriteAccessCondition && cTLV.FileWriteAccessCondition <= 0x7f {
		return errors.New(
			"ControlTLV.check: Write Access Condition has RFU value")
	}
	return nil
}

// Unmarshal parses a byte slice and sets the NDEFFileControlTLV fields
// accordingly.
// It returns the number of bytes parsed or an error if the result does
// not follow the specification.
func (nfcTLV *NDEFFileControlTLV) Unmarshal(buf []byte) (rLen int, err error) {
	// Reuse functions
	tlv := (*ControlTLV)(nfcTLV)
	rLen, err = tlv.Unmarshal(buf)
	if err != nil {
		return rLen, err
	}

	if !tlv.IsNDEFFileControlTLV() {
		return rLen, errors.New("NDEFFileControlTLV.Unmarshal: " +
			"TLV is not a NDEF File Control TLV")
	}

	return rLen, nil
}

// Marshal returns the byte slice representation of a NDEFFileControlTLV.
// It returns an error if the underlying ControlTLV does not follow the
// specification.
func (nfcTLV *NDEFFileControlTLV) Marshal() ([]byte, error) {
	tlv := (*ControlTLV)(nfcTLV)
	return tlv.Marshal()
}

// Unmarshal parses a byte slice and sets the PropietaryFileControlTLV fields
// accordingly.
// It returns the number of bytes parsed or an error if the result does
// not follow the specification.
func (pfcTLV *PropietaryFileControlTLV) Unmarshal(buf []byte) (rLen int, err error) {
	// Reuse functions
	tlv := (*ControlTLV)(pfcTLV)
	rLen, err = tlv.Unmarshal(buf)
	if err != nil {
		return rLen, err
	}

	if !tlv.IsPropietaryFileControlTLV() {
		return rLen, errors.New(
			"PropietaryFileControlTLV.Unmarshal:" +
				"TLV is not a Propietary File Control TLV")
	}

	return rLen, nil
}

// Marshal returns the byte slice representation of a PropietaryFileControlTLV.
// It returns an error if the underlying ControlTLV does not follow the
// specification.
func (pfcTLV *PropietaryFileControlTLV) Marshal() ([]byte, error) {
	tlv := (*ControlTLV)(pfcTLV)
	return tlv.Marshal()
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
