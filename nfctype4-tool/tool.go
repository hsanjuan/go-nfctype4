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

// Package main provides a simple tool to read and write nfctype4 tags
//
// The read output depends on the ndef.Message.String().
// The write input is plain text and is stored in a plain text NDEF Message
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hsanjuan/go-ndef"
	"github.com/hsanjuan/go-nfctype4"
	"github.com/hsanjuan/go-nfctype4/drivers/libnfc"
)

// Command line flags
var (
	driverFlag string
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage: go-nfctype4-tool "+
				"[options] <read|write|format> [payload]\n")
		fmt.Fprintf(os.Stderr, "Operations:\n")
		fmt.Fprintf(os.Stderr, " - read: read the contents from a tag.\n")
		fmt.Fprintf(os.Stderr, " - write: update a tag with the given payload.\n")
		fmt.Fprintf(os.Stderr, " - format: erase the contents of a tag.\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
	}
	flag.StringVar(&driverFlag, "driver", "libnfc",
		"available drivers: libnfc")
	flag.Parse()
}

func main() {
	cmd := flag.Arg(0)
	switch cmd {
	case "read":
		doRead()
	case "write":
		doWrite()
	case "format":
		doFormat()
	case "":
		fmt.Fprintf(os.Stderr, "Command argument is missing.\n\n")
		flag.Usage()
		os.Exit(2)
	default:
		fmt.Fprintf(os.Stderr, "Unrecognized command %s.\n\n", cmd)
		flag.Usage()
		os.Exit(2)
	}
}

func selectDriver() nfctype4.CommandDriver {
	switch driverFlag {
	case "libnfc":
		return new(libnfc.Driver)
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid driver selected.\n\n")
		os.Exit(2)
	}
	return nil
}

func makeDevice() *nfctype4.Device {
	driver := selectDriver()
	device := new(nfctype4.Device)
	device.Setup(driver)
	return device
}

func doRead() {
	device := makeDevice()
	ndefMessage, err := device.Read()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		fmt.Println(ndefMessage)
	}
}

func doWrite() {
	payload := flag.Arg(1)
	if payload == "" {
		fmt.Fprintf(os.Stderr, "Write operation needs a payload.\n\n")
	}
	device := makeDevice()
	msg := new(ndef.Message)
	msg.Payload = []byte(payload)
	msg.TNF = ndef.NFCForumWellKnownType
	msg.Type = []byte("T")
	err := device.Update(msg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		fmt.Println("Updated successful.")
	}
}

func doFormat() {
	device := makeDevice()
	err := device.Format()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		fmt.Println("Format operation successful.")
	}
}
