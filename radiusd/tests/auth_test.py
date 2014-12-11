#!/usr/bin/python

import socket, sys
import pyrad.packet
from pyrad.client import Client
from pyrad.dictionary import Dictionary

def test_auth():
    srv=Client(server="127.0.0.1",secret="123456",dict=Dictionary("dictionary"))
    req=srv.CreateAuthPacket(code=pyrad.packet.AccessRequest,User_Name="wjt001@cmcc")
    req["NAS-IP-Address"]     = "192.168.1.10"
    req["NAS-Port"]           = 0
    req["Service-Type"]       = "Login-User"
    req["NAS-Identifier"]     = "trillian"
    req["Called-Station-Id"]  = "00-04-5F-00-0F-D1"
    req["Calling-Station-Id"] = "00-01-24-80-B3-9C"
    req["Framed-IP-Address"]  = "10.0.0.100"
    req["User-Password"] = req.PwCrypt('888888')

    try:
        print "Sending authentication request"
        reply=srv.SendPacket(req)
    except pyrad.client.Timeout:
        print "RADIUS server does not reply"
    except socket.error, error:
        print "Network error: " + error[1]

    assert reply.code==pyrad.packet.AccessAccept

    print "Attributes returned by server:"
    for i in reply.keys():
        print "%s: %s" % (i, reply[i])

if __name__ == '__main__':
    test_auth()
