package org.toughradius.entity;

import org.toughradius.form.WlanParam;

import java.util.Date;

public class WlanSession {

    private String username;
    private String sessionId;
    private WlanParam wlanParam;
    private int loginStatus;
    private Date lastLogin;

    public WlanSession() {
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }


    public String getSessionId() {
        return sessionId;
    }

    public void setSessionId(String sessionId) {
        this.sessionId = sessionId;
    }

    public Date getLastLogin() {
        return lastLogin;
    }

    public void setLastLogin(Date lastLogin) {
        this.lastLogin = lastLogin;
    }

    public WlanParam getWlanParam() {
        return wlanParam;
    }

    public void setWlanParam(WlanParam wlanParam) {
        this.wlanParam = wlanParam;
    }

    public int getLoginStatus() {
        return loginStatus;
    }

    public void setLoginStatus(int loginStatus) {
        this.loginStatus = loginStatus;
    }
}
