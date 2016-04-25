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

// Package helpers provides some useful functions common to nfctype4.
package helpers

import (
	"bytes"
	"errors"
	"runtime"
)

// BytesToUint16 takes a 2-byte array and returns the corresponding
// uint16 value (BigEndian).
func BytesToUint16(field [2]byte) uint16 {
	return uint16(field[0])<<8 | uint16(field[1])
}

// Uint16ToBytes takes an uint16 value and returns the corresponding
// 2-byte array (BigEndian).
func Uint16ToBytes(value uint16) [2]byte {
	byte0 := byte(value >> 8)
	byte1 := byte(0x00ff & value) //Probably the casting would suffice
	return [2]byte{byte0, byte1}
}

// GetBytes reads n bytes from a bytes.Buffer and panics with an error
// when the buffer cannot provide them because there is no more
// to read.
// It is like an out of bounds panic when reading from a slice, except
// it is not a runtime.Error panic. This is used when Unmarshaling
// data. The bytes.Buffer keeps track of the parsing progress,
// while this wrapper allows to catch out of bounds errors in
// a single place and save dozens of "if error!=nil" blocks.
func GetBytes(b *bytes.Buffer, n int) []byte {
	slice := make([]byte, n)
	nread, err := b.Read(slice)
	if err != nil || nread != n {
		panic(errors.New("Unexpected end of data."))
	}
	return slice
}

// GetByte reads a single byte from a bytes.Buffer and panics with an error
// when the buffer cannot provide it because there is no more
// to read. See GetBytes for justification and usage.
func GetByte(b *bytes.Buffer) byte {
	byte, err := b.ReadByte()
	if err != nil {
		panic(errors.New("Unexpected end of data"))
	}
	return byte
}

// HandleErrorPanic recovers from custom panics (those which are not
// runtime.Error) and sets the value of the error variable passed as
// reference.
func HandleErrorPanic(err *error, functionName string) {
	if r := recover(); r != nil {
		if _, ok := r.(runtime.Error); ok {
			panic(r)
		}
		perr := r.(error)
		*err = errors.New(functionName + ": " + perr.Error())
	}
}
