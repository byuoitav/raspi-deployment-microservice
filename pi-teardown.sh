#!/usr/bin/env bash

apt-get install purge awesome
apt-get install purge chromium-browser

apt-get -y autoremove
apt-get -y autoclean
