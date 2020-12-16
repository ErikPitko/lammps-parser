exec=lammps-parser

all: build

build: cmd/main.go
	mkdir -p bin
	go build -o bin/${exec} $<
	# upx ${exec}

build-win: cmd/main.go
	mkdir -p bin
	CGO_ENABLED="1" CGO_CFLAGS="-O2" CGO_CXXFLAGS="-O2" CGO_FFLAGS="-O2" CGO_LDFLAGS="-O2" GOOS=windows GOARCH=amd64 go build -o bin/${exec}.exe $<
	# upx ${exec}.exe
