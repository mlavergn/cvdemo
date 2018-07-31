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