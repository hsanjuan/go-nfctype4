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

package capabilitycontainer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hsanjuan/nfctype4/helpers"
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
	CCLEN              [2]byte             // Size of this capability container - 000Fh to FFFEh
	MappingVersion     byte                // Major-Minor version (4 bits each)
	MLe                [2]byte             // Maximum data read with ReadBinary. 000Fh-FFFFh
	MLc                [2]byte             // Maximum data to write with UpdateBinary. 0001h-FFFFh
	NDEFFileControlTLV *NDEFFileControlTLV // NDEF file information
	TLVBlocks          []*TLV              // Optional TLVs
}

// Reset clears all the fields of the CapabilityContainer to their
// default values.
func (cc *CapabilityContainer) Reset() {
	cc.CCLEN = [2]byte{0, 0}
	cc.MappingVersion = 0
	cc.MLe = [2]byte{0, 0}
	cc.MLc = [2]byte{0, 0}
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
func (cc *CapabilityContainer) Unmarshal(buf []byte) (int, error) {
	cc.Reset()
	if len(buf) < 15 {
		return 0, errors.New(
			"CapabilityContainer.Unmarshal: " +
				"not enough bytes to parse")
	}
	i := 0
	cc.CCLEN[0] = buf[0]
	cc.CCLEN[1] = buf[1]
	cc.MappingVersion = buf[2]
	cc.MLe[0] = buf[3]
	cc.MLe[1] = buf[4]
	cc.MLc[0] = buf[5]
	cc.MLc[1] = buf[6]
	i += 7

	fcTLV := new(NDEFFileControlTLV)
	parsed, err := fcTLV.Unmarshal(buf[i : i+8])
	if err != nil {
		return 0, err
	}
	cc.NDEFFileControlTLV = fcTLV
	i += parsed

	cclen := helpers.BytesToUint16(cc.CCLEN)
	for i < int(cclen) {
		extraTLV := new(TLV)
		parsed, err = extraTLV.Unmarshal(buf[i:len(buf)])
		if err != nil {
			return 0, err
		}
		cc.TLVBlocks = append(cc.TLVBlocks, extraTLV)
		i += parsed
	}
	if i != int(cclen) { // They'd better be equal
		return 0, fmt.Errorf("CapabilityContainer.Unmarshal: "+
			"expected %d bytes but parsed %d bytes",
			cclen, i)
	}

	if err = cc.check(); err != nil {
		return 0, err
	}
	return i, nil
}

// Marshal returns the byte slice representation of a CapabilityContainer.
// It returns an error if the fields in the struct are breaking the
// specification in some way, or if there is some other problem.
func (cc *CapabilityContainer) Marshal() ([]byte, error) {
	if err := cc.check(); err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.Write(cc.CCLEN[:])
	buffer.WriteByte(cc.MappingVersion)
	buffer.Write(cc.MLe[:])
	buffer.Write(cc.MLc[:])
	fcTLVBytes, err := cc.NDEFFileControlTLV.Marshal()
	if err != nil {
		return nil, err
	}
	buffer.Write(fcTLVBytes)
	for _, tlv := range cc.TLVBlocks {
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
	cclen := helpers.BytesToUint16(cc.CCLEN)
	if (0x0000 <= cclen && cclen <= 0x000e) || cclen == 0xffff {
		return errors.New("CapabilityContainer.check: CCLEN is RFU")
	}

	mle := helpers.BytesToUint16(cc.MLe)
	if 0x0000 <= mle && mle <= 0x000e {
		return errors.New("CapabilityContainer.check: MLe is RFU")
	}

	mlc := helpers.BytesToUint16(cc.MLc)
	if 0x0000 == mlc {
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
