import redis
import inspect
from bottle import PluginError
from gevent import socket
redis.connection.socket = socket

class RedisPlugin(object):
    name = 'redis'
    api =2

    def __init__(self, host='localhost', port=6379, database=0, keyword='rdb'):
      self.host = host
      self.port = port
      self.database = database
      self.keyword = keyword
      self.redisdb = None

    def setup(self, app):
        for other in app.plugins:
            if not isinstance(other, RedisPlugin): continue
            if other.keyword == self.keyword:
                raise PluginError("Found another redis plugin with "\
                        "conflicting settings (non-unique keyword).")
        if self.redisdb is None:
            self.redisdb = redis.ConnectionPool(host=self.host, port=self.port, db=self.database)

    def apply(self, callback, route):
        conf = route.config.get('redis') or {}
        args = inspect.getargspec(route.callback)[0]
        keyword = conf.get('keyword', self.keyword)
        if keyword not in args:
            return callback

        def wrapper(*args, **kwargs):
            kwargs[self.keyword] = redis.StrictRedis(connection_pool=self.redisdb)
            rv = callback(*args, **kwargs)
            return rv
        return wrapper

Plugin = RedisPlugin