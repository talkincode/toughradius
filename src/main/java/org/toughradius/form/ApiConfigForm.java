package org.toughradius.form;

public class ApiConfigForm {

    private String apiType;
    private String apiUsername;
    private String apiPasswd;
    private String apiAllowIplist;
    private String apiBlackIplist;

    public String getApiType() {
        return apiType;
    }

    public void setApiType(String apiType) {
        this.apiType = apiType;
    }

    public String getApiUsername() {
        return apiUsername;
    }

    public void setApiUsername(String apiUsername) {
        this.apiUsername = apiUsername;
    }

    public String getApiPasswd() {
        return apiPasswd;
    }

    public void setApiPasswd(String apiPasswd) {
        this.apiPasswd = apiPasswd;
    }

    public String getApiAllowIplist() {
        return apiAllowIplist;
    }

    public void setApiAllowIplist(String apiAllowIplist) {
        this.apiAllowIplist = apiAllowIplist;
    }

    public String getApiBlackIplist() {
        return apiBlackIplist;
    }

    public void setApiBlackIplist(String apiBlackIplist) {
        this.apiBlackIplist = apiBlackIplist;
    }
}
