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
			TLVBlocks: []*ControlTLV{
				&ControlTLV{
					T:                        0x05,
					L:                        0x06,
					FileID:                   0xE104,
					MaximumFileSize:          0x05,
					FileReadAccessCondition:  0x00,
					FileWriteAccessCondition: 0x00,
				},
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
			// This TLV Block should be ignored acc to the specs
			TLVBlocks: []*ControlTLV{
				&ControlTLV{
					T:                        0x00, //bad!
					L:                        0x06,
					FileID:                   0xE104,
					MaximumFileSize:          0x05,
					FileReadAccessCondition:  0x00,
					FileWriteAccessCondition: 0x00,
				},
			},
		},
	}

	testcasesbad := map[string]*CapabilityContainer{
		"bad_cclen": &CapabilityContainer{
			CCLEN:          3, // bad
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
		"bad_mle": &CapabilityContainer{
			CCLEN:          20,
			MappingVersion: 0x20,
			MLe:            0, //bad
			MLc:            255,
			NDEFFileControlTLV: &NDEFFileControlTLV{
				T:                        0x04,
				L:                        0x06,
				FileID:                   0xE104,
				MaximumFileSize:          90,
				FileReadAccessCondition:  0,
				FileWriteAccessCondition: 0,
			},
			TLVBlocks: []*ControlTLV{
				&ControlTLV{
					T:                        0x05,
					L:                        0x06,
					FileID:                   0xE104,
					MaximumFileSize:          0x05,
					FileReadAccessCondition:  0x00,
					FileWriteAccessCondition: 0x00,
				},
			},
		},
		"bad_mlc": &CapabilityContainer{
			CCLEN:          20,
			MappingVersion: 0x20,
			MLe:            255,
			MLc:            0, //bad
			NDEFFileControlTLV: &NDEFFileControlTLV{
				T:                        0x04,
				L:                        0x06,
				FileID:                   0xE104,
				MaximumFileSize:          90,
				FileReadAccessCondition:  0,
				FileWriteAccessCondition: 0,
			},
			TLVBlocks: []*ControlTLV{
				&ControlTLV{
					T:                        0x05,
					L:                        0x06,
					FileID:                   0xE104,
					MaximumFileSize:          0x05,
					FileReadAccessCondition:  0x00,
					FileWriteAccessCondition: 0x00,
				},
			},
		},
		"bad_ftlv": &CapabilityContainer{
			CCLEN:          20,
			MappingVersion: 0x20,
			MLe:            255,
			MLc:            255,
			NDEFFileControlTLV: &NDEFFileControlTLV{
				T:                        0x04,
				L:                        0x06,
				FileID:                   0xE102, //bad
				MaximumFileSize:          90,
				FileReadAccessCondition:  0,
				FileWriteAccessCondition: 0,
			},
			TLVBlocks: []*ControlTLV{
				&ControlTLV{
					T:                        0x05,
					L:                        0x06,
					FileID:                   0xE104,
					MaximumFileSize:          0x05,
					FileReadAccessCondition:  0x00,
					FileWriteAccessCondition: 0x00,
				},
			},
		},
	}
	t.Log("Testing with good CCs")
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
	t.Log("Testing with bad CCs")
	for k, c := range testcasesbad {
		_, err := c.Marshal()
		if err == nil {
			t.Error("Testcase", k, "should have failed")
		} else {
			// FIXME: are we getting the error we expect?
			t.Logf("%s: %s", k, err.Error())
		}
	}

}
