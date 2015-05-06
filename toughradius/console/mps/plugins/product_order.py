#coding:utf-8

from cyclone.util import ObjectDict
from toughradius.console import models

__name__ = 'product_order'

def test(data, msg=None, bot=None, db=None,**kwargs):
    return data.strip() == '3' or data.strip()  in u"在线订购"

def respond(data, msg=None,db=None,config=None,mpsapi=None,**kwargs):
    products = db.query(models.SlcRadProduct).filter(
        models.SlcRadProduct.mps_flag == 1,
        models.SlcRadProduct.product_status == 0
    ).limit(7)

    articles =[]
    for item in products:
        article=ObjectDict()
        article.title= item.product_name
        article.description = ''
        article.url = "%s/order?openid=%s&product_id=%s" % (
            config.get('mps','server_base'),
            msg.fromuser,
            item.id
        )
        article.picurl = '%s/static/img/mps/order_online.jpg' % config.get('mps','server_base')
        articles.append(article)
    return articles