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

// Package capabilitycontainer provides support for Capability Containers
// and TLV Blocks.
package capabilitycontainer

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/hsanjuan/go-nfctype4/helpers"
)

// CCID is the Capability container ID.
const CCID = uint16(0xE103)

// CapabilityContainer represents a Capability Container File as defined in the
// section 5.1 of the specification. The main function of the capability
// container file is to store the NDEFFileControlTLV (see docs for that struct)
// along with some maximum data length boundaries for reading and writing
// (MLe and MLc).
//
// The CapabilityContainer also indicates which version of the specification
// is the Tag compatible with.
type CapabilityContainer struct {
	CCLEN              uint16              // Size of this capability container - 000Fh to FFFEh
	MappingVersion     byte                // Major-Minor version (4 bits each)
	MLe                uint16              // Maximum data read with ReadBinary. 000Fh-FFFFh
	MLc                uint16              // Maximum data to write with UpdateBinary. 0001h-FFFFh
	NDEFFileControlTLV *NDEFFileControlTLV // NDEF file information
	TLVBlocks          []*ControlTLV       // Optional TLVs
}

// Reset clears all the fields of the CapabilityContainer to their
// default values.
func (cc *CapabilityContainer) Reset() {
	cc.CCLEN = 0
	cc.MappingVersion = 0
	cc.MLe = 0
	cc.MLc = 0
	cc.NDEFFileControlTLV = nil
	cc.TLVBlocks = nil
}

// Unmarshal parses a byte slice and sets the CapabilityContainer fields
// correctly. This involves parsing the NDEFFileControl TLV and any
// optional TLV fields if present. It always resets the CapabilityContainer
// before parsing.
//
// It returns the number of bytes read and an error if something looks wrong
// (it uses check() to check for the integrity of the result).
func (cc *CapabilityContainer) Unmarshal(buf []byte) (rLen int, err error) {
	defer helpers.HandleErrorPanic(&err, "RAPDU.Unmarshal")
	bytesBuf := bytes.NewBuffer(buf)
	cc.Reset()

	if len(buf) < 15 {
		return 0, errors.New(
			"CapabilityContainer.Unmarshal: " +
				"not enough bytes to parse")
	}
	i := 0
	cc.CCLEN = helpers.BytesToUint16([2]byte{
		helpers.GetByte(bytesBuf),
		helpers.GetByte(bytesBuf)})
	cc.MappingVersion = helpers.GetByte(bytesBuf)
	cc.MLe = helpers.BytesToUint16([2]byte{
		helpers.GetByte(bytesBuf),
		helpers.GetByte(bytesBuf)})
	cc.MLc = helpers.BytesToUint16([2]byte{
		helpers.GetByte(bytesBuf),
		helpers.GetByte(bytesBuf)})
	i += 7

	fcTLV := new(NDEFFileControlTLV)
	parsed, err := fcTLV.Unmarshal(helpers.GetBytes(bytesBuf, 8))
	if err != nil {
		return len(buf) - bytesBuf.Len(), err
	}
	cc.NDEFFileControlTLV = fcTLV
	i += parsed

	tlvBytes := bytesBuf.Bytes()
	rLen = len(buf) - len(tlvBytes)
	for rLen < int(cc.CCLEN) {
		// First parse a regular TLV so we can look at its type
		extraTLV := new(TLV)
		parsed, err = extraTLV.Unmarshal(buf[rLen:])
		if err != nil {
			rLen += parsed
			return rLen, err
		}
		// The Specs say: NFC Forum Devices shall ignore and
		// jump over those TLV blocks that make use
		// of reserved tag field values.
		if extraTLV.T != TypeNDEFFileControlTLV &&
			extraTLV.T != TypePropietaryFileControlTLV {
			rLen += parsed
			continue
		}

		// Then let's parse it as ControlTLV
		extraControlTLV := new(ControlTLV)
		parsed, err = extraControlTLV.Unmarshal(buf[rLen:])
		rLen += parsed
		if err != nil {
			return rLen, err
		}
		cc.TLVBlocks = append(cc.TLVBlocks, extraControlTLV)
	}
	if rLen != int(cc.CCLEN) { // They'd better be equal
		return rLen, fmt.Errorf("CapabilityContainer.Unmarshal: "+
			"expected %d bytes but parsed %d bytes",
			cc.CCLEN, i)
	}

	if err = cc.check(); err != nil {
		return rLen, err
	}
	return rLen, nil
}

// Marshal returns the byte slice representation of a CapabilityContainer.
// It returns an error if the fields in the struct are breaking the
// specification in some way, or if there is some other problem.
func (cc *CapabilityContainer) Marshal() ([]byte, error) {
	if err := cc.check(); err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	cclen := helpers.Uint16ToBytes(cc.CCLEN)
	buffer.Write(cclen[:])
	buffer.WriteByte(cc.MappingVersion)
	mle := helpers.Uint16ToBytes(cc.MLe)
	buffer.Write(mle[:])
	mlc := helpers.Uint16ToBytes(cc.MLc)
	buffer.Write(mlc[:])
	fcTLVBytes, err := cc.NDEFFileControlTLV.Marshal()
	if err != nil {
		return nil, err
	}
	buffer.Write(fcTLVBytes)
	for _, tlv := range cc.TLVBlocks {
		// Do not write TLV which are to be ignored
		// by the NFC Forum devices according to the
		// specs
		if !tlv.IsNDEFFileControlTLV() &&
			!tlv.IsPropietaryFileControlTLV() {
			continue
		}

		tlvBytes, err := tlv.Marshal()
		if err != nil {
			return nil, err
		}
		buffer.Write(tlvBytes)
	}
	return buffer.Bytes(), nil
}

// BUG(hector): Currently we don't check that the CapabilityContainer
// mapping version matches the specification version implemented by this
// library.

// Check tests that a CapabilityContainer follows the specification and
// returns an error if a problem is found.
func (cc *CapabilityContainer) check() error {
	if (0x0000 <= cc.CCLEN && cc.CCLEN <= 0x000e) || cc.CCLEN == 0xffff {
		return errors.New("CapabilityContainer.check: CCLEN is RFU")
	}

	if 0x0000 <= cc.MLe && cc.MLe <= 0x000e {
		return errors.New("CapabilityContainer.check: MLe is RFU")
	}

	if 0x0000 == cc.MLc {
		return errors.New("CapabilityContainer.check: MLc is RFU")
	}

	// Test that TLVs look ok
	if err := (*ControlTLV)(cc.NDEFFileControlTLV).check(); err != nil {
		return err
	}

	for _, tlv := range cc.TLVBlocks {
		if err := tlv.check(); err != nil {
			return err
		}
	}
	return nil
}
