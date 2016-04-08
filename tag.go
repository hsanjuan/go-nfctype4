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
	"github.com/hsanjuan/go-nfctype4/apdu"
)

// Tag represents a software implementation of a NFC Type 4 Tag.
// The communication between the Devices and the Tags is performed
// via APDUs, with Command APDUs being received and answered with
// Response APDUs.
//
// Tags must implement a Command function which should take care of
// processing CAPDUs and providing responses.
// It falls withing the Tag implementation to be consistent with the
// specification. The modules under nfctype4/tags offer examples of Tags.
//
// The `nfctype4/drivers/swtag` driver provides the binary interface
// for software tags. Check the `swtag` documentation to get an overview
// of its different applications.
type Tag interface {
	Command(capdu *apdu.CAPDU) *apdu.RAPDU
}
