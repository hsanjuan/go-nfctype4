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
	"fmt"
)

type CapabilityContainer struct {
	CCLEN          [2]byte // Size of this capability container - 000Fh to FFFEh
	MappingVersion byte    // Major-Minor version (4 bits each)
	MLe            [2]byte // Maximum data size that can be read using
	// ReadBinary. 000Fh-FFFFh
	MLc [2]byte // Maximum data size that can be sent using UpdateBinary.
	// 0001h-FFFFh
	NDEFFileControlTLV *NDEFFileControlTLV // Information to control and manage the NDEF file
	TLVBlocks          []*TLV              // Optional TLVs
}

func (cc *CapabilityContainer) ParseBytes(bytes []byte) (int, error) {
	if len(bytes) < 15 {
		return 0, errors.New("Not enough bytes to parse a Capability Container")
	}
	i := 0
	cc.CCLEN[0] = bytes[0]
	cc.CCLEN[1] = bytes[1]
	cc.MappingVersion = bytes[2]
	cc.MLe[0] = bytes[3]
	cc.MLe[1] = bytes[4]
	cc.MLc[0] = bytes[5]
	cc.MLc[1] = bytes[6]
	i += 7

	fc_tlv := new(NDEFFileControlTLV)
	parsed, err := fc_tlv.ParseBytes(bytes[i : i+8])
	if err != nil {
		return 0, err
	}
	cc.NDEFFileControlTLV = fc_tlv
	i += parsed

	cclen := BytesToUint16(cc.CCLEN)
	for i < int(cclen) {
		extra_tlv := new(TLV)
		parsed, err = extra_tlv.ParseBytes(bytes[i:len(bytes)])
		if err != nil {
			return 0, err
		}
		cc.TLVBlocks = append(cc.TLVBlocks, extra_tlv)
		i += parsed
	}
	if i != int(cclen) { // They'd better be equal
		return 0, fmt.Errorf("Capability Container expected %dBytes but parsed %dB",
			cclen, i)
	}

	if err = cc.Test(); err != nil {
		return 0, err
	}
	return i, nil
}

// Convert the CapabilityContainer to its byte representation
func (cc *CapabilityContainer) Bytes() ([]byte, error) {
	if err := cc.Test(); err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	buffer.Write(cc.CCLEN[:])
	buffer.WriteByte(cc.MappingVersion)
	buffer.Write(cc.MLe[:])
	buffer.Write(cc.MLc[:])
	ndef_tlv_bytes, err := cc.NDEFFileControlTLV.Bytes()
	if err != nil {
		return nil, err
	}
	buffer.Write(ndef_tlv_bytes)
	for _, tlv := range cc.TLVBlocks {
		tlv_bytes, err := tlv.Bytes()
		if err != nil {
			return nil, err
		}
		buffer.Write(tlv_bytes)
	}
	return buffer.Bytes(), nil
}

// Tests that a CC follows the standard
func (cc *CapabilityContainer) Test() error {
	cclen := BytesToUint16(cc.CCLEN)
	if (0x0000 <= cclen && cclen <= 0x000e) || cclen == 0xffff {
		return errors.New("Capability Container CCLEN is RFU")
	}

	mle := BytesToUint16(cc.MLe)
	if 0x0000 <= mle && mle <= 0x000e {
		return errors.New("Capability Container MLe is RFU")
	}

	mlc := BytesToUint16(cc.MLc)
	if 0x0000 == mlc {
		return errors.New("Capability Container MLc is RFU")
	}

	// Test that TLVs look ok
	if err := (*ControlTLV)(cc.NDEFFileControlTLV).Test(); err != nil {
		return err
	}

	for _, tlv := range cc.TLVBlocks {
		if err := tlv.Test(); err != nil {
			return err
		}
	}
	return nil
}
