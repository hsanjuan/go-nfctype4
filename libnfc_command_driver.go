// +build !nolibnfc

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
	"errors"
	"fmt"
	"github.com/fuzxxl/nfc/2.0/nfc"
)

// LibnfcCommandDriver implements the CommandDriver interface by
// using libnfc bindings to send and receive data from the NFC device.
//
// For this driver to work, libnfc needs to be correctly installed and
// configured in the system.
type LibnfcCommandDriver struct {
	Modulation   nfc.Modulation // The modulation to use
	DeviceNumber int            // The libnfc devices number to choose
	device       *nfc.Device
	deviceList   []string
	target       *nfc.ISO14443aTarget
}

// Initialize performs the necessary operations to make sure that the
// driver is in conditions to TransceiveBytes.
//
// For the LibnfcCommandDriver this involves detecting available nfc devices,
// selecting one and setting it up as an Initiator, using it to scan for targets
// and selecting the first target available (or fail). This means that
// for initialization to work, the NFC device needs to be visible to the reader
// already, as otherwise there is no target to work with.
//
// It returns an error when some step fails.
func (driver *LibnfcCommandDriver) Initialize() error {
	driver.Modulation = nfc.Modulation{Type: nfc.ISO14443a, BaudRate: nfc.Nbr212}

	deviceList, err := nfc.ListDevices()
	if err != nil {
		return err
	}
	driver.deviceList = deviceList

	if len(deviceList) == 0 {
		return errors.New("no libnfc devices detected")
	}
	if len(deviceList) <= driver.DeviceNumber {
		return fmt.Errorf("libnfc does not provide device %d",
			driver.DeviceNumber)
	}
	device, err := nfc.Open(deviceList[driver.DeviceNumber])
	if err != nil {
		return err
	}

	driver.device = &device
	err = driver.device.InitiatorInit()
	if err != nil {
		return err
	}

	var targets []nfc.Target
	targets, err = driver.device.InitiatorListPassiveTargets(driver.Modulation)
	if len(targets) == 0 {
		return errors.New("no targets detected. Place tag on reader and retry")
	}
	driver.target = targets[0].(*nfc.ISO14443aTarget)

	_, err = driver.device.InitiatorSelectPassiveTarget(
		driver.Modulation,
		driver.target.UID[0:driver.target.UIDLen])
	if err != nil {
		return err
	}
	return nil
}

// String returns some information extracted from libnfc about the NFC device
// and the target that was selected. It should be used after calling
// Initialize().
func (driver *LibnfcCommandDriver) String() string {
	var str string
	str += fmt.Sprintf("NeoRead uses libnfc %s\n", nfc.Version())
	str += fmt.Sprintf("Modulation: Type: %d, BaudRate: %d\n",
		driver.Modulation.Type,
		driver.Modulation.BaudRate)

	str += fmt.Sprintln("Detected devices:")
	for i, d := range driver.deviceList {
		str += fmt.Sprintf("  * [%d] %s\n", i, d)
	}
	str += fmt.Sprintln()
	info, err := driver.device.Information()
	if err == nil {
		str += fmt.Sprintln("Device information: ")
		str += fmt.Sprintln(info)
	} else {
		str += fmt.Sprintln("No device information.")
	}
	if driver.target != nil {
		str += fmt.Sprintln("Target information: ")
		str += fmt.Sprintln(driver.target)
	} else {
		str += fmt.Sprintln("No target information.")
	}
	return str
}

// TransceiveBytes is used to send and receive bytes from the libnfc device.
// It receives a byte slice to send, and an expected maximum length to receive.
// It returns the received data or an error when something fails.
func (driver *LibnfcCommandDriver) TransceiveBytes(tx []byte, rxLen int) ([]byte, error) {
	rx := make([]byte, rxLen) //buffer to receive bytes
	n, err := driver.device.InitiatorTransceiveBytes(tx, rx, -1)
	if err != nil {
		if err.(nfc.Error) == nfc.EOVFLOW {
			return nil, fmt.Errorf("Libnfc: expected to "+
				"read %d but the buffer"+
				"was overflowed with %d bytes", rxLen, n)
		}
		return nil, err
	}
	fmt.Print("{ ")
	for _, b := range rx[0:n] {
		fmt.Printf("0x%02x, ", b)
	}
	fmt.Println(" },")
	return rx[0:n], nil
}

// Close shuts down the driver correctly by closing the device that was used.
func (driver *LibnfcCommandDriver) Close() {
	if driver.device != nil {
		driver.device.Close()
	}
}
