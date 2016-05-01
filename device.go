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
	"fmt"

	"github.com/hsanjuan/go-ndef"
	"github.com/hsanjuan/go-nfctype4/capabilitycontainer"
	"github.com/hsanjuan/go-nfctype4/helpers"
)

// Device represents an NFC Forum device, that is, an application
// which allows to perform Read and Update operations on a NFC Type 4 Tag,
// by following the operation instructions stated in the specification.
//
// Interaction with physical and software Tags is done via CommandDrivers.
// A device is configured with a CommandDriver which is
// in charge of sending and receiving bytes from the Tags.
// The `nfctype4/drivers/libnfc` driver, for example, supports using a
// libnfc-supported reader to talk to a real NFC Type 4 Tag.
type Device struct {
	MajorVersion byte // 2
	MinorVersion byte // 0
	commander    *Commander
}

// tagState is used to store the relevant information obtained from a
// NDEF Detection Procedure
type tagState struct {
	NLEN               uint16
	MaxReadBinaryLen   uint16
	MaxUpdateBinaryLen uint16
	MaxNDEFLen         uint16
	ReadOnly           bool
}

// New returns a pointer to a new Device configured
// with the provided CommandDriver to perform
// operations on the Tags.
func New(cmdDriver CommandDriver) *Device {
	return &Device{
		MajorVersion: NFCForumMajorVersion,
		MinorVersion: NFCForumMinorVersion,
		commander: &Commander{
			Driver: cmdDriver,
		},
	}
}

// Setup [re]configures this device to use the provided
// command driver to perform operations on the tags.
func (dev *Device) Setup(cmdDriver CommandDriver) {
	dev.commander = &Commander{
		Driver: cmdDriver,
	}
}

// Read performs a full read operation on a NFC Type 4 tag.
//
// The CommandDriver provided with Setup is initialized and
// closed at the end of the operation.
//
// Read performs the NDEF Detect Procedure and, if successful,
// performs a read operation on the NDEF File.
//
// It returns the NDEFMessage stored in the tag, or an error
// if something went wrong.
func (dev *Device) Read() (*ndef.Message, error) {
	if err := dev.checkReady(); err != nil {
		return nil, err
	}

	// Initialize driver and make sure we close it at the end
	err := dev.commander.Driver.Initialize()
	defer dev.commander.Driver.Close()
	if err != nil {
		return nil, err
	}

	detectState, err := dev.ndefDetectProcedure()
	if err != nil {
		return nil, err
	}

	if detectState.NLEN == 0 {
		return nil, errors.New(
			"Device.Read: no NDEF Message detected.")
	}

	// Message detected
	// readLen represents what is the maximum amount of data we are going
	// to read from the Tag in one go.
	// It needs to be the minimum between maxReadBinaryLen and nlen
	readLen := detectState.MaxReadBinaryLen
	nlen := detectState.NLEN
	if nlen < readLen {
		readLen = nlen
	}
	// Read messages doing as many ReadBinary calls as necessary
	totalRead := uint16(0)
	var buffer bytes.Buffer // to hold what we are reading
	for totalRead < nlen {
		if nlen-totalRead < readLen { //last round
			readLen = nlen - totalRead
		}
		// Always offset the nlen bytes (2)
		chunk, err := dev.commander.ReadBinary(2+totalRead, readLen)
		if err != nil {
			return nil, err
		}
		buffer.Write(chunk)
		totalRead += readLen
	}

	ndefBytes := buffer.Bytes()

	// We finally have the NDEF Message. Parse it.
	ndefMessage := new(ndef.Message)
	if _, err := ndefMessage.Unmarshal(ndefBytes); err != nil {
		return nil, err
	}

	// Finally, return the parsed NDEF Message
	return ndefMessage, nil
}

