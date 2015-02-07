#!/usr/bin/env python
# -*- coding: utf-8 -*-
import sys,os
sys.path.insert(0,os.path.split(__file__)[0])
from console import models
from console.libs import utils
import argparse,json

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-c','--conf', type=str,default='./config.json',dest='conf',help='conf file')
    parser.add_argument('-u','--update',nargs='?', type=bool,default=False,dest='update',help='update option')
    parser.add_argument('-i','--install',nargs='?', type=bool,default=False,dest='install',help='install option')
    parser.add_argument('-t','--test',nargs='?', type=bool,default=False,dest='test',help='install test data')
    args =  parser.parse_args(sys.argv[1:])    
    config =json.loads(open(args.conf,'rb').read())
    utils.update_secret(config['secret'])
    dbconf = config['database']
    if args.update:
        models.update(dbconf)
    elif args.install:
        models.install2(dbconf)
    else:
        models.install(dbconf)
    if args.test:
        models.install_test(dbconf)
        