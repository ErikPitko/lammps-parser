exec=lammps-parser

ENV_WIN=CGO_ENABLED="1" CGO_CFLAGS="-O2" CGO_CXXFLAGS="-O2" CGO_FFLAGS="-O2" CGO_LDFLAGS="-O2" GOOS=windows GOARCH=amd64

all: build

build: cmd/main.go
	mkdir -p bin
	go build -o bin/${exec} $<
	# upx ${exec}

build-win: cmd/main.go
	mkdir -p bin
	$(ENV_WIN) go build -o bin/${exec}.exe $<
	# upx ${exec}.exe
