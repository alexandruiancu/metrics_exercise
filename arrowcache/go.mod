module me/arrowcache

go 1.20

require (
	capnproto.org/go/capnp/v3 v3.1.0-alpha.2
	github.com/apache/arrow/go/arrow v0.0.0-20211112161151-bc219186db40
	github.com/pebbe/zmq4 v1.4.0
	me/bldrec v0.0.0-00010101000000-000000000000
)

require (
	github.com/colega/zeropool v0.0.0-20230505084239-6fb4a4f75381 // indirect
	github.com/google/flatbuffers v23.5.26+incompatible // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	golang.org/x/exp v0.0.0-20241009180824-f66d83c29e7c // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gonum.org/v1/gonum v0.16.0 // indirect
)

replace me/common => ../common

replace me/bldrec => ../bldrec
