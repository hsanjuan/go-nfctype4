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
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/hsanjuan/go-ndef"
	"github.com/hsanjuan/go-ndef/types/absoluteuri"
	"github.com/hsanjuan/go-ndef/types/ext"
	"github.com/hsanjuan/go-ndef/types/generic"
	"github.com/hsanjuan/go-ndef/types/media"
	"github.com/hsanjuan/go-ndef/types/wkt/text"
	"github.com/hsanjuan/go-ndef/types/wkt/uri"
	"github.com/hsanjuan/go-nfctype4"
	"github.com/hsanjuan/go-nfctype4/drivers/libnfc"
)

// Description provides a description of the functionality of the tool
// for --help output.
const Description = `
nfctype4-tool allows to easily read and write NFC Forum Type 4 Tags.

Read operations will return the value read from the tag. If the -raw
flag is not specified, the program tries to produce a printable output
for the NDEF Message.

Write operations can take the payload as a command line argument or read
it from a file. The TNF and Type fields can be controlled with their respective
flags.

`

// Command line flags
var (
	driverFlag string
	fileFlag   string
	rawFlag    bool
	tnfFlag    string
	typeFlag   string
	writeFlag  string
	wait       bool
)

var waitDelay = 200 * time.Millisecond

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"Usage: nfctype4-tool "+
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
	flag.StringVar(&fileFlag, "file", "",
		"Read the payload from file (takes precedence over the payload argument)")
	flag.StringVar(&driverFlag, "driver", "libnfc",
		"available drivers: libnfc")
	flag.BoolVar(&wait, "wait", false, "Wait for the reader to detect the tag when not present")
	flag.StringVar(&writeFlag, "output", "",
		"Write output to path")
	flag.BoolVar(&rawFlag, "raw", false, "Output raw NDEF File contents")
	flag.StringVar(&tnfFlag, "tnf", "wkt",
		"Type Name Format: "+
			"wkt (Well-Known), "+
			"ext (External), "+
			"media (MIME)")
	flag.StringVar(&typeFlag, "type", "T",
		"The type of the message. Defaults to T[text]")
	flag.Parse()
}

func argError(msg string) {
	fmt.Fprint(os.Stderr, msg+"\n\n")
	flag.Usage()
	os.Exit(2)
}

func check(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

func main() {
	cmd := flag.Arg(0)
	var err error
	for {
		switch cmd {
		case "read":
			err = doRead()
		case "write":
			err = doWrite()
		case "format":
			err = doFormat()
		case "inspect":
			err = doInspect()
		case "":
			argError("Command argument is missing.")
		default:
			argError("Unrecognized command " + cmd)
		}

		if err == libnfc.ErrNoTargetsDetected {
			time.Sleep(waitDelay)
			continue
		}
		check(err)
		return
	}
}

func selectDriver() nfctype4.CommandDriver {
	switch driverFlag {
	case "libnfc":
		return new(libnfc.Driver)
	default:
		argError("Error: invalid driver selected.")
	}
	return nil
}

func makeDevice() *nfctype4.Device {
	driver := selectDriver()
	device := nfctype4.New(driver)
	return device
}

func doRead() error {
	device := makeDevice()
	ndefMessage, err := device.Read()
	if err != nil {
		return err
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
	return nil
}

func doWrite() error {
	var payload []byte

	if fileFlag == "" {
		payload = []byte(flag.Arg(1))
		if len(payload) == 0 {
			argError("Write operation needs a payload or --file.")
		}
	} else {
		var err error
		payload, err = ioutil.ReadFile(fileFlag)
		if err != nil {
			return err
		}
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
			record.Payload = uri.New(string(payload))
		case "T":
			record.Payload = text.New(string(payload), "en")
		default:
			record.Payload = &generic.Payload{
				Payload: []byte(payload),
			}
		}
	case ndef.AbsoluteURI:
		record.Payload = absoluteuri.New(typeFlag, payload)
	case ndef.MediaType:
		record.Payload = media.New(typeFlag, payload)
	case ndef.NFCForumExternalType:
		record.Payload = ext.New(typeFlag, payload)
	}
	msg.Records[0] = record

	err := device.Update(msg)
	if err != nil {
		return err
	}
	fmt.Println("Updated successful.")
	return nil
}

func doFormat() error {
	device := makeDevice()
	err := device.Format()
	if err != nil {
		return err
	}
	fmt.Println("Format operation successful.")
	return nil
}

func doInspect() error {
	device := makeDevice()
	ndefMessage, err := device.Read()
	if err != nil {
		return err
	}
	output([]byte(ndefMessage.Inspect()))
	return nil
}

func output(t []byte) {
	if writeFlag != "" {
		err := ioutil.WriteFile(writeFlag, t, 0644)
		check(err)
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
	case "uri":
		return ndef.AbsoluteURI
	default:
		argError("Error: non-supported TNF provided")
	}
	return 0
}
