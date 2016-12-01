#!/usr/bin/env python
# coding=utf-8


import types
from urllib import urlencode
from hashlib import md5
from twisted.internet import defer
from cyclone import httpclient

class AliPay:

    GATEWAY = 'https://mapi.alipay.com/gateway.do?'

    def __init__(self, settings={}):
        self.settings = settings

    def event_alipay_setup(self,settings):
        self.settings = settings

    def safestr(self, val, errors='strict'):
        encoding = self.settings.get('ALIPAY_INPUT_CHARSET','utf-8')
        if val is None:
            return ''
        if isinstance(val, unicode):
            return val.encode('utf-8',errors)
        elif isinstance(val, str):
            return val.decode('utf-8', errors).encode(encoding, errors)
        elif isinstance(val, (int,float)):
            return str(val)
        elif isinstance(val, Exception):
            return ' '.join([self.safestr(arg, encoding,errors) for arg in val])
        else:
            try:
                return str(val)
            except:
                return unicode(val).encode(encoding, errors)
        return val

    def make_sign(self, **msg):
        ks = msg.keys()
        ks.sort()
        sign_str = '&'.join([ '%s=%s'%(k,msg[k]) for k in ks ])
        if 'MD5' == self.settings.ALIPAY_SIGN_TYPE:
            return md5(sign_str + self.settings.ALIPAY_KEY).hexdigest()

        raise Exception('not support sign type %s' % settings.ALIPAY_SIGN_TYPE)


    def check_sign(self, **msg):
        if "sign" not in msg:
            return False
        params = {self.safestr(k):self.safestr(msg[k]) for k in msg if k in ('sign','sign_type')}
        local_sign = make_sign(params)
        return msg['sign'] == local_sign


    def make_request_url(self,**params):
        params.pop('sign',None)
        params.pop('sign_type',None)
        _params = {self.safestr(k):self.safestr(v) for k,v in params.iteritems() if v not in ('', None) }
        _params['sign'] = self.make_sign(**_params)
        _params['sign_type'] = self.settings.ALIPAY_SIGN_TYPE
        return AliPay.GATEWAY + urlencode(_params)


    def create_direct_pay_by_user(self, tn, subject, body, total_fee):
        params = {}
        params['service']       = 'create_direct_pay_by_user'
        params['payment_type']  = '1'
        
        # 获取配置文件
        params['partner']           = self.settings.ALIPAY_PARTNER
        params['seller_email']      = self.settings.ALIPAY_SELLER_EMAIL
        params['return_url']        = self.settings.ALIPAY_RETURN_URL
        params['notify_url']        = self.settings.ALIPAY_NOTIFY_URL
        params['_input_charset']    = self.settings.ALIPAY_INPUT_CHARSET
        params['show_url']          = self.settings.ALIPAY_SHOW_URL
        
        # 从订单数据中动态获取到的必填参数
        params['out_trade_no']  = tn        # 请与贵网站订单系统中的唯一订单号匹配
        params['subject']       = subject   # 订单名称，显示在支付宝收银台里的“商品名称”里，显示在支付宝的交易管理的“商品名称”的列表里。
        params['body']          = body      # 订单描述、订单详细、订单备注，显示在支付宝收银台里的“商品描述”里
        params['total_fee']     = total_fee # 订单总金额，显示在支付宝收银台里的“应付总额”里
        
        # 扩展功能参数——网银提前
        params['paymethod'] = 'directPay'   # 默认支付方式，四个值可选：bankPay(网银); cartoon(卡通); directPay(余额); CASH(网点支付)
        params['defaultbank'] = ''          # 默认网银代号，代号列表见http://club.alipay.com/read.php?tid=8681379
        
        # 扩展功能参数——防钓鱼
        params['anti_phishing_key'] = ''
        params['exter_invoke_ip'] = ''
        
        # 扩展功能参数——自定义参数
        params['buyer_email'] = ''
        params['extra_common_param'] = ''
        
        # 扩展功能参数——分润
        params['royalty_type'] = ''
        params['royalty_parameters'] = ''
        
        return self.make_request_url(**params)
    
    @defer.inlineCallbacks
    def notify_verify(self, request):
        params = {}

        params['is_success'] = request.get_argument('is_success', '')
        params['partnerId'] = request.get_argument('partnerId', '')

        params['notify_id'] = request.get_argument('notify_id', '')
        params['notify_type'] = request.get_argument('notify_type', '')
        params['notify_time'] = request.get_argument('notify_time', '')
        params['sign'] = request.get_argument('sign', '')
        params['sign_type'] = request.get_argument('sign_type', '')

        params['trade_no'] = request.get_argument('trade_no', '')
        params['subject'] = request.get_argument('subject', '')
        params['price'] = request.get_argument('price', '')
        params['quantity'] = request.get_argument('quantity', '')
        params['seller_email'] = request.get_argument('seller_email', '')
        params['seller_id'] = request.get_argument('seller_id', '')
        params['buyer_email'] = request.get_argument('buyer_email', '')
        params['buyer_id'] = request.get_argument('buyer_id', '')
        params['discount'] = request.get_argument('discount', '')
        params['total_fee'] = request.get_argument('total_fee', '')
        params['trade_status'] = request.get_argument('trade_status', '')
        params['is_total_fee_adjust'] = request.get_argument('is_total_fee_adjust', '')
        params['use_coupon'] = request.get_argument('use_coupon', '')
        params['body'] = request.get_argument('body', '')
        params['exterface'] = request.get_argument('exterface', '')
        params['out_trade_no'] = request.get_argument('out_trade_no', '')
        params['payment_type'] = request.get_argument('payment_type', '')
        params['logistics_type'] = request.get_argument('logistics_type', '')
        params['logistics_fee'] = request.get_argument('logistics_fee', '')
        params['logistics_payment'] = request.get_argument('logistics_payment', '')
        params['gmt_logistics_modify'] = request.get_argument('gmt_logistics_modify', '')
        params['buyer_actions'] = request.get_argument('buyer_actions', '')
        params['seller_actions'] = request.get_argument('seller_actions', '')
        params['gmt_create'] = request.get_argument('gmt_create', '')
        params['gmt_payment'] = request.get_argument('gmt_payment', '')
        params['refund_status'] = request.get_argument('refund_status', '')
        params['gmt_refund'] = request.get_argument('gmt_refund', '')
        params['receive_name'] = request.get_argument('receive_name', '')
        params['receive_address'] = request.get_argument('receive_address', '')
        params['receive_zip'] = request.get_argument('receive_zip', '')
        params['receive_phone'] = request.get_argument('receive_phone', '')
        params['receive_mobile'] = request.get_argument('receive_mobile', '')

        if not self.check_sign(request.get_argument('sign', '')):
            defer.returnValue(False)
        
        # 二级验证--查询支付宝服务器此条信息是否有效
        params = {}
        params['partner'] = self.settings.ALIPAY_PARTNER
        params['notify_id'] = request.get_argument('notify_id', '')
        if self.settings.ALIPAY_TRANSPORT == 'https':
            params['service'] = 'notify_verify'
            gateway = 'https://mapi.alipay.com/gateway.do'
        else:
            gateway = 'http://notify.alipay.com/trade/notify_query.do'

        resp = yield httpclient.fetch(gateway,postdata=urlencode(params))
        veryfy_result = resp.body

        defer.returnValue(veryfy_result.lower().strip() == 'true')


