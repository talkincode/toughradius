from twisted.internet import protocol
from twisted.internet import reactor
import os,signal

class MPProtocol(protocol.ProcessProtocol):
    
    def __init__(self,name,log,verb=False):
        self.parent_id = os.getpid()
        self.name = name
        self.log = log
        self.verb = verb

    def connectionMade(self):
        self.log.info("%s created! master pid - %s, worker pid - %s" % \
            (self.name, self.parent_id, self.transport.pid))

    def outReceived(self, data):
        if self.verb:
            self.log.info(data.strip())

    def errReceived(self, data):
        self.log.error(data.strip())

    def processExited(self, reason):
        self.log.info("%s exit, status %s" % (self.name, reason.value.exitCode)) 

    def processEnded(self, reason):
        self.log.info("%s ended, status %s" % (self.name, reason.value.exitCode)) 

class MP:

    def __init__(self,log,verb=False):
        self.log = log
        self.procs = {}
        self.verb = verb

    def spawn(self,name, executable, args=(), env={}, path=None, uid=None, gid=None, usePTY=0, childFDs=None):
        pp = MPProtocol(name,self.log,self.verb)
        ps = self.procs.setdefault(name,[])
        ps.append(pp)
        reactor.spawnProcess(pp,executable, args=args, path=path, env=env,
            uid=uid, gid=gid, usePTY=usePTY, childFDs=childFDs)

    def kill(self,name):
        for pp in self.procs.get(name,[]):
            try:
                pp.transport.signalProcess('TERM')
            except:
                pass

    def killall(self):
        for ppname in self.procs:
            self.kill(ppname)

