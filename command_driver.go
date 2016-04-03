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

// CommandDriver is the minimal set of methods the drivers used to communicate
// with an NFC device need to satisfy. A command driver allows to use a typical
// NFC reader to send an receive data from an NFC device.
//
// nfctype4 provides a LibnfcCommandDriver driver, which uses libnfc
// to use a connected NFC reader and read from the NFC Type 4 Tag device.
// It also provides a DummyDriver for testing.
type CommandDriver interface {
	Initialize() error // Makes the driver ready for TransceiveBytes
	Close()            // Tells the driver it won't be used anymore
	String() string    // Provides information about the driver's state
	// Sends and receive bytes to the NFC device
	TransceiveBytes(tx []byte, rxLen int) ([]byte, error)
}
