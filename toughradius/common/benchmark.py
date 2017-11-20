#!/usr/bin/env python
# coding: utf-8

import sys
import six
import argparse
import gevent
import logging
import logging.config
from toughradius import settings
import random
from toughradius.pyrad import  statistics
from toughradius.pyrad.radius import packet
from hashlib import md5
import time
import uuid

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

    def __init__(self, userip,interim_update=10,stat=None, args=None):
        self.args = args
        self.stat = stat
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
                session.stop()


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
        auth_req["Service-Type"] = "Login-User"
        auth_req["NAS-Identifier"] = args.nasid
        auth_req["Calling-Station-Id"] = usermac
        auth_req["Framed-IP-Address"] = userip
        resp = radclient.send_auth(args.server,port=args.auth_port,secret=six.b(args.secret),debug=args.debug,timeout=int(args.timeout),result=True, **auth_req)
        self.stat.incr('auth_req')
        if resp is None:
            self.stat.incr('auth_drop')
            return

        if resp != packet.AccessAccept:
            self.stat.incr('auth_reject')

        self.stat.incr('auth_accept')


        acct_session = {
            'User-Name': args.user,
            'Acct-Session-Time': 0,
            'Acct-Status-Type': 1,
            'Session-Timeout': args.session_timeout,
            'Acct-Session-Id': self.session_id,
            "NAS-IP-Address": args.nasaddr,
            "NAS-Port": 0,
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
        acct_reply = radclient.send_acct(args.server,port=args.acct_port,secret=six.b(args.secret),debug=args.debug,timeout=int(args.timeout),result=True, **acct_session)
        self.stat.incr('acct_start')
        if acct_reply is None:
            self.stat.incr('acct_drop')
            return

        if acct_reply.code == packet.AccountingResponse:
            self.stat.incr('acct_resp')
            self.stat.incr('online')
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
            self.session_data["Acct-Output-Octets"] += random.randint(1048576, 10485760)
            self.session_data["Acct-Input-Octets"] += random.randint(1048576, 10485760)
            self.session_data["Acct-Output-Packets"] += random.randint(1024, 8192)
            self.session_data["Acct-Input-Packets"] += random.randint(1024, 8192)
            self.session_data['Acct-Session-Time'] = (int(time.time()) - self.session_start)
            resp = radclient.send_acct(args.server, port=args.acct_port, secret=six.b(args.secret), debug=args.debug,
                                       timeout=int(args.timeout), result=True, **self.session_data)
            self.stat.incr('acct_update')
            if resp is None:
                self.stat.incr('acct_drop')
            elif resp.code == packet.AccountingResponse:
                self.stat.incr('acct_resp')
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
            self.session_data["Acct-Output-Octets"] += random.randint(1048576, 10485760)
            self.session_data["Acct-Input-Octets"] += random.randint(1048576, 10485760)
            self.session_data["Acct-Output-Packets"] += random.randint(1024, 8192)
            self.session_data["Acct-Input-Packets"] += random.randint(1024, 8192)
            self.session_data['Acct-Session-Time'] = (int(time.time()) - self.session_start)
            resp = radclient.send_acct(args.server, port=args.acct_port, secret=six.b(args.secret), debug=args.debug,
                                       timeout=int(args.timeout), result=True, **self.session_data)
            self.stat.incr('acct_stop')
            if resp is None:
                self.stat.incr('acct_drop')
            else:
                self.stat.incr('online', -1)
                self.stat.incr('acct_resp')
            gevent.sleep(0)
        except:
            logger.exception("session stop error")

    def check_session(self):
        session_time = int(time.time()) - self.session_start
        if session_time >= self.session_data['Session-Timeout']:
            self.stop()
        else:
            if not self.running:
                return
            self.update()
            gevent.spawn_later(self.interim_update, self.check_session)


def start():
    parser = argparse.ArgumentParser()
    parser.add_argument('-H','--server', default='127.0.0.1', dest="server",type=str)
    parser.add_argument('--auth-port', default='1812', dest="auth_port",type=int)
    parser.add_argument('--acct-port', default='1813', dest="acct_port",type=int)
    parser.add_argument('-s','--secret', default='testing123', dest="secret",type=str)
    parser.add_argument('-u','--user', default='test01', dest="user",type=str)
    parser.add_argument('-p','--passwd', default='888888', dest="passwd",type=str)
    parser.add_argument('-t','--timeout', default='5', dest="timeout",type=int)
    parser.add_argument('-e','--encrypt', default='pap', dest="encrypt",type=str)
    parser.add_argument('--nasaddr', default='192.168.0.1', dest="nasaddr",type=str)
    parser.add_argument('--nasid', default='toughac', dest="nasid",type=str)
    parser.add_argument('--total', default='1', dest="total",type=int)
    parser.add_argument('--interim', default='10', dest="interim_update",type=int)
    parser.add_argument('--session-timeout', default='120', dest="session_timeout",type=int)
    parser.add_argument('-x','--debug', action='store_true',default=False,dest='debug',help='debug option')
    args = parser.parse_args(sys.argv[1:])
    logger.info(args)

    stat = statistics.MessageStat()

    def run_stat():
        while True:
            stat.run_stat(delay=10)
            logger.info("=========================================")
            logger.info("# auth_req = {0}".format(stat.auth_req))
            logger.info("# auth_accept = {0}".format(stat.auth_accept))
            logger.info("# auth_drop = {0}".format(stat.auth_drop))
            logger.info("# acct_start = {0}".format(stat.acct_start))
            logger.info("# acct_update = {0}".format(stat.acct_update))
            logger.info("# acct_stop = {0}".format(stat.acct_stop))
            logger.info("# acct_resp = {0}".format(stat.acct_resp))
            logger.info("# acct_drop = {0}".format(stat.acct_drop))
            logger.info("=========================================\n\n")
            gevent.sleep(10)

    def start_test():
        for ip in ip_array(args.total):
            session = Session(ip, args.interim_update, stat, args)
            gevent.spawn(session.start)
            gevent.sleep(0.001)


    gevent.spawn(run_stat)
    gevent.spawn(start_test)
    gevent.wait()

if __name__ == "__main__":
    start()



























