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
	"github.com/hsanjuan/ndef"
)

// Tag is used to performs operations on a NFC Type 4 Tag, like reading.
type Tag struct {
	MajorVersion byte // unused
	MinorVersion byte // unused
}

// Read performs a full read operation on a NFC Type 4 tag using
// the configured Driver (CommandDriver) to
// communicate with the reader device.
//
// Read takes care of initializing and closing the driver.
//
// It returns the NDEFMessage stored in the tag, or an error
// if something went wrong.
func (t *Tag) Read() (*ndef.Message, error) {
	if Driver == nil {
		errTxt := "The Command Driver has not been configured. " +
			"You can set it set it with \"nfctype4.Driver = new(" +
			"YourCommandDriver|" +
			"DummyCommandDriver|" +
			"LibnfcCommandDriver" +
			")\""
		return nil, errors.New(errTxt)
	}

	// Initialize driver and make sure we close it at the end
	err := Driver.Initialize()
	defer Driver.Close()
	if err != nil {
		return nil, err
	}
	// Select NDEF Application
	if err := NDEFApplicationSelect(); err != nil {
		return nil, err
	}

	// Select Capability Container
	if err := CapabilityContainerSelect(); err != nil {
		return nil, err
	}

	// Read Capability Container and parse it
	ccBytes, err := CapabilityContainerRead()
	if err != nil {
		return nil, err
	}
	cc := new(CapabilityContainer)
	if _, err := cc.ParseBytes(ccBytes); err != nil {
		return nil, err
	}

	// Check that we can read the tag
	fcTlv := cc.NDEFFileControlTLV
	if !(*ControlTLV)(fcTlv).IsFileReadable() {
		return nil, errors.New(
			"Tag.Read: NDEF File is marked as not readable")
	}

	// Select the NDEF File
	if err := Select(fcTlv.FileID[:]); err != nil {
		return nil, err
	}

	// Detect NDEF Message procedure 5.4.1
	maxReadBinaryLen := BytesToUint16(cc.MLe)
	maxNdefLen := BytesToUint16(fcTlv.MaximumFileSize)
	nlenBytes, err := ReadBinary(0, 2)
	if err != nil {
		return nil, err
	}
	nlen := BytesToUint16([2]byte{nlenBytes[0], nlenBytes[1]})
	if nlen == 0 {
		return nil, errors.New(
			"Tag.Read: no NDEF Message to read Detected")
	} else if nlen > maxNdefLen-2 {
		return nil, errors.New(
			"Tag.Read: Tag is not in a valid state")
	}

	// Message detected
	// Read length needs to be the minimum between maxReadBinaryLen and nlen
	readLen := maxReadBinaryLen
	if nlen < readLen {
		readLen = nlen
	}
	// Read messages doing as many ReadBinary calls as necessary
	totalRead := uint16(0)
	var buffer bytes.Buffer
	for totalRead < nlen {
		if nlen-totalRead < readLen { //last round
			readLen = nlen - totalRead
		}
		// Always offset the nlen bytes (2)
		chunk, err := ReadBinary(2+totalRead, readLen)
		if err != nil {
			return nil, err
		}
		if _, err = buffer.Write(chunk); err != nil {
			return nil, err
		}
		totalRead += readLen
	}

	ndefBytes := buffer.Bytes()
	ndefMessage := new(ndef.Message)
	if _, err := ndefMessage.ParseBytes(ndefBytes); err != nil {
		return nil, err
	}

	// Finally, return the parsed NDEF Message
	return ndefMessage, nil
}
