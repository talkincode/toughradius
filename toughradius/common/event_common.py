#!/usr/bin/env python
# coding:utf-8


from toughlib import dispatch

"""触发邮件,短信发送公共方法"""


def trigger_notify(obj, user_info, **kwargs):
    if int(obj.get_param_value("webhook_notify_enable", 0)) > 0 and kwargs.get('webhook_notify'):
        dispatch.pub(kwargs['webhook_notify'], user_info, async=False)

    if int(obj.get_param_value("mail_notify_enable", 0)) > 0:
        if obj.get_param_value("mail_mode", 'smtp') == 'toughcloud' and \
                obj.get_param_value("toughcloud_license", None) and kwargs.get('toughcloud_mail'):
            dispatch.pub(kwargs['toughcloud_mail'], user_info, async=False)
        if obj.get_param_value("mail_mode", 'smtp') == 'smtp' and kwargs.get('smtp_mail'):
            dispatch.pub(kwargs['smtp_mail'], user_info, async=False)

    if int(obj.get_param_value("sms_notify_enable", 0)) > 0 and \
            obj.get_param_value("toughcloud_license", None) and kwargs.get('toughcloud_sms'):
        dispatch.pub(kwargs['toughcloud_sms'], user_info, async=False)


