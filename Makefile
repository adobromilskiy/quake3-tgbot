TELEGRAM=$(shell cat .env | grep TELEGRAM | cut -d '=' -f2)
OPENAI=$(shell cat .env | grep OPENAI | cut -d '=' -f2)
Q3SERV=$(shell cat .env | grep Q3SERVURL | cut -d '=' -f2)
CHATID=$(shell cat .env | grep CHATID | cut -d '=' -f2)

build:
	cd app && go build -mod=vendor -o ../.bin/app

run:
	cd .bin && ./app -v --telegram=$(TELEGRAM) --openai=$(OPENAI) --server=$(Q3SERV) --chat=$(CHATID)

image:
	docker build -t q3bot .

container:
	docker run -it --rm q3bot -v --telegram=$(TELEGRAM) --openai=$(OPENAI) --server=$(Q3SERV) --chat=$(CHATID)