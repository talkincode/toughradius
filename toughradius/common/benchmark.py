#!/usr/bin/env python
# coding: utf-8

import argparse
import logging.config
import random
import sys
import time
import uuid
import gevent
from gevent.pool import Pool
from toughradius import settings
from toughradius.pyrad import statistics
from toughradius.pyrad.radius import packet
from toughradius.common import six

logging.config.dictConfig(settings.LOGGER)

logger = logging.getLogger()
import radclient

def gen_mac(ip):
    return ':'.join(map(lambda x: "%02x" % x, [0x52, 0x54]+[int(i) for i in ip.split('.')]))

def ip_array(ilen):
    if ilen > 100000:
        ilen = 100000
    count = 0
    for n in range(ilen):
        for i in range(255):
            for k in range(255):
                for j in range(255):
                    yield '.'.join(['10', str(i), str(k), str(j)])
                    count += 1
                    if count == ilen:
                        return

class Session:

    sessions = {}

    def __init__(self, userip,interim_update=10,stat=None, pool=None, args=None):
        self.args = args
        self.stat = stat
        self.pool = pool
        self.running = False
        self.session_start = int(time.time())
        self.session_id = uuid.uuid1().hex.upper()
        self.session_data = {}
        self.interim_update = interim_update
        self.userip = userip

    @staticmethod
    def stop_session(session_id=None):
        if session_id:
            session = Session.sessions.pop(session_id,None)
            if session:
                session.pool.spawn(session.stop)
        else:
            for _, session in Session.sessions.iteritems():
                session.pool.spawn(session.stop)




    def start(self):
        args = self.args
        userip = self.userip
        usermac = gen_mac(userip)
        auth_req = {'User-Name': args.user}
        if args.encrypt == "pap":
            auth_req['User-Password'] = args.passwd
        elif args.encrypt == "chap":
            auth_req['CHAP-Password-Plaintext'] = args.passwd

        auth_req["NAS-IP-Address"] = args.nasaddr
        auth_req["NAS-Port"] = 0
        auth_req["Framed-Protocol"] = "PPP"
        auth_req["Service-Type"] = "Login-User"
        auth_req["NAS-Identifier"] = args.nasid
        auth_req["Calling-Station-Id"] = usermac
        auth_req["Framed-IP-Address"] = userip
        resp = radclient.send_auth(args.server, port=args.auth_port, secret=six.b(args.secret), debug=args.debug,
                                   timeout=int(args.timeout), result=True, stat=self.stat, **auth_req)
        if resp is None or resp.code != packet.AccessAccept:
            return

        acct_session = {
            'User-Name': args.user,
            'Acct-Session-Time': 0,
            'Acct-Status-Type': 1,
            'Session-Timeout': args.session_timeout,
            'Acct-Session-Id': self.session_id,
            "NAS-IP-Address": args.nasaddr,
            "NAS-Port": 0,
            "Event-Timestamp" : int(time.time()),
            "NAS-Identifier": args.nasid,
            "Calling-Station-Id": usermac,
            "Framed-IP-Address": userip,
            "Acct-Output-Octets": 0,
            "Acct-Input-Octets": 0,
            "Acct-Output-Packets": 0,
            "Acct-Input-Packets": 0,
            "NAS-Port-Id": "3/0/1:0.0"
        }
        self.session_data.update(acct_session)
        acct_reply = radclient.send_acct(args.server, port=args.acct_port, secret=six.b(args.secret), debug=args.debug,
                                         timeout=int(args.timeout), result=True, stat=self.stat, **acct_session)

        if acct_reply is not None and acct_reply.code == packet.AccountingResponse:
            self.running = True
            if args.debug:
                logger.info('start session  %s %s' % (args.user, self.session_id))
            Session.sessions[self.session_id] = self
            gevent.spawn_later(self.interim_update,self.check_session)

    def update(self):
        try:
            args = self.args
            if args.debug:
                logger.info('update session  %s %s' % (args.user, self.session_id))
            self.session_data['Acct-Status-Type'] = 3
            self.session_data["Event-Timestamp"] = int(time.time())
            self.session_data["Acct-Output-Octets"] += random.randint(1048576, 10485760)
            self.session_data["Acct-Input-Octets"] += random.randint(1048576, 10485760)
            self.session_data["Acct-Output-Packets"] += random.randint(1024, 8192)
            self.session_data["Acct-Input-Packets"] += random.randint(1024, 8192)
            self.session_data['Acct-Session-Time'] = (int(time.time()) - self.session_start)
            radclient.send_acct(args.server, port=args.acct_port, secret=six.b(args.secret), debug=args.debug,
                                       timeout=int(args.timeout), result=True,stat=self.stat, **self.session_data)
            gevent.sleep(0)
        except:
            logger.exception("session update error")

    def stop(self):
        try:
            args = self.args
            if args.debug:
                logger.info('Stop session  %s' % self.session_id)
            self.running = False
            self.session_data['Acct-Status-Type'] = 2
            self.session_data["Event-Timestamp"] = int(time.time())
            self.session_data["Acct-Output-Octets"] += random.randint(1048576, 10485760)
            self.session_data["Acct-Input-Octets"] += random.randint(1048576, 10485760)
            self.session_data["Acct-Output-Packets"] += random.randint(1024, 8192)
            self.session_data["Acct-Input-Packets"] += random.randint(1024, 8192)
            self.session_data['Acct-Session-Time'] = (int(time.time()) - self.session_start)
            radclient.send_acct(args.server, port=args.acct_port, secret=six.b(args.secret), debug=args.debug,
                                       timeout=int(args.timeout), result=True, stat=self.stat, **self.session_data)
            gevent.sleep(0)
        except:
            logger.exception("session stop error")

    def check_session(self):
        session_time = int(time.time()) - self.session_start
        if session_time >= self.session_data['Session-Timeout']:
            self.pool.spawn(self.stop)
        else:
            if not self.running:
                return
            self.pool.spawn(self.update)
            gevent.spawn_later(self.interim_update, self.check_session)


