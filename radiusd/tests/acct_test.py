#!/usr/bin/python

import random, socket, sys
import pyrad.packet
from pyrad.client import Client
from pyrad.dictionary import Dictionary

def SendPacket(srv, req):
    try:
        return srv.SendPacket(req)
    except pyrad.client.Timeout:
        print "RADIUS server does not reply"
    except socket.error, error:
        print "Network error: " + error[1]

def test_acct():
    srv=Client(server="127.0.0.1",secret="123456",dict=Dictionary("dictionary"))
    req=srv.CreateAcctPacket(User_Name="wjt001")

    req["NAS-IP-Address"]="192.168.1.10"
    req["NAS-Port"]=0
    req["NAS-Identifier"]="trillian"
    req["Called-Station-Id"]="00-04-5F-00-0F-D1"
    req["Calling-Station-Id"]="00-01-24-80-B3-9C"
    req["Framed-IP-Address"]="10.0.0.100"

    print "Sending accounting start packet"
    req["Acct-Status-Type"]="Start"
    reply = SendPacket(srv, req)
    print reply.code

    print "Sending accounting stop packet"
    req["Acct-Status-Type"]="Stop"
    req["Acct-Input-Octets"] = random.randrange(2**10, 2**30)
    req["Acct-Output-Octets"] = random.randrange(2**10, 2**30)
    req["Acct-Session-Time"] = random.randrange(120, 3600)
    req["Acct-Terminate-Cause"] = random.choice(["User-Request", "Idle-Timeout"])
    reply = SendPacket(srv, req)
    print reply.code

if __name__ == '__main__':
        test_acct()    

