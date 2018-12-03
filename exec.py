
from flask import request
import sys
import os
from flask import Flask
app = Flask(__name__)
dir = os.environ['CURRENT']

def shutdown_server():
    func = request.environ.get('werkzeug.server.shutdown')
    if func is None:
        raise RuntimeError('Not running with the Werkzeug Server')
    func()


@app.route('/<path:mypath>', methods=['GET'])
def route(mypath):
    if mypath.endswith('.geojson'):
        shutdown_server()
        return open(mypath,'rb').read()

    if mypath.endswith('index.html'):
        return open(mypath,'rb').read()
    return ''
  