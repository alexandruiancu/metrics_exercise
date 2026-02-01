module me/laarrow

go 1.24.4

require (
	capnproto.org/go/capnp/v3 v3.1.0-alpha.2
	github.com/apache/arrow/go/arrow v0.0.0-20211112161151-bc219186db40
	github.com/pebbe/zmq4 v1.4.0
	me/bldrec v0.0.0-00010101000000-000000000000
)

require (
	github.com/colega/zeropool v0.0.0-20230505084239-6fb4a4f75381 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

replace me/common => ../common

replace me/bldrec => ../bldrec
