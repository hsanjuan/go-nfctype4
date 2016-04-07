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
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	testcases := []*CapabilityContainer{
		&CapabilityContainer{
			CCLEN:          15,
			MappingVersion: 0x20,
			MLe:            255,
			MLc:            255,
			NDEFFileControlTLV: &NDEFFileControlTLV{
				T:                        0x04,
				L:                        0x06,
				FileID:                   0xE104,
				MaximumFileSize:          90,
				FileReadAccessCondition:  0,
				FileWriteAccessCondition: 0,
			},
		},
		&CapabilityContainer{
			CCLEN:          20,
			MappingVersion: 0x20,
			MLe:            255,
			MLc:            255,
			NDEFFileControlTLV: &NDEFFileControlTLV{
				T:                        0x04,
				L:                        0x06,
				FileID:                   0xE104,
				MaximumFileSize:          90,
				FileReadAccessCondition:  0,
				FileWriteAccessCondition: 0,
			},
			TLVBlocks: []*TLV{
				&TLV{
					T: 0x05,
					L: [3]byte{0x01, 0x00, 0x00},
					V: []byte{0xFF},
				},
			},
		},
	}

	for _, c := range testcases {
		ccBytes, _ := c.Marshal()
		tempcc := &CapabilityContainer{}
		tempcc.Unmarshal(ccBytes)
		tempccBytes, _ := tempcc.Marshal()
		t.Logf("Expected: % 02X", ccBytes)
		t.Logf("Got     : % 02X ", tempccBytes)
		if !bytes.Equal(ccBytes, tempccBytes) {
			t.Fail()
		}
	}
}
