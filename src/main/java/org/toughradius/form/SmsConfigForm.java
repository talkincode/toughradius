package org.toughradius.form;

public class SmsConfigForm {

    private String SMS_GATEWAY;
    private String SMS_APPID;
    private String SMS_APPKEY;

    public String getSMS_GATEWAY() {
        return SMS_GATEWAY;
    }

    public void setSMS_GATEWAY(String SMS_GATEWAY) {
        this.SMS_GATEWAY = SMS_GATEWAY;
    }

    public String getSMS_APPID() {
        return SMS_APPID;
    }

    public void setSMS_APPID(String SMS_APPID) {
        this.SMS_APPID = SMS_APPID;
    }

    public String getSMS_APPKEY() {
        return SMS_APPKEY;
    }

    public void setSMS_APPKEY(String SMS_APPKEY) {
        this.SMS_APPKEY = SMS_APPKEY;
    }
}
