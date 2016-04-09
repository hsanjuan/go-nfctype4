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

// Package static provides the implementation of a static software-based
// NFC Forum Type 4 Tag which holds a NDEF Message.
package static

import (
	//	"fmt"
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/hsanjuan/go-ndef"
	"github.com/hsanjuan/go-nfctype4/apdu"
	"github.com/hsanjuan/go-nfctype4/capabilitycontainer"
	"github.com/hsanjuan/go-nfctype4/helpers"
)

// BUG(hector): Tag is not super-strict with the error responses
// in case of unexpected Commands.

// The Default ID of the NDEF File for the tags.
const defaultNDEFFileID = 0xE104

// Version of the specification implemented by this tag
const (
	NFCForumMajorVersion = 2
	NFCForumMinorVersion = 0
)

// NDEFAPPLICATION is the name for the NDEF Application.
const NDEFAPPLICATION = uint64(0xD2760000850101)

// Tag implements a static NFC Type 4 Tags which holds a NDEFMessage.
//
// It is static because the message that is returned is always the same
// regardless of how many times it is Read.
//
// Since the static Tag implements the `tags.Tag` interface, this
// tag can be used with the `nfctype4/drivers/swtag`. Please check the `swtag`
// module documentation for more information on the different uses.
type Tag struct {
	// The fileID for it (optional)
	FileID uint16
	// what has been selected
	selectedFileID uint16
	// A shadow buffer for updates
	file []byte
}

// Initialize sets this Tag in Initialized state.
// This means that the tag is empty and the length of the NDEF File
// is set to 0.
func (tag *Tag) Initialize() {
	tag.file = []byte{0x00, 0x00}
}

// SetMessage programs the NDEF message for this tag.
// It returns an error if the m.Marshal() does (which
// would indicate and invalid message).
func (tag *Tag) SetMessage(m *ndef.Message) error {
	mBytes, err := m.Marshal()
	if err != nil {
		return err
	}
	nlen := len(mBytes)
	if nlen > 0xFFFE { // 0xFFFF is RFU accoring to specs
		return errors.New("Tag.SetMessage: message too long")
	}

	// Write the NDEF File
	var buf bytes.Buffer
	nlenBytes := helpers.Uint16ToBytes(uint16(nlen))
	buf.Write(nlenBytes[:])
	buf.Write(mBytes)
	tag.file = buf.Bytes()
	return nil
}

// GetMessage allows to retrieve the NDEF message stored
// in the tag.
// It returns nil when there is nothing stored.
func (tag *Tag) GetMessage() *ndef.Message {
	if len(tag.file) < 2 {
		return nil
	}
	nlen := helpers.BytesToUint16([2]byte{tag.file[0], tag.file[1]})
	if nlen == 0 {
		return nil
	}
	mBytes := tag.file[2:]
	msg := new(ndef.Message)
	// if this fails, we will return nil too
	msg.Unmarshal(mBytes)
	return msg
}

// Command lets the Software tag receive Commands (CAPDUs) and
// provide respones (RAPDUs) according to each command.
// It is the heart of the behaviour of a NFC Type 4 Tag.
func (tag *Tag) Command(capdu *apdu.CAPDU) *apdu.RAPDU {
	if len(tag.file) < 2 {
		return apdu.NewRAPDU(apdu.RAPDUInactiveState)
	}

	switch capdu.INS {
	case apdu.INSSelect:
		return tag.doSelect(capdu)
	case apdu.INSRead:
		return tag.doRead(capdu)
	case apdu.INSUpdate:
		return tag.doUpdate(capdu)
	default:
		return apdu.NewRAPDU(apdu.RAPDUCommandNotAllowed)
	}
}

