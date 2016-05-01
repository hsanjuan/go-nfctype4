Go-nfctype4
===========

| Master/stable | Unstable | Reference |
|:-------------:|:--------:|:---------:|
| [![Build Status](https://travis-ci.org/hsanjuan/go-nfctype4.svg?branch=master)](https://travis-ci.org/hsanjuan/go-nfctype4) [![Coverage Status](https://coveralls.io/repos/github/hsanjuan/go-nfctype4/badge.svg?branch=master)](https://coveralls.io/github/hsanjuan/go-nfctype4?branch=master) | [![Build Status](https://travis-ci.org/hsanjuan/go-nfctype4.svg?branch=unstable)](https://travis-ci.org/hsanjuan/go-nfctype4) [![Coverage Status](https://coveralls.io/repos/github/hsanjuan/go-nfctype4/badge.svg?branch=unstable)](https://coveralls.io/github/hsanjuan/go-nfctype4?branch=unstable) | [![GoDoc](https://godoc.org/github.com/hsanjuan/go-nfctype4?status.svg)](http://godoc.org/github.com/hsanjuan/go-nfctype4) |

Package `go-nfctype4` implements the NFC Forum Type 4 Tag Operation Specification Version 2.0.

The implementation acts as an NFC Forum Device (an entity that can read and write NFC Forum Tags). For such. it provides a `Device` type which can `Read`, `Update` and `Format` NFC tags.

The module and submodules contain all the pieces to impelment software-based NFC Type 4 Tags as well. For more information about this check the documentation. You can also check out this [snippet](https://gitlab.com/snippets/18718) showing how this is done using the static tag implementation provided.

nfctype4-tool
-------------

`nfctype4-tool` is a command-line tool to read and write NFC Type 4 tags. It can be installed with:

`go install github.com/hsanjuan/go-nfctype4/nfctype4-tool`

You can then run `nfctype4-tool -h` to get going.

Note: to turn a Mifare Desfire EV2 (4k) card into an NFC Type 4 Tag check: https://gitlab.com/snippets/18476 .

Packages
--------

`go-nfctype4` and its subpackages offer access to the implementation of all the entities described in the NFC Forum Type 4 Tag specification. These are the links to the reference documentation of the most relevant packages:

  * https://godoc.org/github.com/hsanjuan/go-nfctype4 : Provides the `Device`, `CommandDriver` and `Commander`. They are the main entry point to interact with NFC tags.
  * https://godoc.org/github.com/hsanjuan/go-nfctype4/apdu : Provides support for creating and serializing Command APDUs and Response APDUs.
  * https://godoc.org/github.com/hsanjuan/go-nfctype4/capabilitycontainer : Provides support for creating and serializing Capability Containers and TLV Blocks.
  * https://godoc.org/github.com/hsanjuan/go-nfctype4/drivers/libnfc : Provides libnfc support to read and write to hardware tags with an NFC reader.
  * https://godoc.org/github.com/hsanjuan/go-nfctype4/drivers/swtag : Provides a binary interface for a software `Tag`.
  * https://godoc.org/github.com/hsanjuan/go-nfctype4/tags : Provides the `Tag` interface, on which software tags that use this library should be based-on.
  * https://godoc.org/github.com/hsanjuan/go-nfctype4/tags/static : Provides the implementation of a software-based static NFC Type 4 tag.

