/***
    Copyright (c) 2020, Hector Sanjuan

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
	"errors"
	"fmt"

	"github.com/hsanjuan/go-nfctype4/helpers"
)

// RAPDU types which come handy
const (
	RAPDUCommandCompleted = iota
	RAPDUCommandNotAllowed
	RAPDUFileNotFound
	RAPDUInactiveState
)

// RAPDU represents a Response APDU, which is received as an answer to
// Command APDUs. Response APDUs may contain data, along with two trailer
// bytes indicating the status.
type RAPDU struct {
	ResponseBody []byte //Data bytes
	SW1          byte   //Status Word 1
	SW2          byte   //Status Word 2
}

// Reset clears the fields of the RAPDU to their default values
func (apdu *RAPDU) Reset() {
	apdu.ResponseBody = []byte{}
	apdu.SW1 = 0
	apdu.SW2 = 0
}

// String provides a string representation of the RAPDU
func (apdu *RAPDU) String() string {
	str := ""
	str += fmt.Sprintf("SW1: %02x SW2: %02x | Data: ", apdu.SW1, apdu.SW2)
	for _, b := range apdu.ResponseBody {
		str += fmt.Sprintf("%02x ", b)
	}
	return str
}

// Unmarshal parses a byte slice and sets the RAPDU fields accordingly.
// It always resets the RAPDU before parsing.
// It returns the number of bytes parsed or an error if something goes wrong.
func (apdu *RAPDU) Unmarshal(buf []byte) (rLen int, err error) {
	defer helpers.HandleErrorPanic(&err, "RAPDU.Unmarshal")
	bytesBuf := bytes.NewBuffer(buf)
	apdu.Reset()

	// This could be done without the helpers.
	// But let's be consistent with the rest of Unmarshals...
	dataLen := bytesBuf.Len() - 2
	if dataLen < 0 {
		return 0, errors.New("RAPDU.Unmarshal: " +
			"Not enough data to parse response")
	}
	apdu.ResponseBody = helpers.GetBytes(bytesBuf, dataLen)
	apdu.SW1 = helpers.GetByte(bytesBuf)
	apdu.SW2 = helpers.GetByte(bytesBuf)
	return len(buf), nil
}

// Marshal returns the byte slice representation of the RAPDU
func (apdu *RAPDU) Marshal() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Write(apdu.ResponseBody)
	buffer.WriteByte(apdu.SW1)
	buffer.WriteByte(apdu.SW2)
	return buffer.Bytes(), nil
}

// CommandCompleted checks if the RAPDU indicates a successful
// completion of a command.
func (apdu *RAPDU) CommandCompleted() bool {
	return apdu.SW1 == 0x90 && apdu.SW2 == 0x00
}

// FileNotFound checks if the RAPDU indicates that a file
// was not found (usually in response to a Select operation).
func (apdu *RAPDU) FileNotFound() bool {
	return apdu.SW1 == 0x6A && apdu.SW2 == 0x82
}

// NewRAPDU provides a quick way to obtain some commonly
// used Response APDUs. See the RAPDU constants for
// the types which are supported
func NewRAPDU(which int) *RAPDU {
	switch which {
	case RAPDUCommandCompleted:
		return &RAPDU{
			SW1: 0x90,
			SW2: 0x00,
		}
	case RAPDUCommandNotAllowed:
		return &RAPDU{
			SW1: 0x69,
			SW2: 0x00,
		}
	case RAPDUFileNotFound:
		return &RAPDU{
			SW1: 0x6A,
			SW2: 0x82,
		}
	case RAPDUInactiveState:
		return &RAPDU{
			SW1: 0x69,
			SW2: 0x01,
		}
	}
	return nil
}
