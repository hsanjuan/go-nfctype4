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
	"testing"
)

func TestGetLc(t *testing.T) {
	testcases := []struct {
		Lc       []byte
		Expected uint16
	}{
		{[]byte{}, 0},
		{[]byte{253}, 253},
		{[]byte{0xFF, 0x00, 54}, 54},
	}

	for _, c := range testcases {
		apdu := &CAPDU{
			Lc: c.Lc,
		}
		if r := apdu.GetLc(); r != c.Expected {
			t.Errorf("GetLc: expected %d. Got %d.",
				c.Expected, r)
		}
	}
}

func TestSetLc(t *testing.T) {
	testcases := []uint16{0, 1, 255, 256, 0xFFFF}
	for _, c := range testcases {
		apdu := &CAPDU{}
		apdu.SetLc(c)
		if r := apdu.GetLc(); r != c {
			t.Errorf("SetLc: expected %d. Got %d.",
				c, r)
		}
	}
}

func TestGetLe(t *testing.T) {
	testcases := []struct {
		Le       []byte
		Expected uint16
	}{
		{[]byte{}, 0},
		{[]byte{0}, 256},
		{[]byte{1}, 1},
		{[]byte{0xFF, 0xFE}, 65534},
		{[]byte{0x00, 0xFF, 0xFE}, 65534},
	}

	for _, c := range testcases {
		apdu := &CAPDU{
			Le: c.Le,
		}
		if r := apdu.GetLe(); r != c.Expected {
			t.Errorf("GetLe: expected %d. Got %d.",
				c.Expected, r)
		}
	}
}

func TestSetLe(t *testing.T) {
	testcases := []uint16{0, 1, 256, 65535}
	for _, c := range testcases {
		apdu := &CAPDU{}
		apdu.SetLe(c)
		if r := apdu.GetLe(); r != c {
			t.Errorf("SetLe: expected %d. Got %d.",
				c, r)
		}
	}

	// Corner case for Lc exiting
	apdu := &CAPDU{}
	apdu.SetLc(54)
	apdu.SetLe(2222)
	if len(apdu.Le) != 2 {
		t.Error("apdu.Le should use 2 bytes in APDUs with Lc")
	}

	apdu.SetLc(0)
	apdu.SetLe(2222)
	if len(apdu.Le) != 3 {
		t.Error("apdu.Le should use 3 bytes in APDUs without Lc")
	}
}

func TestCAPDUMarshalUnmarshal(t *testing.T) {
	testcases := []struct {
		Input    []byte
		Expected *CAPDU
	}{
		{
			[]byte{
				0x00,
				0x01,
				0x02,
				0x00,
			},
			&CAPDU{
				CLA: 0x00,
				INS: 0x01,
				P1:  0x02,
				P2:  0x00,
			},
		},
		{
			[]byte{
				0x00,
				0x01,
				0x02,
				0x00,
				0x23,
			},
			&CAPDU{
				CLA: 0x00,
				INS: 0x01,
				P1:  0x02,
				P2:  0x00,
				Le:  []byte{0x23},
			},
		},
		{
			[]byte{
				0x00,
				0x01,
				0x02,
				0x00,
				0x01,
				0xbb,
			},
			&CAPDU{
				CLA:  0x00,
				INS:  0x01,
				P1:   0x02,
				P2:   0x00,
				Lc:   []byte{0x01},
				Data: []byte{0xbb},
			},
		},
		{
			[]byte{
				0x00,
				0x01,
				0x02,
				0x00,
				0x01,
				0xbb,
				0xcc,
			},
			&CAPDU{
				CLA:  0x00,
				INS:  0x01,
				P1:   0x02,
				P2:   0x00,
				Lc:   []byte{0x01},
				Data: []byte{0xbb},
				Le:   []byte{0xcc},
			},
		},
		{
			[]byte{
				0x00,
				0x01,
				0x02,
				0x00,
				0x00,
				0xbb,
				0xcc,
			},
			&CAPDU{
				CLA: 0x00,
				INS: 0x01,
				P1:  0x02,
				P2:  0x00,
				Le:  []byte{0x00, 0xbb, 0xcc},
			},
		},
	}
	for i, c := range testcases {
		capdu := &CAPDU{}
		capdu.Unmarshal(c.Input)
		capduBytes, _ := capdu.Marshal()
		expectedBytes, _ := c.Expected.Marshal()
		t.Logf("Expected: % 02X", expectedBytes)
		t.Logf("Got     : % 02X ", capduBytes)

		if !bytes.Equal(capduBytes, expectedBytes) {
			t.Error("Unmarshal. Case", i, "failed.")
		}
		if capdu.GetLe() != c.Expected.GetLe() ||
			capdu.GetLc() != c.Expected.GetLc() {
			t.Error("Unmarshal. Case", i, "failed. Lc/Le mismatch")
		}
	}
}

