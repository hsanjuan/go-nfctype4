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
	"bytes"
	"encoding/binary"
	"github.com/hsanjuan/ndef"
	"github.com/hsanjuan/nfctype4"
	"github.com/hsanjuan/nfctype4/apdu"
	"github.com/hsanjuan/nfctype4/capabilitycontainer"
	"github.com/hsanjuan/nfctype4/helpers"
)

// BUG(hector): Tag is not super-strict with the error responses
// in case of unexpected Commands.

// BUG(hector): Update operations are not implemented.

// The Default ID of the NDEF File for the tags.
const defaultNDEFFileID = 0xE104

// Version of the specification implemented by this tag
const (
	NFCForumMajorVersion = 2
	NFCForumMinorVersion = 0
)

// Tag implements a static NFC Type 4 Tags which holds an NDEFMessage.
//
// It is static because the message that is returned is always the same
// regardless of how many times it is Read.
//
// Since the static Tag implements the `nfctype4.Tag` interface, this
// tag can be used with the `nfctype4/drivers/swtag`. Please check the `swtag`
// module documentation for more information on the different uses.
type Tag struct {
	// The NDEF Message held in this Tag, which can be Read by an
	// NFC Type 4 compliant device
	Message *ndef.Message
	// The fileID for it (optional)
	FileID uint16
	// what has been selected
	selectedFileID uint16
}

// Command lets the Software tag receive Commands (CAPDUs) and
// provide respones (RAPDUs) according to each command.
// It is the heart of the behaviour of a NFC Type 4 Tag.
func (tag *Tag) Command(capdu *apdu.CAPDU) *apdu.RAPDU {
	if tag.Message == nil {
		return apdu.NewRAPDU(apdu.RAPDUInactiveState)
	}
	// Test message can be serialized
	_, err := tag.Message.Marshal()
	if err != nil {
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
		capdu.GetLc() == 0x07 &&
		binary.BigEndian.Uint64(capdu.Data) == nfctype4.NDEFAPPLICATION:
		// Selecting NDEF Application. Yes OK!
		return apdu.NewRAPDU(apdu.RAPDUCommandCompleted)
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
	switch tag.selectedFileID {
	case 0xE103: //CC
		if capdu.GetLe() < 15 {
			// Bad Le. 0x0F is the correct exact Le
			return &apdu.RAPDU{
				SW1: 0x6C,
				SW2: 0x0F,
			}
		}
		fileID := tag.FileID
		if fileID == 0 {
			fileID = defaultNDEFFileID
		}
		mBytes, _ := tag.Message.Marshal()
		// FIXME: Overflows?
		fileLen := helpers.Uint16ToBytes(uint16(len(mBytes)) + 2)
		tlv := &capabilitycontainer.NDEFFileControlTLV{
			T:      0x04,
			L:      0x06,
			FileID: fileID,
			// 2 NDEF Len bytes
			MaximumFileSize:         fileLen,
			FileReadAccessCondition: 0x00,
			// FIXME: Make this configurable
			FileWriteAccessCondition: 0x00,
		}
		cc := &capabilitycontainer.CapabilityContainer{
			CCLEN: [2]byte{0x00, 0x0F},
			MappingVersion: byte(NFCForumMajorVersion)<<4 |
				byte(NFCForumMinorVersion),
			MLe:                [2]byte{0xFF, 0xFF},
			MLc:                [2]byte{0xFF, 0xFF},
			NDEFFileControlTLV: tlv,
		}
		rapdu := apdu.NewRAPDU(apdu.RAPDUCommandCompleted)
		rbytes, _ := cc.Marshal()
		rapdu.ResponseBody = rbytes
		return rapdu
	case defaultNDEFFileID, tag.FileID: //NDEF File
		rapdu := apdu.NewRAPDU(apdu.RAPDUCommandCompleted)
		ndefBytes, _ := tag.Message.Marshal()
		// FIXME: what about very long messages
		ndefLen := helpers.Uint16ToBytes(uint16(len(ndefBytes)))
		var buffer bytes.Buffer
		buffer.Write(ndefLen[:])
		buffer.Write(ndefBytes)
		rapdu.ResponseBody = buffer.Bytes()
		return rapdu

	default:
		return apdu.NewRAPDU(apdu.RAPDUFileNotFound)
	}
}

// Unimplemented
func (tag *Tag) doUpdate(capdu *apdu.CAPDU) *apdu.RAPDU {
	return apdu.NewRAPDU(apdu.RAPDUInactiveState)
}
