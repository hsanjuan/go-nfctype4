Go nfctype4
===========

| Master/stable | Unstable | Reference |
|:-------------:|:--------:|:---------:|
| [![Build Status](https://travis-ci.org/hsanjuan/go-nfctype4.svg?branch=master)](https://travis-ci.org/hsanjuan/go-nfctype4) [![Coverage Status](https://coveralls.io/repos/github/hsanjuan/go-nfctype4/badge.svg?branch=master)](https://coveralls.io/github/hsanjuan/go-nfctype4?branch=master) | [![Build Status](https://travis-ci.org/hsanjuan/go-nfctype4.svg?branch=unstable)](https://travis-ci.org/hsanjuan/go-nfctype4) [![Coverage Status](https://coveralls.io/repos/github/hsanjuan/go-nfctype4/badge.svg?branch=unstable)](https://coveralls.io/github/hsanjuan/go-nfctype4?branch=unstable) | [![GoDoc](https://godoc.org/github.com/hsanjuan/go-ndef?status.svg)](http://godoc.org/github.com/hsanjuan/go-ndef) |

Package `go-nfctype4` implements the NFC Forum Type 4 Tag Operation Specification.

It provides a `Device` type which allows to interact with NFC Devices (like readers) and perform `Read` and `Update` operations on the NFC tags.

It also allows to easily implement software-based NFC Type 4 compliant tags, which can be easily used to provide hardware NFC Readers in target-mode with the necessary functionality to adjust to the specification and act like real Type 4 Tags.

Usage and documentation
-----------------------

```
$ go get github.com/hsanjuan/go-nfctype4
```


```go
import (
	"github.com/hsanjuan/go-nfctype4"
)
```

`go-nfctype4` uses godoc for documentation and examples. You can read it at https://godoc.org/github.com/hsanjuan/go-nfctype4 .

