#!/bin/bash

export DISPLAY=:0
xssstart curl -X PUT http://localhost:8888/screenoff &

echo "Waiting for screenoff events."
