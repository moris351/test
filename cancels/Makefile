# This is how we want to name the binary output
OUTPUT=cancels
# These are the values we want to pass for Version and BuildTime
GITTAG=`git describe --tags`
BUILD_TIME=`date +%FT%T%z`
# Setup the -ldflags option for go build here, interpolate the variable values

VERTAG=
ifdef vertag
	ifeq (${vertag},fromgit)
		VERTAG=${GITTAG}
	else 
		VERTAG=${vertag}
	endif
endif

LDFLAGS=-ldflags "-X main.VerTag=${VERTAG} -X main.BuildTime=${BUILD_TIME}"
all:
	go build ${LDFLAGS} -o ${OUTPUT} main.go

