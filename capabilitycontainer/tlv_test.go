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

import "testing"

func TestControlTLVMarshalErrors(t *testing.T) {
	testcases := map[string]*ControlTLV{
		"file_id_reserved": &ControlTLV{
			T:      0x04,
			L:      0x06,
			FileID: 0xe102,
		},
		"file_id_rfu": &ControlTLV{
			T:      0x04,
			L:      0x06,
			FileID: 0xffff,
		},
		"maximum_size_rfu": &ControlTLV{
			T:               0x04,
			L:               0x06,
			FileID:          0xe104,
			MaximumFileSize: 0x03,
		},
		"readaccess_rfu": &ControlTLV{
			T:                       0x04,
			L:                       0x06,
			FileID:                  0xe104,
			MaximumFileSize:         0x05,
			FileReadAccessCondition: 0x02,
		},
		"writeaccess_rfu": &ControlTLV{
			T:                        0x04,
			L:                        0x06,
			FileID:                   0xe104,
			MaximumFileSize:          0x05,
			FileWriteAccessCondition: 0x02,
		},
	}

	for k, tlv := range testcases {
		_, err := tlv.Marshal()
		if err == nil {
			t.Error("Testcase", k, "should have failed")
		} else {
			// FIXME: are we getting the error we expect?
			t.Logf("%s: %s", k, err.Error())
		}
	}
}

func TestTLVUmarshal(t *testing.T) {
	testcasesBad := map[string][]byte{
		"bad_long_length":  []byte{0x04, 0xFF, 0x00, 0x01, 0xdd},
		"bad_long_length2": []byte{0x04, 0xFF, 0x01, 0x01, 0xdd},
		"bad_short_length": []byte{0x04, 0xFF, 0xdd},
		"length_mismatch":  []byte{0x04, 0x05, 0xdd, 0xdd},
	}

	for k, tlvB := range testcasesBad {
		tlv := &TLV{}
		_, err := tlv.Unmarshal(tlvB)
		if err == nil {
			t.Error("Testcase", k, "should have failed")
		} else {
			// FIXME: are we getting the error we expect?
			t.Logf("%s: %s", k, err.Error())
		}
	}
}

func TestTLVMarshal(t *testing.T) {
	testcasesBad := map[string]*TLV{
		"bad_t": &TLV{
			T: 0x07,
		},
		"bad_long_length": &TLV{
			T: 0x05,
			L: [3]byte{0xFF, 0x00, 0x01},
		},
		"bad_long_length2": &TLV{
			T: 0x05,
			L: [3]byte{0xFF, 0xFF, 0xFF},
		},
		"length_mismatch": &TLV{
			T: 0x05,
			L: [3]byte{0x03},
			V: []byte{0xdd},
		},
	}

	for k, tlv := range testcasesBad {
		_, err := tlv.Marshal()
		if err == nil {
			t.Error("Testcase", k, "should have failed")
		} else {
			// FIXME: are we getting the error we expect?
			t.Logf("%s: %s", k, err.Error())
		}
	}
}

func TestControlTLVIsFuncs(t *testing.T) {
	testcases := []struct {
		TLV                                       *ControlTLV
		Readable, Writeable, Readonly, Propietary bool
	}{
		{
			&ControlTLV{
				T:                        0x04,
				L:                        0x06,
				FileID:                   0xe104,
				MaximumFileSize:          0x00,
				FileReadAccessCondition:  0x00,
				FileWriteAccessCondition: 0x00,
			}, true, true, false, false,
		},
		{
			&ControlTLV{
				T:                        0x04,
				L:                        0x06,
				FileID:                   0xe104,
				MaximumFileSize:          0x00,
				FileReadAccessCondition:  0x00,
				FileWriteAccessCondition: 0xFF,
			}, true, false, true, false,
		},
		{
			&ControlTLV{
				T:                        0x05,
				L:                        0x06,
				FileID:                   0xe104,
				MaximumFileSize:          0x00,
				FileReadAccessCondition:  0x00,
				FileWriteAccessCondition: 0x00,
			}, true, true, false, true,
		},
	}

	for i, stru := range testcases {
		if stru.TLV.IsFileReadable() != stru.Readable {
			t.Error("TLV should be readable. Case", i)
		}
		if stru.TLV.IsFileWriteable() != stru.Writeable {
			t.Error("TLV should be writeable. Case", i)
		}
		if stru.TLV.IsFileReadOnly() != stru.Readonly {
			t.Error("TLV should be read only. Case", i)
		}
		if stru.TLV.IsPropietaryFileControlTLV() != stru.Propietary {
			t.Error("TLV should be a Propietary TLV. Case", i)
		}
	}
}
