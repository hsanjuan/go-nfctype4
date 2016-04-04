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

// BUG(hector): Update (write/erase) operations are not yet supported

// Package nfctype4 implements the NFC Forum Type 4 Tag Operation
// Specification, which allows to read the information contained
// in this popular type of NFC Tags.
//
// Use the Device type to perform Read() and Update() operations on the Tag.
//
// Devices must be Setup() first with a CommandDriver. nfctype4 offers
// a libnfc command driver, which allows to work with any libnfc-detected
// device, but custom command drivers are supported too.
package nfctype4

// This is the NFC Type 4 Tag standard version that we are following.
const (
	NFCForumMajorVersion = 2
	NFCForumMinorVersion = 0
)