def start():
    parser = argparse.ArgumentParser()
    parser.add_argument('-H', '--server', default='127.0.0.1', dest="server",type=str, help="radius server ipaddr")
    parser.add_argument('--auth-port', default='1812', dest="auth_port", type=int)
    parser.add_argument('--acct-port', default='1813', dest="acct_port", type=int)
    parser.add_argument('-s', '--secret', default='secret', dest="secret", type=str, help="radius share key")
    parser.add_argument('-u', '--user', default='test01', dest="user",type=str, help="auth user")
    parser.add_argument('-p', '--passwd', default='888888', dest="passwd",type=str)
    parser.add_argument('-t', '--timeout', default='5', dest="timeout",type=int, help="radius request timeout")
    parser.add_argument('-e', '--encrypt', default='pap', dest="encrypt",type=str, help="radius valid type pap|chap")
    parser.add_argument('--nasaddr', default='192.168.0.1', dest="nasaddr",type=str, help="nas server ipaddr")
    parser.add_argument('--nasid', default='toughac', dest="nasid",type=str, help="NAS-Identifier value")
    parser.add_argument('--total', default='1', dest="total",type=int, help="test user count")
    parser.add_argument('--pool', default='32', dest="pool_size",type=int, help="Concurrent pool size")
    parser.add_argument('--interim', default='5', dest="interim_update",type=int, help="radius interim_update interval")
    parser.add_argument('--session-timeout', default='60', dest="session_timeout",type=int, help="Radius session_timeout")
    parser.add_argument('--stat-interval', default='1', dest="stat_interval",type=int, help="Statistical interval")
    parser.add_argument('-x','--debug', action='store_true',default=False,dest='debug', help='debug option')
    args = parser.parse_args(sys.argv[1:])

    logger.info(args)

    starttime = time.time()

    stat = statistics.MessageStat(quemax=args.session_timeout)
    pool = Pool(args.pool_size)

    def run_stat():
        exit_couter = 0
        while True:
            stat.run_stat(delay=args.stat_interval)
            loginfo = []
            loginfo.append("#" * 80)
            loginfo.append("# - TEST : total={0}, interim_update={1} sec, session_timeout={2} sec".format(args.total, args.interim_update, args.session_timeout).ljust(79,' ')+'#')
            loginfo.append("# - Authentication : req={0}, accept={1}, drop={2}".format(stat.auth_req, stat.auth_accept, stat.auth_drop).ljust(79,' ')+'#')
            loginfo.append("# - Accounting : start={0}, update={1}, stop={2}, resp={3}, drop={4} ".format(
                stat.acct_start, stat.acct_update, stat.acct_stop, stat.acct_resp, stat.acct_drop).ljust(79,' ')+'#')
            loginfo.append("# - Max Request per second : {0} [#/sec] at {1}".format(stat.last_max_req, stat.last_max_req_date).ljust(79,' ')+'#')
            loginfo.append("# - Max Response per second : {0} [#/sec] at {1}".format(stat.last_max_resp, stat.last_max_resp_date).ljust(79,' ')+'#')
            loginfo.append("# - Send Request length : {0} kbytes [ max {1} kbytes/sec ]".format(
                round(stat.req_bytes/1024.0, 2),round(stat.last_max_req_bytes/1024.0, 2)).ljust(79,' ')+'#')
            loginfo.append("# - Send Response length : {0} kbytes [ max {1} kbytes/sec ]".format(
                round(stat.resp_bytes/1024.0, 2), round(stat.last_max_resp_bytes/1024.0, 2)).ljust(79,' ')+'#')
            loginfo.append("# - Running times : {0} sec".format(round(time.time()-starttime, 2)).ljust(79,' ')+'#')
            loginfo.append("# - Current online : {0}".format(stat.online).ljust(79,' ')+'#')
            loginfo.append("# - Current Pool : total={0}, used={1}, free={2}".format(pool.size, len(pool), pool.free_count()).ljust(79,' ')+'#')
            loginfo.append("#" * 80)
            logger.info("Radius Statistics Data ::::::::::::: \n" + '\n'.join(loginfo) + "\n\n")

            if stat.online == 0:
                exit_couter += 1
            if exit_couter >= 5:
                break
            gevent.sleep(args.stat_interval)

    def start_test():
        for ip in ip_array(args.total):
            session = Session(ip, args.interim_update, stat, pool, args)
            pool.spawn(session.start)
            gevent.sleep(0)

    gevent.spawn_later(args.stat_interval, run_stat)
    pool.spawn(start_test)
    gevent.wait()

if __name__ == "__main__":
    start()



























