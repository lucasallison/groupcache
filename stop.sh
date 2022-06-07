#!/bin/bash

pid=$(lsof -i:8081 -t) 
if [[ $pid -gt 0 ]]; then
	kill -TERM $pid || kill -KILL $pid
	echo "Stopped cache running on port 8081 with pid $pid"
fi 

pid=$(lsof -i:8082 -t); 
if [[ $pid -gt 0 ]]; then
	kill -TERM $pid || kill -KILL $pid
	echo "Stopped cache running on port 8082 with pid $pid"
fi

pid=$(lsof -i:8083 -t); 
if [[ $pid -gt 0 ]]; then
	kill -TERM $pid || kill -KILL $pid
	echo "Stopped cache running on port 8083 with pid $pid"
fi

