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

// NDEFFileAddress Address in which the NDEF File is stored.
// It is initialized to a default of 0x8888.
//
// The valid ranges are 0x0001-E101,0xE104-3EFF, 0x4000-FFFE.
// Values 0x0000, 0xE102, 0xE103, 0x3F00, 0x3FFF are reserved.
// 0xFFFF is RFU.
const NDEFFileAddress = uint16(0x8888)

// Version of the specification implemented by this tag
const (
	NFCForumMajorVersion = 2
	NFCForumMinorVersion = 0
)

// NDEFAPPLICATION is the name for the NDEF Application.
const NDEFAPPLICATION = uint64(0xD2760000850101)

// Tag implements a static NFC Type 4 Tags which holds a NDEFMessage.
//
// It called static because the message that is returned is always the same
// regardless of how many times it is Read.
//
// Since the static Tag implements the `tags.Tag` interface, this
// tag can be used with the `nfctype4/drivers/swtag`. Please check the `swtag`
// module documentation for more information on the different uses.
//
// Please use static.New() to create tags, or remember to do a Tag.Initialize()
// as otherwise tags will refuse to work.
type Tag struct {
	// what has been selected
	selectedFileID uint16
	// A shadow buffer for updates
	memory map[uint16][]byte
}

// New returns a new *Tag in Initialized state (empty)
func New() *Tag {
	t := new(Tag)
	t.Initialize()
	return t
}

// Initialize resets a Tag to an initialized state (empty)
// It will drop the memory contents if they previously existed
// and de-select any files.
func (tag *Tag) Initialize() {
	tag.selectedFileID = 0
	tag.memory = make(map[uint16][]byte)

	// Set the capability container
	cc := &capabilitycontainer.CapabilityContainer{
		CCLEN: 15,
		MappingVersion: byte(NFCForumMajorVersion)<<4 |
			byte(NFCForumMinorVersion),
		// FIXME: This is actually important and should
		// stay below the maximum frame values specified in
		// the RATs command
		MLe: 0x000F, // We could put more... or less
		MLc: 0x000F,
		NDEFFileControlTLV: &capabilitycontainer.NDEFFileControlTLV{
			T:                        0x04,
			L:                        0x06,
			FileID:                   NDEFFileAddress,
			MaximumFileSize:          0xFFFE,
			FileReadAccessCondition:  0x00,
			FileWriteAccessCondition: 0x00, // FIXME: Make configurable

		},
	}
	ccBytes, _ := cc.Marshal()
	tag.memory[capabilitycontainer.CCID] = ccBytes

	// Set an empty NDEF file
	tag.memory[NDEFFileAddress] = []byte{0, 0} // NLEN to 0
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
	tag.memory[NDEFFileAddress] = buf.Bytes()
	return nil
}

// GetMessage allows to retrieve the NDEF message stored
// in the tag.
// It returns nil when there is nothing stored.
func (tag *Tag) GetMessage() *ndef.Message {
	file := tag.memory[NDEFFileAddress]
	if len(file) < 2 {
		return nil
	}

	nlen := helpers.BytesToUint16([2]byte{file[0], file[1]})
	if nlen == 0 {
		return nil
	}

	mBytes := file[2:]
	msg := new(ndef.Message)
	// if this fails, we will return nil too
	msg.Unmarshal(mBytes)
	return msg
}

// Command lets the Software tag receive Commands (CAPDUs) and
// provide respones (RAPDUs) according to each command.
// It is the heart of the behaviour of a NFC Type 4 Tag.
func (tag *Tag) Command(capdu *apdu.CAPDU) *apdu.RAPDU {
	if tag.memory == nil {
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
		addr := helpers.BytesToUint16([2]byte{
			capdu.Data[0],
			capdu.Data[1]})
		_, ok := tag.memory[addr]
		if !ok {
			return apdu.NewRAPDU(apdu.RAPDUFileNotFound)
		}

		// We have something in that address
		tag.selectedFileID = addr
		return apdu.NewRAPDU(apdu.RAPDUCommandCompleted)
	default:
		return apdu.NewRAPDU(apdu.RAPDUFileNotFound)
	}
}

func (tag *Tag) doRead(capdu *apdu.CAPDU) *apdu.RAPDU {
	rBytes, ok := tag.memory[tag.selectedFileID]
	if !ok {
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
	if tag.selectedFileID == capabilitycontainer.CCID {
		// No, you cannot write the CC
		apdu.NewRAPDU(apdu.RAPDUCommandNotAllowed)
	}
	_, ok := tag.memory[tag.selectedFileID]
	if !ok {
		return apdu.NewRAPDU(apdu.RAPDUFileNotFound)
	}

	offset := int(helpers.BytesToUint16([2]byte{capdu.P1, capdu.P2}))
	data := capdu.Data

	file := tag.memory[tag.selectedFileID]
	newFileLen := offset + len(data)
	if newFileLen > len(file) {
		// increase the size of the file
		newFile := make([]byte, newFileLen)
		copy(newFile, file)
		tag.memory[tag.selectedFileID] = newFile
	}
	copy(tag.memory[tag.selectedFileID][offset:], data)
	return apdu.NewRAPDU(apdu.RAPDUCommandCompleted)
}
