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
	"fmt"
	"testing"

	"github.com/hsanjuan/go-ndef"
	"github.com/hsanjuan/go-ndef/types"
	"github.com/hsanjuan/go-ndef/types/wkt/text"
	"github.com/hsanjuan/go-ndef/types/wkt/uri"
	"github.com/hsanjuan/go-nfctype4/drivers/dummy"
	"github.com/hsanjuan/go-nfctype4/drivers/swtag"
	"github.com/hsanjuan/go-nfctype4/tags/static"
)

var dummyTestSets = map[string][][]byte{
	"yubikey_ok": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x00, 0x90, 0x00}, // CC binary read
		{0x90, 0x00},             // NDEF File Select
		{0x00, 0x43, 0x90, 0x00}, // NDEF File detect
		{0xd1, 0x01, 0x3f, 0x55, 0x04, 0x6d, 0x79, 0x2e, 0x79, 0x75, 0x62, 0x69, 0x63, 0x6f, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x65, 0x6f, 0x2f, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x90, 0x00}, // NDEF File Read
	},
}

var dummyTestSetsBad = map[string][][]byte{
	"bad_ndef_select": {
		{0x00, 0x00}, // NDEF app select
	},
	"cc_file_not_found": {
		{0x90, 0x00}, // NDEF app select
		{0x6A, 0x82}, // CC select (bad result)
	},
	"bad_cc_cclen": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0e, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x00, 0x90, 0x00}, // CC binary read. Set CCLEN to 0x000e
	},
	"bad_cc_read": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x00, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x90, 0x00}, // CC binary read. removed 1 byte from response
	},
	"bad_cc_mle": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x01, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x00, 0x90, 0x00}, // CC binary read. Mle to 0x00,0x01 (RFU)
	},
	"bad_cc_mlc": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x7f, 0x00, 0x00, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x00, 0x90, 0x00}, // CC binary read. Mlc to 0x00,0x00 (RFU)
	},
	"bad_cc_control_tlv_type": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x05, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x00, 0x90, 0x00}, // CC binary read. TLV type is 0x05 instead of 0x04
	},
	"bad_cc_control_tlv_access_conditions": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x01, 0x01, 0x90, 0x00}, // CC binary read. Access condition bytes set to 0x01 (RFU)
	},
	"ndef_file_read_protected": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x80, 0x00, 0x90, 0x00}, // CC binary read. Read access flag set to 0x80 (propietary)
	},
	"ndef_file_not_found": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x00, 0x90, 0x00}, // CC binary read
		{0x6A, 0x82}, // NDEF File Select. Not found
	},
	"ndef_file_select_error": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x00, 0x90, 0x00}, // CC binary read
		{0x00, 0x00}, // NDEF File Select
	},
	"ndef_file_zero_length": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x00, 0x90, 0x00}, // CC binary read
		{0x90, 0x00},             // NDEF File Select
		{0x00, 0x00, 0x90, 0x00}, // NDEF File detect. Size to 0
	},
	"device_invalid_state": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x00, 0x90, 0x00}, // CC binary read
		{0x90, 0x00},             // NDEF File Select
		{0xFF, 0xFF, 0x90, 0x00}, // NDEF File detect. Set size to 0xFFFF
	},
	"ndef_file_read_error": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x00, 0x90, 0x00}, // CC binary read
		{0x90, 0x00},             // NDEF File Select
		{0x00, 0x43, 0x90, 0x00}, // NDEF File detect
		{0xd1, 0x01, 0x3f, 0x55, 0x04, 0x6d, 0x79, 0x2e, 0x79, 0x75, 0x62, 0x69, 0x63, 0x6f, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x65, 0x6f, 0x2f, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x00, 0x00}, // NDEF File Read. Changed SW1 to 0x00
	},
	"ndef_file_bad_record": {
		{0x90, 0x00}, // NDEF app select
		{0x90, 0x00}, // CC select
		{0x00, 0x0f, 0x20, 0x00, 0x7f, 0x00, 0x7f, 0x04, 0x06, 0xe1, 0x04, 0x00, 0x7f, 0x00, 0x00, 0x90, 0x00}, // CC binary read
		{0x90, 0x00},             // NDEF File Select
		{0x00, 0x43, 0x90, 0x00}, // NDEF File detect
		{0xf1, 0x01, 0x3f, 0x55, 0x04, 0x6d, 0x79, 0x2e, 0x79, 0x75, 0x62, 0x69, 0x63, 0x6f, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6e, 0x65, 0x6f, 0x2f, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x90, 0x00}, // NDEF File Read. Changed first byte to enable CF
	},
}

func ExampleDevice_Read_dummy() {
	// This example uses the dummy.Driver, but
	// is exactly the same with the
	// libnfc.Driver, and would allow you to read
	// your Yubikey directly with your libnfc-device
	dummyDriver := &dummy.Driver{
		// ReceiveBytes should be set in the dummy so there is
		// something to answer. In this case, we simulate
		// a Yubikey.
		ReceiveBytes: dummyTestSets["yubikey_ok"],
	}
	device := new(Device)
	device.Setup(dummyDriver) // This device will use the dummyDriver
	message, err := device.Read()
	if err != nil {
		fmt.Println(err)
	} else {
		// Since Yubikeys provide a type 'U' NDEF
		// message, we can print the url like this
		fmt.Println(message)
	}
	// Output:
	// urn:nfc:wkt:U:https://my.yubico.com/neo/cccccccccccccccccccccccccccccccccccccccccccc
}

