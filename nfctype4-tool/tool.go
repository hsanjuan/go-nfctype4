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

// Package main provides a simple tool to read and write nfctype4 tags.
// See the Description variable for more information
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/hsanjuan/go-ndef"
	"github.com/hsanjuan/go-ndef/types"
	"github.com/hsanjuan/go-ndef/types/wkt/text"
	"github.com/hsanjuan/go-ndef/types/wkt/uri"
	"github.com/hsanjuan/go-nfctype4"
	"github.com/hsanjuan/go-nfctype4/drivers/libnfc"
)

// Description provides a description of the functionality of the tool
// for both Godoc and the --help output.
const Description = `
go-nfctype4-tool allows to easily read and write NFC Forum Type 4 Tags.

The tool attempts to provide a readable message on stdout when the contents
of the NDEF message can be interpreted. Otherwise, an explanatory message is
shown. The raw contents of the NDEF file can be printed with the -raw
flag in all cases.

The payload for writing can be provided directly as an argument or via stdin.
Note that the type is set by default to Text, but the specific format can be
edited with the -tnf and -type flags.

`

// Command line flags
var (
	driverFlag string
	rawFlag    bool
	tnfFlag    string
	typeFlag   string
	writeFlag  string
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage: go-nfctype4-tool "+
				"[options] <inspect|read|write|format> [payload]\n")
		fmt.Fprintf(os.Stderr, Description)

		fmt.Fprintf(os.Stderr, "Operations:\n")
		fmt.Fprintf(os.Stderr, " - inspect: print information about the NDEF Message.\n")
		fmt.Fprintf(os.Stderr, " - read: read the contents from a tag.\n")
		fmt.Fprintf(os.Stderr, " - write: update a tag with the given payload.\n")
		fmt.Fprintf(os.Stderr, " - format: erase the contents of a tag.\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
	}
	flag.StringVar(&driverFlag, "driver", "libnfc",
		"available drivers: libnfc")
	flag.BoolVar(&rawFlag, "raw", false, "Output raw NDEF File contents")
	flag.StringVar(&tnfFlag, "tnf", "wkt",
		"Type Name Format: "+
			"wkt, "+
			"ext, "+
			"media")
	flag.StringVar(&typeFlag, "type", "T",
		"The type of the message. Defaults to T[text]")
	flag.StringVar(&writeFlag, "write", "",
		"Write output to path")
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
	}

	if rawFlag {
		var buf bytes.Buffer
		for _, r := range ndefMessage.Records {
			buf.Write(r.Payload.Marshal())
		}
		output(buf.Bytes())
	} else {
		output([]byte(ndefMessage.String()))
	}
}

func doWrite() {
	payload := flag.Arg(1)
	if payload == "" {
		fmt.Fprintf(os.Stderr, "Write operation needs a payload.\n\n")
	}
	device := makeDevice()

	msg := new(ndef.Message)
	msg.Records = make([]*ndef.Record, 1)
	record := &ndef.Record{
		TNF:  tnfToCode(tnfFlag),
		Type: typeFlag,
	}

	switch tnfToCode(tnfFlag) {
	case ndef.NFCForumWellKnownType:
		switch typeFlag {
		case "U":
			record.Payload = uri.New(payload)
		case "T":
			record.Payload = text.New(payload, "en")
		default:
			record.Payload = &types.Generic{
				Payload: []byte(payload),
			}
		}
	case ndef.MediaType:
		record.Payload = &types.Generic{
			Payload: []byte(payload),
		}
	case ndef.NFCForumExternalType:
		record.Payload = &types.Generic{
			Payload: []byte(payload),
		}
	}
	msg.Records[0] = record

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

func output(t []byte) {
	if writeFlag != "" {
		file, err := os.Create(writeFlag)
		defer file.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		file.Write(t)
	} else {
		fmt.Println(string(t))
	}
}

func tnfToCode(tnf string) byte {
	switch tnf {
	case "wkt":
		return ndef.NFCForumWellKnownType
	case "ext":
		return ndef.NFCForumExternalType
	case "media":
		return ndef.MediaType
	default:
		fmt.Fprintln(os.Stderr, "Error: non-supported TNF provided")
		os.Exit(2)
	}
	return 0
}
