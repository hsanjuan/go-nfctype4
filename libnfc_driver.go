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

const RECV_BUFFER_SIZE = 1024

/*
 * Implements a CommandDriver using libnfc bindings
 *
 */

type LibNFCCommandDriver struct {
	device      *nfc.Device
	device_list []string
	Modulation  nfc.Modulation
	target      *nfc.ISO14443aTarget
}

func (driver *LibNFCCommandDriver) Initialize() error {
	driver.Modulation = nfc.Modulation{Type: nfc.ISO14443a, BaudRate: nfc.Nbr212}

	device_list, err := nfc.ListDevices()
	if err != nil {
		return err
	}
	driver.device_list = device_list

	if len(device_list) == 0 {
		return errors.New("No libnfc devices detected")
	}
	device, err := nfc.Open(device_list[0])
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
		return errors.New("No targets detected. Place tag on reader and retry")
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

func (driver *LibNFCCommandDriver) String() string {
	var str string
	str += fmt.Sprintf("NeoRead uses libnfc %s\n", nfc.Version())
	str += fmt.Sprintf("Modulation: Type: %d, BaudRate: %d\n",
		driver.Modulation.Type,
		driver.Modulation.BaudRate)

	str += fmt.Sprintln("Detected devices:")
	for i, d := range driver.device_list {
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

func (driver *LibNFCCommandDriver) TransceiveBytes(tx []byte, rx_len int) ([]byte, error) {
	rx := make([]byte, rx_len) //buffer to receive bytes
	n, err := driver.device.InitiatorTransceiveBytes(tx, rx, -1)
	if err != nil {
		if err.(nfc.Error) == nfc.EOVFLOW {
			return nil, fmt.Errorf("LibNFC: expected to read %d but the buffer"+
				"was overflowed with %d bytes", rx_len, n)
		} else {
			return nil, err
		}
	}
	return rx[0:n], nil
}

func (driver *LibNFCCommandDriver) Close() {
	if driver.device != nil {
		driver.device.Close()
	}
}