func TestRead_goodExamples(t *testing.T) {
	dummyDriver := new(dummy.Driver)
	device := new(Device)
	device.Setup(dummyDriver)
	for name, byteSet := range dummyTestSets {
		t.Log("Testing:", name)
		dummyDriver.ReceiveBytes = byteSet
		_, err := device.Read()
		if err != nil {
			t.Fail()
		}
	}
}

func TestRead_badExamples(t *testing.T) {
	expectedMessages := map[string]string{
		"bad_ndef_select":                      "Commander.NDEFApplicationSelect: unknown error. SW1: 00h. SW2: 00h",
		"cc_file_not_found":                    "Commander.Select: File e103h not found",
		"bad_cc_read":                          "CapabilityContainer.Unmarshal: not enough bytes to parse",
		"bad_cc_size":                          "CapabilityContainer.ParseBytes: not enough bytes to parse",
		"bad_cc_cclen":                         "CapabilityContainer.Unmarshal: expected 14 bytes but parsed 15 bytes",
		"bad_cc_mlc":                           "CapabilityContainer.check: MLc is RFU",
		"bad_cc_mle":                           "CapabilityContainer.check: MLe is RFU",
		"bad_cc_control_tlv_type":              "NDEFFileControlTLV.Unmarshal: TLV is not a NDEF File Control TLV",
		"bad_cc_control_tlv_access_conditions": "ControlTLV.check: Read Access Condition has RFU value",
		"ndef_file_read_protected":             "Device.Read: NDEF File is marked as not readable.",
		"ndef_file_not_found":                  "Commander.Select: File e104h not found",
		"ndef_file_select_error":               "Select: Unknown error. SW1: 00h. SW2: 00h",
		"ndef_file_zero_length":                "Device.Read: no NDEF Message detected.",
		"device_invalid_state":                 "Device.Read: Device is not in a valid state",
		"ndef_file_read_error":                 "Commander.ReadBinary: Error. SW1: 00h. SW2: 00h",
		"ndef_file_bad_record":                 "checkChunks: A single record cannot have the Chunk flag set",
	}
	device := new(Device)
	for name, byteSet := range dummyTestSetsBad {
		dummyDriver := &dummy.Driver{
			ReceiveBytes: byteSet,
		}
		device.Setup(dummyDriver)
		t.Log("Testing:", name)
		_, err := device.Read()
		if err != nil {
			if err.Error() != expectedMessages[name] {
				t.Error("Failed with unexpected message:", err)
			} else {
				t.Log("OK err: ", err)
			}
		} else {
			t.Error("Device.Read should have errored")
		}
	}
}

func TestUpdate(t *testing.T) {
	// We will use the software tags

	tag := static.New()

	driver := &swtag.Driver{
		Tag: tag,
	}
	device := new(Device)
	device.Setup(driver)

	// First test with a very simple message
	simpleMsg := &ndef.Message{
		Records: []*ndef.Record{
			&ndef.Record{
				TNF:  ndef.NFCForumWellKnownType,
				Type: "U",
				Payload: &uri.URI{
					IdentCode: 4,
					URIField:  "url.com",
				},
			},
		},
	}

	err := device.Update(simpleMsg)
	if err != nil {
		t.Error(err.Error())
	}

	readMsg, err := device.Read()
	if err != nil {
		t.Error(err.Error())
	}
	if !bytes.Equal(simpleMsg.Records[0].Payload.Marshal(),
		readMsg.Records[0].Payload.Marshal()) {
		t.Error("Payloads don't match for simpleMsg")
	}

	// Now test with a very long size
	longMsg := &ndef.Message{
		Records: []*ndef.Record{
			&ndef.Record{
				TNF:  ndef.NFCForumWellKnownType,
				Type: "U",
			},
		},
	}
	longMsg.Records[0].Payload = &types.Generic{
		Payload: make([]byte, 0xFFE0),
	}
	err = device.Update(longMsg)
	if err != nil {
		t.Error(err.Error())
	}

	readMsg, err = device.Read()
	if err != nil {
		t.Error(err.Error())
	}
	if !bytes.Equal(
		longMsg.Records[0].Payload.Marshal(),
		readMsg.Records[0].Payload.Marshal()) {
		t.Error("Payloads don't match for longMsg")
	}

	// Now test with a message over the maximum size
	badMsg := &ndef.Message{
		Records: []*ndef.Record{
			&ndef.Record{
				TNF:     ndef.NFCForumWellKnownType,
				Type:    "T",
				Payload: &types.Generic{},
			},
		},
	}
	badMsg.Records[0].Payload.Unmarshal(make([]byte, 0xFFFE))
	err = device.Update(badMsg)
	if err == nil {
		t.Error("Update with badMsg should have failed")
	} else {
		t.Log("The expected error was:", err)
	}
}

func TestFormat(t *testing.T) {
	// We will use the software tags

	tag := static.New()

	mRecordPl := text.New("This is a text message", "en")

	// First assume our tag has a simple message
	simpleMsg := &ndef.Message{
		Records: []*ndef.Record{
			&ndef.Record{
				TNF:     ndef.NFCForumWellKnownType,
				Type:    "T",
				Payload: mRecordPl,
			},
		},
	}
	tag.SetMessage(simpleMsg)

	driver := &swtag.Driver{
		Tag: tag,
	}
	device := new(Device)
	device.Setup(driver)

	// Format the tag
	err := device.Format()
	if err != nil {
		t.Error(err.Error())
	}

	// Try to read
	_, err = device.Read()
	if err == nil {
		t.Error("Reading from an empty tag should have failed")
	} else {
		if err.Error() != "Device.Read: no NDEF Message detected." {
			t.Error("Unexpected error happened: ", err.Error())
		}
	}
}
