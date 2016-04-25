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

package static

import (
	//	"testing"
	"fmt"

	"github.com/hsanjuan/go-ndef"
	"github.com/hsanjuan/go-ndef/types/wkt/text"
	"github.com/hsanjuan/go-nfctype4"
	"github.com/hsanjuan/go-nfctype4/drivers/swtag"
)

func ExampleTag_read() {
	// Let's create a NDEF Message first
	ndefMessage := ndef.NewMessage(
		ndef.NFCForumWellKnownType, "U", "",
		[]byte("\x01test.payload"))

	// Store this message in a static tag
	tag := New()
	err := tag.SetMessage(ndefMessage)
	if err != nil {
		fmt.Println(err)
	}

	// To read our tag we need a nfctype4.Device configured
	// with the swtag driver. The driver is connected to
	// our Tag.
	driver := &swtag.Driver{
		Tag: tag,
	}

	device := &nfctype4.Device{}
	device.Setup(driver)

	// Now we can read the message using the NFC Type 4 Tag
	// operation specification
	receivedMessage, err := device.Read()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(receivedMessage)
	}
	// Output:
	// urn:nfc:wkt:U:http://www.test.payload
}

func ExampleTag_write() {
	// Store this message in a static tag
	tag := New()

	// To read/write our tag we need a nfctype4.Device configured
	// with the swtag driver. The driver is connected to
	// our Tag.
	driver := &swtag.Driver{
		Tag: tag,
	}

	device := &nfctype4.Device{}
	device.Setup(driver)

	// Now we can update the message using the NFC Type 4 Tag
	// operation specification with a new message
	ndefMessage := new(ndef.Message)
	ndefMessage.Records = []*ndef.Record{
		&ndef.Record{
			TNF:     ndef.NFCForumWellKnownType,
			Type:    "T",
			Payload: text.New("This is a new message", "en"),
		},
	}
	err := device.Update(ndefMessage)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Finally let's peek at the message stored in the tag
	tagMessage := tag.GetMessage()
	fmt.Println(tagMessage)

	// Output:
	// urn:nfc:wkt:T:This is a new message
}
