package org.toughradius.form;
import org.toughradius.common.ValidateUtil;

import java.util.Date;

public class WlanParam {

    private String template;
    private String wlanuserip;
    private String wlanusername;
    private String wlanusermac;
    private String wlanstamac;
    private String wlanacname;
    private String wlanacip;
    private String wlanapmac;
    private String ssid;
    private String wlanuserfirsturl;
    private String vendor;
    private String error;
    private String username;
    private String domain;
    private String authmode;
    private String url;
    private String srcacip;
    private String phone;
    private String smscode;
    private String rememberPwd;
    private long starttime;
    private boolean rmflag;

    public WlanParam() {
        starttime = new Date().getTime();
    }

    public String getWlanuserip() {
        return wlanuserip;
    }

    public void setWlanuserip(String wlanuserip) {
        this.wlanuserip = wlanuserip;
    }

    public String getWlanusername() {
        return wlanusername;
    }

    public void setWlanusername(String wlanusername) {
        this.wlanusername = wlanusername;
    }

    public String getWlanusermac() {
        return wlanusermac;
    }

    public void setWlanusermac(String wlanusermac) {
        this.wlanusermac = wlanusermac;
    }

    public String getWlanacname() {
        return wlanacname;
    }

    public void setWlanacname(String wlanacname) {
        this.wlanacname = wlanacname;
    }

    public String getWlanacip() {
        return wlanacip;
    }

    public void setWlanacip(String wlanacip) {
        this.wlanacip = wlanacip;
    }

    public String getWlanapmac() {
        return wlanapmac;
    }

    public void setWlanapmac(String wlanapmac) {
        this.wlanapmac = wlanapmac;
    }

    public String getSsid() {
        return ssid;
    }

    public void setSsid(String ssid) {
        this.ssid = ssid;
    }

    public String getWlanuserfirsturl() {
        return wlanuserfirsturl;
    }

    public void setWlanuserfirsturl(String wlanuserfirsturl) {
        this.wlanuserfirsturl = wlanuserfirsturl;
    }

    public String getVendor() {
        return vendor;
    }

    public void setVendor(String vendor) {
        this.vendor = vendor;
    }

    public String getError() {
        return error;
    }

    public void setError(String error) {
        this.error = error;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public String getAuthmode() {
        return authmode;
    }

    public void setAuthmode(String authmode) {
        this.authmode = authmode;
    }

    public String getWlanstamac() {
        return wlanstamac;
    }

    public void setWlanstamac(String wlanstamac) {
        this.wlanstamac = wlanstamac;
    }

    public String getClientMac(){
        if(ValidateUtil.isMacAddress(getWlanstamac()))
            return getWlanstamac();
        else if(ValidateUtil.isMacAddress(getWlanusermac()))
            return getWlanusermac();
        return null;
    }

    public String getDomain() {
        return domain;
    }

    public void setDomain(String domain) {
        this.domain = domain;
    }

    public String getUrl() {
        return url;
    }

    public void setUrl(String url) {
        this.url = url;
    }

    public String getSrcacip() {
        return srcacip;
    }

    public void setSrcacip(String srcacip) {
        this.srcacip = srcacip;
    }

    public String getTemplate() {
        return template;
    }

    public void setTemplate(String template) {
        this.template = template;
    }

    private String emptyStr(String src){
        if(src==null){
            return "";
        }
        return src;
    }
    public String getPhone() {
        return phone;
    }

    public void setPhone(String phone) {
        this.phone = phone;
    }

    public String getSmscode() {
        return smscode;
    }

    public void setSmscode(String smscode) {
        this.smscode = smscode;
    }

    public long getStarttime() {
        return starttime;
    }

    public void setStarttime(long starttime) {
        this.starttime = starttime;
    }

    public boolean isRmflag() {
        return rmflag;
    }

    public void setRmflag(boolean rmflag) {
        this.rmflag = rmflag;
    }

    public String getRememberPwd() {
        return rememberPwd;
    }

    public void setRememberPwd(String rememberPwd) {
        this.rememberPwd = rememberPwd;
    }

    public String encodeParams() {
        StringBuilder buff = new StringBuilder();
        buff.append(String.format("wlanuserip=%s&", emptyStr(wlanuserip)));
        buff.append(String.format("wlanusername=%s&", emptyStr(wlanusername)));
        buff.append(String.format("wlanusermac=%s&", emptyStr(wlanusermac)));
        buff.append(String.format("wlanacname=%s&", emptyStr(wlanacname)));
        buff.append(String.format("wlanacip=%s&", emptyStr(wlanacip)));
        buff.append(String.format("wlanapmac=%s&", emptyStr(wlanapmac)));
        buff.append(String.format("ssid=%s&", emptyStr(ssid)));
        buff.append(String.format("wlanuserfirsturl=%s&", emptyStr(wlanuserfirsturl)));
        return buff.toString();
    }

    @Override
    public String toString() {
        return "WlanParam{" +
                "template='" + template + '\'' +
                ", wlanuserip='" + wlanuserip + '\'' +
                ", wlanusername='" + wlanusername + '\'' +
                ", wlanusermac='" + wlanusermac + '\'' +
                ", wlanstamac='" + wlanstamac + '\'' +
                ", wlanacname='" + wlanacname + '\'' +
                ", wlanacip='" + wlanacip + '\'' +
                ", wlanapmac='" + wlanapmac + '\'' +
                ", ssid='" + ssid + '\'' +
                ", wlanuserfirsturl='" + wlanuserfirsturl + '\'' +
                ", vendor='" + vendor + '\'' +
                ", error='" + error + '\'' +
                ", username='" + username + '\'' +
                ", domain='" + domain + '\'' +
                ", authmode='" + authmode + '\'' +
                ", url='" + url + '\'' +
                ", srcacip='" + srcacip + '\'' +
                ", phone='" + phone + '\'' +
                ", smscode='" + smscode + '\'' +
                ", rememberPwd='" + rememberPwd + '\'' +
                ", starttime=" + starttime +
                ", rmflag=" + rmflag +
                '}';
    }


}
