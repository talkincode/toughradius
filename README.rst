ToughRADIUS  Windows Ver.
====================================

python toughctl -radiusd -c ../radiusd.conf << test Radius

python toughctl -admin -c ../radiusd.conf   << Test Admin pycharm


Web Access
================================


Web Access Admin：http://serverip:1816
 
User / pass：admin/root



Web Access Users :http://serverip:1817

###############################################################################
# Select Language
###############################################################################
::

    # 启动

    $ toughctl --start standalone

    # 停止

    $ toughctl --stop standalone

    # 设置开机启动

    $ echo "toughctl --start standalone" >> /etc/rc.local::

    # 启动

    $ toughctl --start standalone

    # 停止

    $ toughctl --stop standalone

    # 设置开机启动

    $ echo "toughctl --start standalone" >> /etc/rc.local
@app.get('/th')
def lang_th_get():
    tr.language = 'th'
    redirect('/')

@app.post('/th')
def lang_th_post():
    redirect('/')

@app.get('/en')
def lang_en():
    tr.language = 'en'
    redirect('/')
@app.get('/cn')
def lang_cn():
    tr.language = ''
    redirect('/')
