package org.toughradius.form;

public class SmsConfigForm {

    private String smsGateway;
    private String smsAppid;
    private String smsAppkey;

    public String getSmsGateway() {
        return smsGateway;
    }

    public void setSmsGateway(String smsGateway) {
        this.smsGateway = smsGateway;
    }

    public String getSmsAppid() {
        return smsAppid;
    }

    public void setSmsAppid(String smsAppid) {
        this.smsAppid = smsAppid;
    }

    public String getSmsAppkey() {
        return smsAppkey;
    }

    public void setSmsAppkey(String smsAppkey) {
        this.smsAppkey = smsAppkey;
    }
}
