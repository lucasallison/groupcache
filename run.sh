#!/bin/bash

if [[ $1 == "-poolcount=3" ]]; then 
	echo "Supplied Pool Count: 3"
	echo "Running three caches on ports 8081, 8082 and 8083"
	go run . -addr=:8081 -pool=http://127.0.0.1:8081,http://127.0.0.1:8082,http://127.0.0.1:8083 &
	go run . -addr=:8082 -pool=http://127.0.0.1:8082,http://127.0.0.1:8081,http://127.0.0.1:8083 &
	go run . -addr=:8083 -pool=http://127.0.0.1:8083,http://127.0.0.1:8081,http://127.0.0.1:8082 &
elif [[ $1 == "-poolcount=2" ]]; then
	echo "Supplied Pool Count: 2"
	echo "Running two caches on ports 8081 and 8082"
	go run . -addr=:8081 -pool=http://127.0.0.1:8081,http://127.0.0.1:8082 &
	go run . -addr=:8082 -pool=http://127.0.0.1:8082,http://127.0.0.1:8081 &
else 
	echo "Supplied Pool Count: 1"
	echo "Running a single cache on port 8081"
	go run . -addr=:8081 -pool=http://127.0.0.1:8081 &
fi