// Update performs an update operation on a NFC Type 4 tag.
//
// The CommandDriver provided with Setup is initialized and
// closed at the end of the operation.
//
// The update operation starts by performing the NDEF
// Detection Procedure and the writing the provided NDEF Message
// to the NDEF File in the Tag, when the tag is not read-only.
//
// Note that update cannot be used to format a tag (clear the
// NDEF Message). For that, use Format().
//
// Update returns an error when there is a problem at some point
// in the process.
func (dev *Device) Update(m *ndef.Message) error {
	if err := dev.checkReady(); err != nil {
		return err
	}

	// Initialize driver and make sure we close it at the end
	err := dev.commander.Driver.Initialize()
	defer dev.commander.Driver.Close()
	if err != nil {
		return err
	}

	detectState, err := dev.ndefDetectProcedure()
	if err != nil {
		return err
	}

	if detectState.ReadOnly {
		return errors.New("Device.Update: the tag is read-only")
	}

	messageBytes, err := m.Marshal()
	if err != nil {
		return err
	}

	if len(messageBytes) > int(detectState.MaxNDEFLen-2) {
		return fmt.Errorf("Message is too large. Max size is %d",
			detectState.MaxNDEFLen-2)
	}

	// Per above, this can be done without risking overflows
	msgLen := uint16(len(messageBytes))

	// The number of bytes to write will be the maximum or,
	// if that's more than the message, just the message size
	writeLen := detectState.MaxUpdateBinaryLen
	if msgLen < writeLen {
		writeLen = msgLen
	}

	// If the msgLen + 2 fits inside the MaxUpdateBinaryLen
	// then we could do this in a single UpdateBinary command.
	// For the moment we do the slow way which works always.
	// Write 0000h in the NLEN field first
	err = dev.commander.UpdateBinary([]byte{0x00, 0x00}, 0)
	if err != nil {
		return err
	}

	// Write the message doing as many UpdateBinary calls as necessary
	totalWrite := uint16(0)
	for totalWrite < msgLen {
		if msgLen-totalWrite < writeLen { //last round
			writeLen = msgLen - totalWrite
		}
		err = dev.commander.UpdateBinary(
			messageBytes[totalWrite:totalWrite+writeLen],
			totalWrite+2) // Always offset the 2 NLEN bytes
		if err != nil {
			return err
		}
		totalWrite += writeLen
	}
	// Finally write NLEN
	msgLenBytes := helpers.Uint16ToBytes(msgLen)
	err = dev.commander.UpdateBinary(msgLenBytes[:], 0)
	if err != nil {
		return err
	}

	return nil
}

// Format performs an update operation which erases a tag.
// It does this by writing to the first two bytes of the NDEF File
// and setting their value to 0 (zero-length for the file).
//
// Be aware that the memory is not wiped or overwritten. An attacker
// may likely recover the values stored in the tag by resetting
// the length of the NDEF File to the maximum.
//
// To wipe the memory, issue an Update() with a Message of the maximum
// length supported by the tag and a randomized/meaningless payload.
//
// Format returns an error when a problem happens.
func (dev *Device) Format() error {
	if err := dev.checkReady(); err != nil {
		return err
	}

	// Initialize driver and make sure we close it at the end
	err := dev.commander.Driver.Initialize()
	defer dev.commander.Driver.Close()
	if err != nil {
		return err
	}

	detectState, err := dev.ndefDetectProcedure()
	if err != nil {
		return err
	}

	if detectState.ReadOnly {
		return errors.New("Device.Update: the tag is read-only")
	}

	err = dev.commander.UpdateBinary([]byte{0, 0}, 0)
	if err != nil {
		return err
	}

	return nil
}

func (dev *Device) ndefDetectProcedure() (*tagState, error) {
	state := new(tagState)
	// Select NDEF Application
	if err := dev.commander.NDEFApplicationSelect(); err != nil {
		return nil, err
	}

	// Select Capability Container
	if err := dev.commander.Select(capabilitycontainer.CCID); err != nil {
		return nil, err
	}

	// Read Capability Container and parse it. It should have 15 bytes.
	ccBytes, err := dev.commander.ReadBinary(0, 15)
	if err != nil {
		return nil, err
	}
	cc := new(capabilitycontainer.CapabilityContainer)
	if _, err := cc.Unmarshal(ccBytes); err != nil {
		return nil, err
	}

	// Check that we can read the tag
	fcTlv := cc.NDEFFileControlTLV
	if !(*capabilitycontainer.ControlTLV)(fcTlv).IsFileReadable() {
		return nil, errors.New(
			"Device.Read: NDEF File is marked as not readable.")
	}

	state.MaxReadBinaryLen = cc.MLe
	state.MaxUpdateBinaryLen = cc.MLc
	state.MaxNDEFLen = fcTlv.MaximumFileSize
	state.ReadOnly = (*capabilitycontainer.ControlTLV)(fcTlv).IsFileReadOnly()

	// Select the NDEF File
	if err := dev.commander.Select(fcTlv.FileID); err != nil {
		return nil, err
	}

	// Detect NDEF Message procedure 5.4.1
	nlenBytes, err := dev.commander.ReadBinary(0, 2)
	if err != nil {
		return nil, err
	}
	nlen := helpers.BytesToUint16([2]byte{nlenBytes[0], nlenBytes[1]})
	if nlen > state.MaxNDEFLen-2 {
		return nil, errors.New(
			"Device.Read: Device is not in a valid state")
	}
	state.NLEN = nlen
	return state, nil
}

func (dev *Device) checkReady() error {
	if dev.commander == nil {
		return errors.New("The Device has not been setup. " +
			"Please run Device.Setup(CommandDriver) first")
	}
	return nil
}
