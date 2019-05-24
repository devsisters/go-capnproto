.PHONY: prepare

prepare:
	go install github.com/devsisters/go-capnproto/capnpc-go
	cd aircraftlib && make


check:
	cat data/check.zdate.cpz | capnp decode aircraftlib/aircraft.capnp  Zdate 

checkp:
	cat data/zdate2.packed.dat | bin/decp

testbuild:
	go test -c -gcflags "-N -l" -v

clean:
	rm -f go-capnproto.test *~
	cd aircraftlib; make clean

test:
	cd capnpc-go; go build; go install
	cd aircraftlib; make
	go test -v

