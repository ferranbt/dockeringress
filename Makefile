
build-dev-docker:
	rm -rf ./bin
	go build --mod=vendor -o ./bin/dockeringress main.go
	docker build -t dockeringress:dev .