if __name__ == '__main__':
    from toughradius.common.storage import Storage
    settings = Storage(
        ALIPAY_KEY = '234234',
        ALIPAY_INPUT_CHARSET = 'utf-8',
        # 合作身份者ID，以2088开头的16位纯数字
        ALIPAY_PARTNER = '234',
        # 签约支付宝账号或卖家支付宝帐户
        ALIPAY_SELLER_EMAIL = 'payment@34.com',
        ALIPAY_SIGN_TYPE = 'MD5',
        # 付完款后跳转的页面（同步通知） 要用 http://格式的完整路径，不允许加?id=123这类自定义参数
        ALIPAY_RETURN_URL='',
        # 交易过程中服务器异步通知的页面 要用 http://格式的完整路径，不允许加?id=123这类自定义参数
        ALIPAY_NOTIFY_URL='',
        ALIPAY_SHOW_URL='',
        # 访问模式,根据自己的服务器是否支持ssl访问，若支持请选择https；若不支持请选择http
        ALIPAY_TRANSPORT='https'
    )
    alipay = AliPay(settings)
    params = {}
    params['service']       = 'create_direct_pay_by_user'
    params['payment_type']  = '1'
    params['aaaa'] = u"好"
    print alipay.make_request_url(**params)
    print alipay.create_direct_pay_by_user('2323525', u"阿士大夫", u"啥打法是否", 0.01)


