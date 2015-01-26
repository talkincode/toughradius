#!/usr/bin/env python
#coding:utf-8
import sys,os
sys.path.insert(0,os.path.split(__file__)[0])
sys.path.insert(0,os.path.abspath(os.path.pardir))
from twisted.python import log
from bottle import Bottle
from bottle import request
from bottle import response
from bottle import redirect
from bottle import run as runserver
from bottle import static_file
from bottle import abort
from bottle import mako_template as render
from hashlib import md5
from tablib import Dataset
from libs import sqla_plugin 
from base import *
from libs import utils
import bottle
import models
import forms
import decimal
import datetime

###############################################################################
# init      
###############################################################################   

app = Bottle()

def init_app():
    ''' install plugins'''
    sqla_pg = sqla_plugin.Plugin(
        models.engine, 
        models.metadata, 
        keyword='db', 
        create=False, 
        commit=False, 
        use_kwargs=False 
    )
    init_context(session=sqla_pg.new_session())
    app.install(sqla_pg)

def auth_mbr(func):
    @functools.wraps(func)
    def warp(*args,**kargs):
        if not get_cookie("member"):
            log.msg("member login timeout")
            return redirect('/member/login')
        else:
            return func(*args,**kargs)
    return warp
    
###############################################################################
# Basic handle         
###############################################################################    
    
@app.error(404)
def error404(error):
    return render("error.html",msg=u"页面不存在 - 请联系管理员!")

@app.error(500)
def error500(error):
    return render("error.html",msg=u"出错了： %s"%error.exception)

@app.route('/static/:path#.+#')
def route_static(path):
    return static_file(path, root='./static')    

###############################################################################
# login handle         
###############################################################################

@app.get('/',apply=auth_mbr)
def member_index(db):
    return render("member/index")

@app.get('/login')
def member_login_get(db):
    form = forms.member_login_form()
    form.next.set_value(request.params.get('next','/member'))
    return render("member/login",form=form)

@app.post('/login')
def member_login_post(db):
    next = request.params.get("next", "/member")
    form = forms.member_login_form()
    if not form.validates(source=request.params):
        return render("member/login", form=form)

    member = db.query(models.SlcMember).filter_by(
        member_name=form.d.username,
        password=md5(form.d.password.encode()).hexdigest()
    ).first()

    if not member:
        return render("member/login", form=form,msg=u"用户名密码不符合")
 
    set_cookie('member',form.d.username)
    set_cookie('member_login_time', utils.get_currtime())
    set_cookie('member_login_ip', request.remote_addr) 
    redirect(next)

@app.get("/logout")
def member_logout():
    set_cookie('member',None)
    set_cookie('member_login_time', None)
    set_cookie('member_login_ip', None)     
    request.cookies.clear()
    redirect('/member/login')


@app.get('/join')
def member_join_get(db):
    form = forms.member_join_form()
    return render("member/join",form=form)
    
@app.post('/join')
def member_join_post(db):
    form = forms.member_join_form()
    if not form.validates(source=request.params):
        return render("member/join", form=form)    
        
    if self.db.query(exists().where(models.SlcMember.member_name == form.d.username)).scalar():
        return render("member/join",form=form,msg=u"用户{0}已被使用".format(form.d.username))
        
    if self.db.query(exists().where(models.SlcMember.email == form.d.email)).scalar():
        return render("member/join",form=form,msg=u"用户邮箱{0}已被使用".format(form.d.email))
   
   

###############################################################################
# run server                                                                 
###############################################################################

def main():
    import argparse,json
    parser = argparse.ArgumentParser()
    parser.add_argument('-http','--httpport', type=int,default=0,dest='httpport',help='http port')
    parser.add_argument('-d','--debug', nargs='?',type=bool,default=False,dest='debug',help='debug')
    parser.add_argument('-c','--conf', type=str,default="../config.json",dest='conf',help='conf file')
    args =  parser.parse_args(sys.argv[1:])

    if not args.conf or not os.path.exists(args.conf):
        print 'no config file user -c or --conf cfgfile'
        return

    _config = json.loads(open(args.conf).read())
    _database = _config['database']
    _customer = _config['customer']

    if args.httpport:
        _customer['httpport'] = args.httpport
    if args.debug:
        _customer['debug'] = bool(args.debug)

    from sqlalchemy import create_engine
    models.engine = create_engine(models.get_db_connstr(_database))

    init_app()
    
    log.startLogging(sys.stdout)    
    runserver(
        app, host='0.0.0.0', 
        port=_customer['httpport'] ,
        debug=bool(_customer['debug']),
        reloader=bool(_customer['debug']),
        server="twisted"
    )

if __name__ == "__main__":
    main()   