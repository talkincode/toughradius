#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys,os
sys.path.insert(0,os.path.split(__file__)[0])
sys.path.insert(0,os.path.abspath(os.path.pardir))
from toughradius.console import models
from toughradius.console.libs import utils
import argparse,ConfigParser

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-c','--conf', type=str,default='../etc/radiusd.conf',dest='conf',help='conf file')
    parser.add_argument('-u','--update',action='store_true',default=False,dest='update',help='update option')
    parser.add_argument('-i','--install',action='store_true',default=False,dest='install',help='install option')
    args =  parser.parse_args(sys.argv[1:])    
    # read config file
    config = ConfigParser.ConfigParser()
    config.read(args.conf)
    utils.aescipher.setup(config.get('DEFAULT','secret'))
    
    if args.update:
        models.update(config)
    elif args.install:
        models.install2(config)
    else:
        models.install(config)

        