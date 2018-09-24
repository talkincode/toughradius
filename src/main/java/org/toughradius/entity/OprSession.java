package org.toughradius.entity;

import java.util.*;

public class OprSession {

    private String username;
    private Date lastLogin;
    private String loginIp;
    private String sysVersion;
    private String mobile;
    private String email;
    private Map<String,String> sysConfig = new HashMap<String,String>();
    private List<MenuItem> menus = new ArrayList<MenuItem>();

    public OprSession() {
    }

    public OprSession(String username, String loginIp, String version) {
        setUsername(username);
        setLoginIp(loginIp);
        setLastLogin(new Date());
        setSysVersion(version);
    }

    public void setSysConfig(String name,String value){
        sysConfig.put(name,value);
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public Date getLastLogin() {
        return lastLogin;
    }

    public void setLastLogin(Date lastLogin) {
        this.lastLogin = lastLogin;
    }

    public String getLoginIp() {
        return loginIp;
    }

    public void setLoginIp(String loginIp) {
        this.loginIp = loginIp;
    }

    public String getMobile() {
        return mobile;
    }

    public void setMobile(String mobile) {
        this.mobile = mobile;
    }

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public List<MenuItem> getMenus() {
        return menus;
    }

    public void setMenus(List<MenuItem> menus) {
        this.menus = menus;
    }

    public String getSysVersion() {
        return sysVersion;
    }

    public void setSysVersion(String sysVersion) {
        this.sysVersion = sysVersion;
    }
}
