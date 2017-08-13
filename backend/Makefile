GOROOT := /usr/local/go
myhostname := $(shell hostname)
ifeq (${myhostname}, jenkins)
	GOPATH := /var/lib/jenkins/workspace/go/bqserver
endif

build: build-stamp
build-stamp: bqserver.go
	go fmt $^
	go build -o bqserver -v $^
	# docs
	markdown README.mkd > doc/README.html
	help2man --no-info  --no-discard-stderr --name "BQserver" ./bqserver > doc/bqserver.roff
	man -Tps -l doc/bqserver.roff > doc/bqserver-man.ps
	ps2pdf doc/bqserver-man.ps doc/bqserver-man.pdf
	# mark as done
	touch $@

clean:
	go clean
	rm -rf build-stamp

