#!/usr/bin/env python
#coding:utf-8
import sys,os
sys.path.insert(0,os.path.split(__file__)[0])
from sqlalchemy.orm import scoped_session, sessionmaker
from console import models
from radiusd import utils
import argparse
import shutil
import time
import random
import ConfigParser

def gen_secret(clen):
    rg = random.SystemRandom()
    r = list('1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ')
    return ''.join([rg.choice(r) for _ in range(clen)])

def update(conf_file,secret_len=32):
    shutil.copy(conf_file,"%s.%s"%(conf_file,int(time.time())))
    config = ConfigParser.ConfigParser()
    config.read(args.conf)
    # utils.aescipher.setup(config.get('default','secret'))

    old_secret = config.get('default','secret')
    config.set('default','secret',gen_secret(secret_len))

    old_AESCipher = utils.AESCipher(old_secret)
    new_AESCipher = utils.AESCipher(config.get('default','secret'))

    engine,_ = models.get_engine(config)
    conn = engine.connect()
    
    # update 
    db = scoped_session(sessionmaker(bind=engine, autocommit=False, autoflush=True))()  
    user_query = db.query(models.SlcRadAccount.password)
    for user in user_query:
        oldpwd = old_AESCipher.decrypt(user.password)
        user.password = new_AESCipher.encrypt(oldpwd)
        
    vcard_query = db.query(models.SlcRechargerCard.card_passwd)
    for vcard in vcard_query:
        oldpwd = old_AESCipher.decrypt(vcard.card_passwd)
        vcard.card_passwd = new_AESCipher.encrypt(oldpwd)
    
    with open(conf_file,'wb') as configfile:
        config.write(configfile)
        
    db.commit()
    
    print 'update secret success'
    
    
if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('-c','--conf', type=str,default='./radiusd.conf',dest='conf',help='conf file')
    parser.add_argument('-l','--len', type=int,default=32,dest='len',help='secret len')
    args =  parser.parse_args(sys.argv[1:])
    update(args.conf,secret_len=args.len)
