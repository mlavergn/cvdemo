###############################################
#
# Makefile
#
###############################################

data:
	rm -f cities_canada-usa.tsv
	curl -L -H "Accept-Charset: utf-8" -o cities_canada-usa.tsv https://goo.gl/zjxkej

run:
	python endpoint.py

demo:
	open "http://localhost:8080/?q=London&latitude=43.70011&longitude=-79.4163"

gorun:
	go run endpoint.go

build:
	go build endpoint.go

setup:
	go get golang.org/x/text/runes
	go get golang.org/x/text/transform
	go get golang.org/x/text/unicode/norm

test:
	go test
