vendor:
	go mod tidy
	go mod vendor

build:
	cd app && go build -mod=vendor -o ../.bin/app

run:
	cd .bin && ./app