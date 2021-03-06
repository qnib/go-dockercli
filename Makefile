all: local alpine linux

local:
	./build.sh

alpine:
	docker run --rm -ti -v $(CURDIR):/go-dockercli/ --workdir /go-dockercli/ qnib/alpn-go-dev ./build.sh

linux:
	docker run --rm -ti -v $(CURDIR):/go-dockercli/ --workdir /go-dockercli/ qnib/golang ./build.sh
  
