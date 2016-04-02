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

const (
	NFC_FORUM_MAJOR_VERSION = 2
	NFC_FORUM_MINOR_VERSION = 0
)

// Drivers
const (
	DUMMY = iota
	LIBNFC
)

type CommandDriver interface {
	Initialize() error
	Close()
	String() string
	TransceiveBytes(tx []byte, rx_len int) ([]byte, error)
}

// Driver can be set up manually or using Setup()
var Driver CommandDriver

func Setup(driver_i int) error {
	switch driver_i {
	case DUMMY:
		dr := new(DummyCommandDriver)
		Driver = dr
	case LIBNFC:
		dr := new(LibNFCCommandDriver)
		Driver = dr
	}
	return nil
}

// Reads a tag
// Returns the NDEFMessage stored in the tag, or an error if something went wrong
//
func Read() (*ndef.Message, error) {
	if Driver == nil {
		err_txt := "The Command Driver has not been configured. " +
			"Use nfctype4.Setup(nfctype4.LIBNFC|...|) or add your custom with " +
			"nfctype4.Driver = YourCommandDriver"
		return nil, errors.New(err_txt)
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
	cc_bytes, err := CapabilityContainerRead()
	if err != nil {
		return nil, err
	}
	cc := new(CapabilityContainer)
	if _, err := cc.ParseBytes(cc_bytes); err != nil {
		return nil, err
	}

	// Check that we can read the tag
	fc_tlv := cc.NDEFFileControlTLV
	if !(*ControlTLV)(fc_tlv).IsFileReadable() {
		return nil, errors.New("NDEF File is marked as not readable")
	}

	// Select the NDEF File
	if err := Select(fc_tlv.FileID[:]); err != nil {
		return nil, err
	}

	// Detect NDEF Message procedure 5.4.1
	maximum_readbinary_size := BytesToUint16(cc.MLe)
	maximum_ndef_file_size := BytesToUint16(fc_tlv.MaximumFileSize)
	nlen_bytes, err := ReadBinary(0, 2)
	if err != nil {
		return nil, err
	}
	nlen := BytesToUint16([2]byte{nlen_bytes[0], nlen_bytes[1]})
	if nlen == 0 {
		return nil, errors.New("No NDEF Message to read Detected")
	} else if nlen > maximum_ndef_file_size-2 {
		return nil, errors.New("Type 4 Tag platform is not in a valid state")
	}

	// Message detected
	// Read length needs to be the minimum between 255, maximum_readbinary_size and neln
	read_length := maximum_readbinary_size
	if nlen < read_length {
		read_length = nlen
	}
	// Read messages doing as many ReadBinary calls as necessary
	total_read := uint16(0)
	var buffer bytes.Buffer
	for total_read < nlen {
		if nlen-total_read < read_length { //last round
			read_length = nlen - total_read
		}
		// Always offset the nlen bytes (2)
		chunk, err := ReadBinary(2+total_read, read_length)
		if _, err = buffer.Write(chunk); err != nil {
			return nil, err
		}
		total_read += read_length
	}

	ndef_bytes := buffer.Bytes()
	ndef_message := new(ndef.Message)
	if _, err := ndef_message.ParseBytes(ndef_bytes); err != nil {
		return nil, err
	}

	// Finally, return the parsed NDEF Message
	return ndef_message, nil
}
