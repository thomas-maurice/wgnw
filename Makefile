all:
	go vet ./...
	go test ./...
	if ! [ -d bin ]; then mkdir bin; fi;
	cd agent && go build -o ../bin/wgnwd
	cd cli && go build -o ../bin/wgnw
	cd server && go build -o ../bin/wgnw-server

gen:
	go generate ./...

clean:
	rm -rf bin
