#!/usr/bin/env bash

# requires firefox and xdotool to be installed.
# install both with `apt install firefox-esr xdotool`

# open firefox with a private window
firefox --private-window http://localhost:8888 &

# enable fullscreen
sleep 20
xdotool key F11