func TestCAPDUNew(t *testing.T) {
	var capdu *CAPDU
	capdu = NewNDEFTagApplicationSelectAPDU()
	if capdu.GetLc() != 7 {
		t.Error("Error making NDEFTagApplicationSelectAPDU")
	}

	capdu = NewReadBinaryAPDU(5, 12)
	if capdu.P1 != 0 ||
		capdu.P2 != 5 ||
		capdu.GetLe() != 12 {
		t.Error("Error making NewReadBinaryAPDU")
	}

	capdu = NewSelectAPDU(256)
	if len(capdu.Data) != 2 ||
		capdu.Data[0] != 1 ||
		capdu.Data[1] != 0 {
		t.Error("Error making NewSelectAPDU")
	}
	capdu = NewCapabilityContainerReadAPDU()
}

func TestCAPDUMarshalBad(t *testing.T) {
	testcases := map[string]*CAPDU{
		"lc_1byte_cannot_be_0": &CAPDU{
			CLA:  0x04,
			INS:  0x02,
			P1:   0x00,
			P2:   0x00,
			Lc:   []byte{0x00},
			Data: []byte{},
			Le:   []byte{0x00},
		},
		"lc_cannot_have_2_bytes": &CAPDU{
			CLA:  0x04,
			INS:  0x02,
			P1:   0x00,
			P2:   0x00,
			Lc:   []byte{0x00, 0x01},
			Data: []byte{},
			Le:   []byte{0x00},
		},
		"lc_3_bytes_first_byte_must_be_0": &CAPDU{
			CLA:  0x04,
			INS:  0x02,
			P1:   0x00,
			P2:   0x00,
			Lc:   []byte{0x03, 0x00, 0x01},
			Data: []byte{},
			Le:   []byte{0x00},
		},
		"lc_3_bytes_all_bytes_cannot_be_0": &CAPDU{
			CLA:  0x04,
			INS:  0x02,
			P1:   0x00,
			P2:   0x00,
			Lc:   []byte{0x00, 0x00, 0x00},
			Data: []byte{},
			Le:   []byte{0x00},
		},
		"lc_cannot_have_more_than_3_bytes": &CAPDU{
			CLA:  0x04,
			INS:  0x02,
			P1:   0x00,
			P2:   0x00,
			Lc:   []byte{0x00, 0x01, 0x00, 0x00},
			Data: []byte{},
			Le:   []byte{0x00},
		},
		"lc_different_than_datalen": &CAPDU{
			CLA:  0x04,
			INS:  0x02,
			P1:   0x00,
			P2:   0x00,
			Lc:   []byte{0x02},
			Data: []byte{0xdd},
			Le:   []byte{},
		},
		"2_byte_le_needs_lc": &CAPDU{
			CLA:  0x04,
			INS:  0x02,
			P1:   0x00,
			P2:   0x00,
			Lc:   []byte{},
			Data: []byte{},
			Le:   []byte{0x00, 0x01},
		},
		"3_byte_le_cannot_have_lc": &CAPDU{
			CLA:  0x04,
			INS:  0x02,
			P1:   0x00,
			P2:   0x00,
			Lc:   []byte{0x01},
			Data: []byte{0xdd},
			Le:   []byte{0x00, 0x01, 0x00},
		},
		"3_byte_le_first_byte_must_be_0": &CAPDU{
			CLA:  0x04,
			INS:  0x02,
			P1:   0x00,
			P2:   0x00,
			Lc:   []byte{},
			Data: []byte{},
			Le:   []byte{0x01, 0x01, 0x00},
		},
		"le_cannot_have_more_than_3_bytes": &CAPDU{
			CLA:  0x04,
			INS:  0x02,
			P1:   0x00,
			P2:   0x00,
			Lc:   []byte{},
			Data: []byte{},
			Le:   []byte{0x01, 0x01, 0x00, 0x00},
		},
	}
	for k, capdu := range testcases {
		_, err := capdu.Marshal()
		if err == nil {
			t.Error("Testcase", k, "should have failed")
		} else {
			// FIXME: are we getting the error we expect?
			t.Logf("%s: %s", k, err.Error())
		}
	}
}
