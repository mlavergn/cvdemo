#!/bin/bash

curl "http://localhost:8080/suggestions?q=Londo&latitude=43.70011&longitude=-79.4163"
printf "\nexpect baseline\n"

curl "http://localhost:8080/?q=Trois+Rivières&latitude=43.70011&longitude=-79.4163"
printf "\nexpect suggestion\n"

curl "http://localhost:8080/?q=Trois+Rivières&latitude=43.70011&longitude=-79.4163"
printf "\nexpect suggestion\n"

curl "http://localhost:8080/?q=Trois+Rivières"
printf "\n expect suggestion\n"

curl "http://localhost:8080/?q=zTrois+Rivières"
printf "\n expect no suggestion\n"

curl "http://localhost:8080/?q=Trois+Rivières&latitude=43.70011x&longitude=-79.4163x"
printf "\n expect suggestion\n"

curl "http://localhost:8080/?q=Trois+Rivières&latitude=43.70011"
printf "\n expect suggestion\n"

curl "http://localhost:8080/?x=Trois+Rivières"
printf "\n expect no suggestion\n"
