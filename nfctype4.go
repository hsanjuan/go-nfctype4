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

// Package nfctype4 is an implementation of the NFC Forum Type 4 Tag
// Operation Specification (Version 2.0)
//
// nfctype4 can be used for both reading Tags, but also for implementing
// software-based Tags that adjust to the specification.
//
// The `Device` type offers functionality to perform `Read` and `Update`
// on NFC Type 4 Tags.
//
// The `Tag` type allows in turn to implement software-based tags. An
// static version of a software tag is provided with 'tags/static'.
//
// The bridge between the `Device` and the hardware is covered by the
// `libnfc4/drivers/*`, which implement the `CommandDriver` interface.
// A `libnfc` driver is provided, which allows working with any
// libnfc-supported hardware.
//
// nfctype4 tries to be very strict when parsing and serializing
// data, and the Marshal() and Unmarshal() operations for the
// different entities involved in the implementation will
// error whenever a trangression to the specification is detected
// (for example, when Reserved for Future Use values are used).
package nfctype4

// This is the NFC Type 4 Tag standard version that we are following.
const (
	NFCForumMajorVersion = 2
	NFCForumMinorVersion = 0
)

// NDEFAPPLICATION is the name for the NDEF Application.
const NDEFAPPLICATION = uint64(0xD2760000850101)
