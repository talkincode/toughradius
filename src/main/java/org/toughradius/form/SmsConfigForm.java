package org.toughradius.form;

public class SmsConfigForm {

    private String smsGateway;
    private String smsAppid;
    private String smsAppkey;
    private String smsVcodeTemplate;

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

    public String getSmsVcodeTemplate() {
        return smsVcodeTemplate;
    }

    public void setSmsVcodeTemplate(String smsVcodeTemplate) {
        this.smsVcodeTemplate = smsVcodeTemplate;
    }
}
