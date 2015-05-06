#coding=utf-8
import jieba.analyse


__name__ = 'filter'

WDS = ((u'傻逼', u'傻比', u'垃圾', u'流氓', u'无耻', u'下流', u'白痴'))


def test(data, msg=None, bot=None, db=None,**kwargs):
    ex = set(jieba.analyse.extract_tags(data, 20))
    return len(ex.intersection(WDS)) > 0


def respond(data, msg=None, bot=None, db=None,**kwargs):
    return u'请文明用语哦'
