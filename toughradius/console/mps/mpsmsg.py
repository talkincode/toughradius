#coding=utf-8
import time
import logging
import uuid
from cyclone.util import ObjectDict
from xml.etree import ElementTree
from twisted.python import log

MSG_TYPE_TEXT = u'text'
MSG_TYPE_LOCATION = u'location'
MSG_TYPE_IMAGE = u'image'
MSG_TYPE_LINK = u'link'
MSG_TYPE_EVENT = u'event'
MSG_TYPE_MUSIC = u'music'
MSG_TYPE_NEWS = u'news'
MSG_TYPE_VOICE = u'voice'
MSG_TYPE_VIDEO = u'video'
MSG_TYPE_CUSTOMER = u'transfer_customer_service'


def get_uuid():
    return uuid.uuid1().hex.upper()

def decode(s):
    if isinstance(s, str):
        s = s.decode('utf-8')
    return s


def parse_node(parser, node, defval=''):
    _nd = parser.find(node)
    return decode(_nd.text) if _nd is not None else defval


def parse_msg(xml):
    if not xml:
        return None
    parser = ElementTree.fromstring(xml)
    msg_id = parse_node(parser, 'MsgId', get_uuid())
    msg_type = parse_node(parser, 'MsgType')
    touser = parse_node(parser, 'ToUserName')
    fromuser = parse_node(parser, 'FromUserName')
    create_at = int(parse_node(parser, 'CreateTime', 0))
    msg = ObjectDict(
        mid=msg_id,
        type=msg_type,
        touser=touser,
        fromuser=fromuser,
        time=create_at
    )
    if msg_type == MSG_TYPE_TEXT:
        msg.content = parse_node(parser, 'Content')
    elif msg_type == MSG_TYPE_LOCATION:
        msg.location_x = parse_node(parser, 'Location_X')
        msg.location_y = parse_node(parser, 'Location_Y')
        msg.scale = int(parse_node(parser, 'Scale'))
        msg.label = parse_node(parser, 'Label')
    elif msg_type == MSG_TYPE_IMAGE:
        msg.picurl = parse_node(parser, 'PicUrl')
    elif msg_type == MSG_TYPE_VOICE:
        msg.media_id = parse_node(parser, 'MediaId')
        msg.format = parse_node(parser, 'Format')
    elif msg_type == MSG_TYPE_VIDEO:
        msg.media_id = parse_node(parser, 'MediaId')
        msg.thumb = parse_node(parser, 'ThumbMediaId')
    elif msg_type == MSG_TYPE_LINK:
        msg.title = parse_node(parser, 'Title')
        msg.description = decode(parser.find('Description').text)
        msg.url = parse_node(parser, 'Url')
    elif msg_type == MSG_TYPE_EVENT:
        msg.event = parse_node(parser, 'Event')
        msg.event_key = parse_node(parser, 'EventKey')
        msg.ticket = parse_node(parser, 'Ticket')
        if msg.event == u'LOCATION':
            msg.latitude = parse_node(parser, 'Latitude')
            msg.longitude = parse_node(parser, 'Longitude')
            msg.precision = parse_node(parser, 'Precision')
    return msg


def reply_text(fromuser, touser, text):
    tpl = """<xml>
    <ToUserName><![CDATA[%s]]></ToUserName>
    <FromUserName><![CDATA[%s]]></FromUserName>
    <CreateTime>%s</CreateTime>
    <MsgType><![CDATA[%s]]></MsgType>
    <Content><![CDATA[%s]]></Content>
    <FuncFlag>0</FuncFlag>
    </xml>
    """

    timestamp = int(time.time())
    return tpl % (touser, fromuser, timestamp, MSG_TYPE_TEXT, text)


def reply_customer(fromuser, touser,kfaccount):
    tpl = """<xml>
    <ToUserName><![CDATA[%s]]></ToUserName>
    <FromUserName><![CDATA[%s]]></FromUserName>
    <CreateTime>%s</CreateTime>
    <MsgType><![CDATA[%s]]></MsgType>
    <TransInfo>
         <KfAccount><![CDATA[%s]]></KfAccount>
     </TransInfo>
    </xml>
    """

    timestamp = int(time.time())
    return tpl % (touser, fromuser, timestamp, MSG_TYPE_CUSTOMER,kfaccount)


def reply_music(fromuser, touser, music):
    tpl = """<xml>
    <ToUserName><![CDATA[%s]]></ToUserName>
    <FromUserName><![CDATA[%s]]></FromUserName>
    <CreateTime>%s</CreateTime>
    <MsgType><![CDATA[%s]]></MsgType>
    <Music>
    <Title><![CDATA[%s]]></Title>
    <Description><![CDATA[%s]]></Description>
    <MusicUrl><![CDATA[%s]]></MusicUrl>
    <HQMusicUrl><![CDATA[%s]]></HQMusicUrl>
    </Music>
    <FuncFlag>0</FuncFlag>
    </xml>
    """

    timestamp = int(time.time())
    return tpl % (touser, fromuser, timestamp, MSG_TYPE_MUSIC,
                  music['titlle'], music['description'],
                  music['music_url'], music['hq_music_url'])


def reply_articles(fromuser, touser, articles):
    tpl = """<xml>
    <ToUserName><![CDATA[%s]]></ToUserName>
    <FromUserName><![CDATA[%s]]></FromUserName>
    <CreateTime>%s</CreateTime>
    <MsgType><![CDATA[%s]]></MsgType>
    <ArticleCount>%s</ArticleCount>
    <Articles>%s</Articles>
    <FuncFlag>0</FuncFlag>
    </xml>
    """
    itemtpl = """<item>
    <Title><![CDATA[%s]]></Title>
    <Description><![CDATA[%s]]></Description>
    <PicUrl><![CDATA[%s]]></PicUrl>
    <Url><![CDATA[%s]]></Url>
    </item>
    """

    timestamp = int(time.time())
    items = []
    if not isinstance(articles, list):
        articles = [articles]
    count = len(articles)
    for article in articles:
        item = itemtpl % (article['title'], article['description'],
                          article['picurl'], article['url'])
        items.append(item)
    article_str = '\n'.join(items)

    return tpl % (touser, fromuser, timestamp, MSG_TYPE_NEWS,
                  count, article_str)


def gen_reply(fromuser, touser, result):
    if result['msg_type'] == MSG_TYPE_NEWS:
        return reply_articles(fromuser, touser, result['response'])
    elif result['msg_type'] == MSG_TYPE_TEXT:
        return reply_text(fromuser, touser, result['response'])
    elif result['msg_type'] == MSG_TYPE_MUSIC:
        return reply_music(fromuser, touser, result['response'])
    elif result['msg_type'] == MSG_TYPE_CUSTOMER:
        return  reply_customer(fromuser,touser,result['kfaccount'])

