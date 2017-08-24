#!/usr/bin/env python

import os
import re
import sys
import json
import requests
from BaseHTTPServer import HTTPServer, BaseHTTPRequestHandler
from termcolor import colored, cprint

PORT = 2000

class RequestHandler(BaseHTTPRequestHandler):

    def do_GET(request):

        if None != re.search('/deploy/*', request.path):

            device = request.path.split('/')[2]
            print 'deploying to device:', device

            url = os.environ['RASPI_DEPLOYMENT_MICROSERVICE_WSO2_ADDRESS'] + '_device/' + device
            headers = {
                    "Accept": "application/json",
                    "Authorization": os.environ['WSO2_TOKEN']
                    }
            print 'sending request to', url
            print 'headers:', headers

            response = requests.get(url, headers=headers)
            
            if response.status_code != 200:
                msg = 'non-200 response: %f' % response.status_code 
                cprint(msg, 'red', attrs=['bold'], file=sys.stderr)
                request.send_response(500)
                request.end_headers()
                request.wfile.write("failed to deploy to device")
                return

            result = json.loads(response.text)
            print result['response']
                    

            request.send_response(200)
            request.end_headers()
            request.wfile.write("successfully triggered deployment")
    
        else:
            request.send_response(400)
            request.end_headers()
            request.wfile.write('please make a request to /deploy/:hostname')




server = HTTPServer(('', PORT), RequestHandler)

msg = colored(str(PORT), 'green', attrs=['bold'])
print 'HTTP server started on port', msg

server.serve_forever();



