Go-nfctype4
===========

| Master/stable | Unstable | Reference |
|:-------------:|:--------:|:---------:|
| [![Build Status](https://travis-ci.org/hsanjuan/go-nfctype4.svg?branch=master)](https://travis-ci.org/hsanjuan/go-nfctype4) [![Coverage Status](https://coveralls.io/repos/github/hsanjuan/go-nfctype4/badge.svg?branch=master)](https://coveralls.io/github/hsanjuan/go-nfctype4?branch=master) | [![Build Status](https://travis-ci.org/hsanjuan/go-nfctype4.svg?branch=unstable)](https://travis-ci.org/hsanjuan/go-nfctype4) [![Coverage Status](https://coveralls.io/repos/github/hsanjuan/go-nfctype4/badge.svg?branch=unstable)](https://coveralls.io/github/hsanjuan/go-nfctype4?branch=unstable) | [![GoDoc](https://godoc.org/github.com/hsanjuan/go-nfctype4?status.svg)](http://godoc.org/github.com/hsanjuan/go-nfctype4) |

Package `go-nfctype4` implements the NFC Forum Type 4 Tag Operation Specification.

It provides a `Device` type which allows to interact with NFC Devices (like readers) and perform `Read` and `Update` operations on the NFC tags.

It also allows to easily implement software-based NFC Type 4 compliant tags, which can be easily used to provide hardware NFC Readers in target-mode with the necessary functionality to adjust to the specification and act like real Type 4 Tags.

go-nfctype4-tool
----------------

`go-nfctype4-tool` is a command-line tool to read and write NFC Type 4 tags. It can be installed with `go get github.com/hsanjuan/go-nfctype4/go-nfctype4-tool`.

You can then run `go-nfctype4-tool -h` for more information.

Packages
--------

`go-nfctype4` and its subpackages offer access to the implementation of all the entities described in the NFC Forum Type 4 Tag specification. These are the links to the reference documentation of the most relevant packages:

  * https://godoc.org/github.com/hsanjuan/go-nfctype4 : Provides the `Device`, `CommandDriver`, `Commander` and `Tag` types. They are the main entry point to interact with NFC tags.
  * https://godoc.org/github.com/hsanjuan/go-nfctype4/apdu : Provides support for creating and serializing Command APDUs and Response APDUs
  * https://godoc.org/github.com/hsanjuan/go-nfctype4/capabilitycontainer : Provides support for creating and serializing Capability Containers and TLV Blocks.
  * https://godoc.org/github.com/hsanjuan/go-nfctype4/drivers/libnfc : Provides libnfc support to read and write to hardware tags with an NFC reader.
  * https://godoc.org/github.com/hsanjuan/go-nfctype4/drivers/swtag : Provides a binary interface for a software `Tag`.
  * https://godoc.org/github.com/hsanjuan/go-nfctype4/tags/static : Provides the implementation of a software-based static NFC Type 4 tag.

