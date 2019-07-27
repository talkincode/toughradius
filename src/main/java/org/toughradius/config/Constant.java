package org.toughradius.config;

import org.toughradius.common.CoderUtil;

public interface Constant {

    public static final String SESSION_USER_KEY = "SESSION_USER_KEY";
    public static final String SESSION_VCODE_KEY = "SESSION_VCODE_KEY";

    public final static String RADIUS_MODULE = "radius";
    public final static String RADIUS_IGNORE_PASSWORD = "radiusIgnorePassword";
    public final static String RADIUS_INTERIM_INTELVAL = "radiusInterimIntelval";
    public final static String RADIUS_TICKET_HISTORY_DAYS = "radiusTicketHistoryDays";
    public final static String RADIUS_EXPORE_ADDR_POOL = "radiusExpireAddrPool";

    public final static String SYSTEM_MODULE = "system";
    public final static String SYSTEM_USERNAME = "systemUsername";
    public final static String SYSTEM_USERPWD = "systemUserpwd";

    public final static String API_MODULE = "api";
    public final static String API_TYPE = "apiType";
    public final static String API_USERNAME = "apiUsername";
    public final static String API_PASSWD = "apiPasswd";
    public final static String API_ALLOW_IPLIST = "apiAllowIplist";
    public final static String API_BLACK_IPLIST = "apiBlackIplist";

    public final static String SMS_MODULE = "sms";
    public final static String SMS_GATEWAY = "smsGateway";
    public final static String SMS_APPID = "smsAppid";
    public final static String SMS_APPKEY = "smsAppkey";
    public final static String SMS_VCODE_TEMPLATE = "smsVcodeTemplate";

    public final static String WLAN_MODULE = "wlan";
    public final static String WLAN_TEMPLATE = "wlanTemplate";
    public final static String WLAN_RESULT_URL = "wlanResultUrl";
    public final static String WLAN_JOIN_URL = "wlanJoinUrl";
    public final static String WLAN_USERAUTH_ENABLED = "wlanUserauthEnabled";
    public final static String WLAN_PWDAUTH_ENABLED = "wlanPwdauthEnabled";
    public final static String WLAN_SMSAUTH_ENABLED = "wlanSmsauthEnabled";
    public final static String WLAN_WXAUTH_ENABLED = "wlanWxauthEnabled";
    public final static String WLAN_WECHAT_SSID = "wlanWechatSsid";
    public final static String WLAN_WECHAT_SHOPID = "wlanWechatShopid";
    public final static String WLAN_WECHAT_APPID = "wlanWechatAppid";
    public final static String WLAN_WECHAT_SECRETKEY = "wlanWechatSecretkey";


    public final static String PORTAL_AUTH_USERPWD = "userauth";
    public final static String PORTAL_AUTH_PASSWORD = "pwdauth";
    public final static String PORTAL_AUTH_SMS = "smsauth";
    public final static String PORTAL_AUTH_WEIXIN = "wxauth";
    public final static String PORTAL_REMBERPWD_COOKIE = "PORTAL_REMBERPWD_COOKIE";
}
