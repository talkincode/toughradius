#!/usr/bin/env python
#coding:utf-8
import sys,os
from sqlalchemy.orm import scoped_session, sessionmaker
from toughradius.console import models
from toughradius.radiusd import utils
from toughradius.tools.dbengine import get_engine
import shutil
import time
import random
import ConfigParser

def gen_secret(clen=32):
    rg = random.SystemRandom()
    r = list('1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ')
    return ''.join([rg.choice(r) for _ in range(clen)])

def update(config):
    conf_file = config.get('DEFAULT','cfgfile')
    shutil.copy(conf_file,"%s.%s"%(conf_file,int(time.time())))
    old_secret = config.get('DEFAULT','secret')
    config.set('DEFAULT','secret',gen_secret(32))

    old_AESCipher = utils.AESCipher(old_secret)
    new_AESCipher = utils.AESCipher(config.get('DEFAULT','secret'))

    engine = get_engine(config)
    conn = engine.connect()
    
    # update 
    db = scoped_session(sessionmaker(bind=engine, autocommit=False, autoflush=True))()  
    user_query = db.query(models.SlcRadAccount)
    total1 = user_query.count()
    for user in user_query:
        oldpwd = old_AESCipher.decrypt(user.password)
        user.password = new_AESCipher.encrypt(oldpwd)
        
    vcard_query = db.query(models.SlcRechargerCard)
    total2 = vcard_query.count()
    for vcard in vcard_query:
        oldpwd = old_AESCipher.decrypt(vcard.card_passwd)
        vcard.card_passwd = new_AESCipher.encrypt(oldpwd)
      
    db.commit()
    with open(conf_file,'wb') as configfile:
        config.write(configfile)
    
    print 'update secret success user %s,vcard %s'%(total1,total2)
    