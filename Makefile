TELEGRAM=$(shell cat .env | grep TELEGRAM | cut -d '=' -f2)
OPENAI=$(shell cat .env | grep OPENAI | cut -d '=' -f2)

build:
	cd app && go build -mod=vendor -o ../.bin/app

run:
	cd .bin && ./app -v --telegram=$(TELEGRAM) --openai=$(OPENAI)