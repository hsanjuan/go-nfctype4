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

package nfctype4

// CommandDriver interface is the minimal set of methods the drivers
// need to satisfy to allow communication between the NFC Device
// (provided by nfctype4) and the NFC Tag.
//
// Command drivers are in charge of implementing the link that
// allows this communication between Device and Tag.
//
// nfctype4 provides a LibnfcCommandDriver driver, which uses libnfc
// to use a connected NFC reader and read from a real NFC Type 4 Tag device.
//
// Other drivers are possible. See the DummyCommandDriver for an example,
// or the swtag which allows communication with software tags as defined
// in the Tag interface from the tags module.
type CommandDriver interface {
	Initialize() error // Makes the driver ready for TransceiveBytes
	Close()            // Tells the driver it won't be used anymore
	String() string    // Provides information about the driver's state
	// Sends and receive bytes to the NFC device
	TransceiveBytes(tx []byte, rxLen int) ([]byte, error)
}