func (tag *Tag) doSelect(capdu *apdu.CAPDU) *apdu.RAPDU {
	// We support 3 types of select: for the NDEFApp, for the CC and for
	// the NDEF File
	switch {
	case capdu.P1 == 0x04 &&
		capdu.P2 == 0x00 &&
		capdu.GetLc() == 0x07:
		// Convert data to Uint64
		data8 := make([]byte, 8)
		copy(data8[1:], capdu.Data)
		dataVal := binary.BigEndian.Uint64(data8)
		if dataVal == NDEFAPPLICATION {
			// Selecting NDEF Application. Yes OK!
			return apdu.NewRAPDU(apdu.RAPDUCommandCompleted)
		}
		return apdu.NewRAPDU(apdu.RAPDUFileNotFound)

	case capdu.P1 == 0x00 && capdu.P2 == 0x0C:
		if capdu.GetLc() != uint16(2) {
			// Lc inconsistent with P1-P2
			return &apdu.RAPDU{
				SW1: 0x6A,
				SW2: 0x87,
			}
		}
		// Selecting by id
		fID := helpers.BytesToUint16([2]byte{
			capdu.Data[0],
			capdu.Data[1]})
		// Cover the cases where we select a valid file
		if fID == 0xE103 || //CC
			(tag.FileID == 0x00 && fID == defaultNDEFFileID) ||
			(tag.FileID != 0x00 && tag.FileID == fID) {
			tag.selectedFileID = fID
			return apdu.NewRAPDU(apdu.RAPDUCommandCompleted)
		}
		fallthrough // File not found
	default:
		return apdu.NewRAPDU(apdu.RAPDUFileNotFound)
	}
}

func (tag *Tag) doRead(capdu *apdu.CAPDU) *apdu.RAPDU {
	// Read the selected file
	if tag.selectedFileID == 0x00 {
		return apdu.NewRAPDU(apdu.RAPDUFileNotFound)
	}

	var rBytes []byte

	switch tag.selectedFileID {
	case capabilitycontainer.CCID: // Capability Container
		// Figure out the File ID
		fileID := tag.FileID
		if fileID == 0 {
			fileID = defaultNDEFFileID
		}

		// Now we can create the ControlTLV
		tlv := &capabilitycontainer.NDEFFileControlTLV{
			T:      0x04,
			L:      0x06,
			FileID: fileID,
			// 2 NDEF Len bytes
			MaximumFileSize:         0xFFFE,
			FileReadAccessCondition: 0x00,
			// FIXME: Make this configurable
			FileWriteAccessCondition: 0x00,
		}

		// Attach it to a CC
		cc := &capabilitycontainer.CapabilityContainer{
			CCLEN: 15,
			MappingVersion: byte(NFCForumMajorVersion)<<4 |
				byte(NFCForumMinorVersion),
			MLe:                0x00FF, // Force chunks for large
			MLc:                0x00FF,
			NDEFFileControlTLV: tlv,
		}
		rBytes, _ = cc.Marshal()

	case defaultNDEFFileID:
		if tag.FileID != 0 && tag.FileID != defaultNDEFFileID {
			// Then there is no file here
			return apdu.NewRAPDU(apdu.RAPDUFileNotFound)
		}
		// Otherwise pretend we are reading the Default
		fallthrough
	case tag.FileID: // Read NDEF File
		rBytes = tag.file
	default:
		return apdu.NewRAPDU(apdu.RAPDUFileNotFound)
	}

	// We have rBytes ready. Let's make sure the response
	// adapts to the offset and Le provided in the CAPDU
	offset := int(helpers.BytesToUint16([2]byte{capdu.P1, capdu.P2}))
	rLen := int(capdu.GetLe())
	rBytesLen := len(rBytes)
	if rLen+offset > rBytesLen {
		rLen = rBytesLen - offset
	}
	rapdu := apdu.NewRAPDU(apdu.RAPDUCommandCompleted)
	rapdu.ResponseBody = rBytes[offset : offset+rLen]
	return rapdu
}

func (tag *Tag) doUpdate(capdu *apdu.CAPDU) *apdu.RAPDU {
	// Read the selected file
	if tag.selectedFileID == 0x00 {
		return apdu.NewRAPDU(apdu.RAPDUFileNotFound)
	}

	// Rule out when it's the default but the FileID is set
	// to something else
	if tag.selectedFileID == defaultNDEFFileID &&
		tag.FileID != 0 && tag.FileID != defaultNDEFFileID {
		return apdu.NewRAPDU(apdu.RAPDUFileNotFound)
	}

	// Rule out the cases when it's not the Default, but also
	// not the FileID
	if tag.selectedFileID != tag.FileID &&
		tag.selectedFileID != defaultNDEFFileID {
		return apdu.NewRAPDU(apdu.RAPDUFileNotFound)
	}

	// We are writing the NDEF File
	offset := int(helpers.BytesToUint16([2]byte{capdu.P1, capdu.P2}))
	data := capdu.Data

	newFileSize := offset + len(data)
	if newFileSize > len(tag.file) {
		// increase the size of the file
		newFile := make([]byte, newFileSize)
		copy(newFile, tag.file)
		tag.file = newFile
	}
	copy(tag.file[offset:], data)
	return apdu.NewRAPDU(apdu.RAPDUCommandCompleted)
}
