#!/usr/bin/env python

import RPi.GPIO as GPIO
import time
import requests
import os
import datetime
import json

def sendAlert():
    host = os.uname()[1]
    data = host.split('-')
    
    payload = {
            'building':data[0],
            'room':data[1],
            'cause':'SECURITY',
            'category':'INFO',
            'name':host,
            'timestamp':str(datetime.datetime.now().isoformat('T'))
            }

    headers = {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    }

    address = 'http://dev-elk-shipper0.byu.edu:5546'
    requests.post(address, json = payload, headers = headers)


GPIO.setmode(GPIO.BCM)
GPIO.setup(18, GPIO.IN)

CONTACT_COUNTER = 1200
ALERT_COUNTER = 5

while (True):

    FLAG = 0
    print 'contacts closed'

    while (GPIO.input(18) == 0):

        print 'contacts broken'
        if FLAG == ALERT_COUNTER:

            sendAlert()
            print 'alert sent'
            FLAG += 1

        if FLAG < CONTACT_COUNTER:

            FLAG += 1

        else:

            FLAG = ALERT_COUNTER

        time.sleep(1)

    time.sleep(1)
