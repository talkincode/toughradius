#!/usr/bin/env python
#coding:utf-8
from toughradius.console import models
from sqlalchemy.orm import scoped_session, sessionmaker
from toughradius.radiusd import utils
from toughradius.radiusd import pyrad
from toughradius.radiusd.pyrad import dictionary
from toughradius.radiusd.pyrad.client import Client 
from toughradius.radiusd.pyrad import packet
from toughradius.tools.dbengine import get_engine
import threading
import os,sys
import six
import random
import datetime
import time
import socket

status_vars = {'start':1,'stop':2,'update':3,'on':7,'off':8}

# random_sleeps = (0.01,0.02,0.03,0.04,0.05)
random_sleeps = (1.1,1.2,1.3,0.6,0.1,0.2)

# random_stop_sleeps = [ random.randrange(128,1200) for n in range(36) ]


class Connect:
    def __init__(self, mkdb):
        self.conn = mkdb()

    def __enter__(self):
        return self.conn   

    def __exit__(self, exc_type, exc_value, exc_tb):
        self.conn.close()
        
class RadSession(object):
    
    base_id = 0

    def __init__(self,user,session_id=None):
        if session_id is None:
            self.session_id = self.gen_id()
        else:
            self.session_id = session_id
        self.user = user
        self.start_time = int(time.time())
        self.last_update = int(time.time())
        self.session_time = 0
        self.stop_time = 0
        self.input_total = 0
        self.output_total = 0
        self.input_pkts = 0
        self.output_pkts = 0
        
    def update(self):
        self.input_total += random.randrange(2**10, 2**20)
        self.output_total += random.randrange(2**10, 2**20)
        self.input_pkts = random.randrange(1024, 81920)
        self.output_pkts = random.randrange(1024, 81920)
        _now = int(time.time())
        self.session_time = _now - self.last_update
        self.last_update = _now
        
            
    def gen_id(self):
        if self.base_id >= 9999:
            self.base_id=0
        self.base_id += 1
        _num = str(self.base_id).zfill(4)
        return datetime.datetime.now().strftime("%Y%m%d%H%M%S") + _num
        

