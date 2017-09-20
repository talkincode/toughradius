#coding:utf-8
from bottle import ServerAdapter
import threading
import gevent
import os

class GeventServer(ServerAdapter):
    #
    def __init__(self, host='127.0.0.1', port=8080, **options):
        self.forever = options.pop('forever',True)
        self.options = options
        self.host = host
        self.port = int(port)

    def run(self, handler):
        from gevent import wsgi, pywsgi, local
        if not isinstance(threading.local(), local.local):
            msg = "Bottle requires gevent.monkey.patch_all() (before import)"
            raise RuntimeError(msg)
        if not self.options.pop('fast', None): wsgi = pywsgi
        self.options['log'] = None if self.quiet else 'default'
        address = (self.host, self.port)
        server = wsgi.WSGIServer(address, handler, **self.options)
        # if 'BOTTLE_CHILD' in os.environ:
        #     import signal
        #     signal.signal(signal.SIGINT, lambda s, f: server.stop())
        if self.forever:
            server.serve_forever()
        else:
            if not server.started:
                server.start()
