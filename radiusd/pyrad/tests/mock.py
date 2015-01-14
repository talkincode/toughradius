import fcntl
import os
from pyrad.packet import PacketError


class MockPacket:
    reply = object()

    def __init__(self, code, verify=False, error=False):
        self.code = code
        self.data = {}
        self.verify = verify
        self.error = error

    def CreateReply(self, packet=None):
        if self.error:
            raise PacketError
        return self.reply

    def VerifyReply(self, reply, rawreply):
        return self.verify

    def RequestPacket(self):
        return "request packet"

    def __contains__(self, key):
        return key in self.data
    has_key = __contains__

    def __setitem__(self, key, value):
        self.data[key] = [value]

    def __getitem__(self, key):
        return self.data[key]


class MockSocket:
    def __init__(self, domain, type, data=None):
        self.domain = domain
        self.type = type
        self.closed = False
        self.options = []
        self.address = None
        self.output = []

        if data is not None:
            (self.read_end, self.write_end) = os.pipe()
            fcntl.fcntl(self.write_end, fcntl.F_SETFL, os.O_NONBLOCK)
            os.write(self.write_end, data)
            self.data = data
        else:
            self.read_end = 1
            self.write_end = None

    def fileno(self):
        return self.read_end

    def bind(self, address):
        self.address = address

    def recv(self, buffer):
        return self.data[:buffer]

    def sendto(self, data, target):
        self.output.append((data, target))

    def setsockopt(self, level, opt, value):
        self.options.append((level, opt, value))

    def close(self):
        self.closed = True


class MockFinished(Exception):
    pass


class MockPoll:
    results = []

    def __init__(self):
        self.registry = []

    def register(self, fd, options):
        self.registry.append((fd, options))

    def poll(self):
        for result in self.results:
            yield result
        raise MockFinished


def origkey(klass):
    return "_originals_" + klass.__name__


def MockClassMethod(klass, name, myfunc=None):
    def func(self, *args, **kwargs):
        if not hasattr(self, "called"):
            self.called = []
        self.called.append((name, args, kwargs))

    key = origkey(klass)
    if not hasattr(klass, key):
        setattr(klass, key, {})
    getattr(klass, key)[name] = getattr(klass, name)
    if myfunc is None:
        setattr(klass, name, func)
    else:
        setattr(klass, name, myfunc)


def UnmockClassMethods(klass):
    key = origkey(klass)
    if not hasattr(klass, key):
        return
    for (name, func) in getattr(klass, key).items():
        setattr(klass, name, func)

    delattr(klass, key)


class MockFd:
    data = object()
    source = object()

    def __init__(self, fd=0):
        self.fd = fd

    def fileno(self):
        return self.fd

    def recvfrom(self, size):
        self.size = size
        return (self.data, self.source)