class RadClient(object):
    
    def __init__(self,config):
        auth_port = config.getint('radiusd','authport')
        acct_port = config.getint('radiusd','acctport')
        rad_addr = '127.0.0.1'
        if config.has_option('radiusd','host'):
            rad_addr = config.getint('radiusd','host')
        dictfile = os.path.join(os.path.split(utils.__file__)[0],'dicts/dictionary')
        _dict = dictionary.Dictionary(dictfile)
        secret = six.b('radtest')
        self.debug = config.getboolean('DEFAULT','debug')
        self.radcli = Client(rad_addr,auth_port,acct_port,secret,_dict)
        self.radius_sessions = {}
        self.user_sessions = {}

    def send_pkt(self,req):
        try:
            print "Sending a radius request"
            
            if self.debug:
                req.source = ('127.0.0.1',0)
                print utils.format_packet_str(req)
            reply=self.radcli.SendPacket(req)
            print "recv a radius request"
            if self.debug:
                reply.source =  ('127.0.0.1',0)
                print utils.format_packet_str(reply)
            return reply
        except pyrad.client.Timeout:
            print "RADIUS server does not reply"
        except socket.error, error:
            print "Network error: " + error[1]

    def send_auth(self,user,pwd):
        req = self.radcli.CreateAuthPacket(
            code=pyrad.packet.AccessRequest,
            User_Name=user
        )
        req['User-Password'] = req.PwCrypt(pwd)
        req["NAS-IP-Address"]     = "192.168.88.10"
        req["Framed-IP-Address"]     = "192.168.88.11"
        req["NAS-Port"]           = 0
        req["Service-Type"]       = "Login-User"
        req["NAS-Identifier"]     = "radtest"
        req["Called-Station-Id"]  = "00-04-5F-00-0F-D1"
        req["Calling-Station-Id"] = "00-01-24-80-B3-9C"
        req['NAS-Port-Id'] = '3/0/1:0.0'

        return self.send_pkt(req)
            
    def send_acct_start(self,user):
        _session = RadSession(user)
        req=self.radcli.CreateAcctPacket(User_Name=user)
        print "Sending accounting start packet"
        req["NAS-IP-Address"]="192.168.88.10"
        req["Framed-IP-Address"] = "192.168.88.11"
        req["NAS-Port"]=0
        req["NAS-Identifier"]="radtest"
        req["Called-Station-Id"]="00-04-5F-00-0F-D1"
        req["Calling-Station-Id"]="00-01-24-80-B3-9C"
        req["Acct-Status-Type"]="Start"
        req['Acct-Session-Id'] = _session.session_id
        reply = self.send_pkt(req)
        self.radius_sessions[_session.session_id] = _session
        if user not in self.user_sessions:
            self.user_sessions[user] = []
        self.user_sessions[user].append(_session.session_id)
        return reply

    def send_acct_update(self,session):
        session.update()
        req=self.radcli.CreateAcctPacket(User_Name=session.user)
        print "Sending accounting update packet"
        req["NAS-IP-Address"]="192.168.88.10"
        req["NAS-Port"]=0
        req["NAS-Identifier"]="radtest"
        req["Framed-IP-Address"] = "192.168.88.11"
        req["Called-Station-Id"]="00-04-5F-00-0F-D1"
        req["Calling-Station-Id"]="00-01-24-80-B3-9C"
        req["Acct-Status-Type"]="Alive"
        req["Acct-Input-Octets"] = session.input_total
        req["Acct-Output-Octets"] = session.output_total
        req['Acct-Input-Packets'] = session.input_pkts
        req['Acct-Output-Packets'] = session.output_total
        req["Acct-Session-Time"] = session.session_time
        req['Acct-Session-Id'] = session.session_id
        return self.send_pkt(req)
        
    def send_acct_stop(self,session):
        session.update()
        req=self.radcli.CreateAcctPacket(User_Name=session.user)
        print "Sending accounting stop packet"
        req["NAS-IP-Address"]="192.168.88.10"
        req["NAS-Port"]=0
        req["NAS-Identifier"]="radtest"
        req["Framed-IP-Address"] = "192.168.88.11"
        req["Called-Station-Id"]="00-04-5F-00-0F-D1"
        req["Calling-Station-Id"]="00-01-24-80-B3-9C"
        req["Acct-Status-Type"]="stop"
        req["Acct-Input-Octets"] = session.input_total
        req["Acct-Output-Octets"] = session.output_total
        req['Acct-Input-Packets'] = session.input_pkts
        req['Acct-Output-Packets'] = session.output_total
        req["Acct-Session-Time"] = session.session_time
        req['Acct-Session-Id'] = session.session_id
        reply = self.send_pkt(req)
        del self.radius_sessions[session_id]
        if session.user not in self.user_sessions:
            self.user_sessions[sessiuon.user] = []
        self.user_sessions[session.user].remove(session.session_id)
        return reply

class Tester(object):
    
    def __init__(self,config):
        self.config = config
        self.mkdb = scoped_session(sessionmaker(
            bind=get_engine(config), autocommit=False, autoflush=True
        ))
        self.load_users()
        self.radcli = RadClient(config)
        self.running = False
        self.threads = []
        utils.aescipher.setup(config.get('DEFAULT','secret'))
        
    def load_users(self):
        with Connect(self.mkdb) as db:
            self.accounts = db.query(models.SlcRadAccount).all()
    
    def do_start(self):
        while self.running:
            user = random.choice(self.accounts)
            reply = self.radcli.send_auth(user.account_number,utils.decrypt(user.password))
            if reply.code == packet.AccessAccept:
                reply = self.radcli.send_acct_start(user.account_number)
            
            time.sleep(random.choice(random_sleeps))
            
    def do_update(self):
        while self.running:
            items = self.radcli.radius_sessions.values()
            if len(items) > 0:
                session = random.choice(items)
                reply = self.radcli.send_acct_update(session)
            time.sleep(random.choice(random_sleeps))
            
    def do_stop(self):
        while self.running:
            items = self.radcli.radius_sessions.values()
            if len(items) > 0:
                session = random.choice(items)
                reply = self.radcli.send_acct_stop(session)
            time.sleep(random.choice(random_sleeps))

    def start(self):
        self.running = True
        self.threads.append(threading.Thread(target=self.do_start))
        self.threads.append(threading.Thread(target=self.do_update))
        self.threads.append(threading.Thread(target=self.do_stop))
        
        # 启动线程
        for t in self.threads:
            t.start()
            
        while 1:
            if raw_input('type q to exit:') == 'q':
                self.running = False
                sys.exit(0)


            

        
            
     
            
      
    
    
        
    