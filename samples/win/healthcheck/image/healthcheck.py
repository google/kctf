#!/usr/bin/env python3

from http import server
import threading
import time
import requests

healthy = False

class Handler(server.BaseHTTPRequestHandler):
  def __init__(self, request, client_address, server):
    super().__init__(request, client_address, server)

  def do_GET(self):
    if healthy:
      response_code = 200
      content = b'ok'
    else:
      response_code = 400
      content = b'err'
    self.send_response(response_code)
    self.send_header("Content-type", "text/plain")
    self.send_header("Content-length", str(len(content)))
    self.end_headers()
    self.wfile.write(content)

def check_health():
  global healthy
  while True:
    try:
      text = requests.get('http://127.0.0.1/').text
      if text.strip() == 'kCTF':
        healthy = True
        print('healthy')
      else:
        healthy = False
        print('== ERROR ==')
        print(text)
    except:
      healthy = False
      print('== exception ==')
    time.sleep(20)

checker_thread = threading.Thread(target=check_health)
checker_thread.start()

server_address = ('', 45281)
httpd = server.HTTPServer(server_address, Handler)
httpd.serve_forever()
