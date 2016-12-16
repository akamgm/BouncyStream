GOPATH := ${PWD}
export GOPATH

all: install-grpc pbgo
	go build ...

pbgo: 
	protoc -I proto/ proto/bounce.proto --go_out=plugins=grpc:proto

install-grpc:
	go get google.golang.org/grpc
	go get -u github.com/golang/protobuf/{proto,protoc-gen-go}

clean:
	go clean ...
	rm -f proto/*.pb.go

full-clean: clean
	rm -Rf $(GOPATH)/src
	rm -Rf $(GOPATH)/pkg
	rm -Rf $(GOPATH)/bin


