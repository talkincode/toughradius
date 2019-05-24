package org.toughradius.form;

public class RouterOSParam {

    private String ssid;
    private String mac;
    private String ip;
    private String username;
    private String linkLogin;
    private String linkLogout;
    private String linkStatus;
    private String linkOrig;
    private Long uptimeSecs;
    private Long bytesIn;
    private Long bytesOut;
    private String error;

    public String getSsid() {
        return ssid;
    }

    public void setSsid(String ssid) {
        this.ssid = ssid;
    }

    public String getMac() {
        return mac;
    }

    public void setMac(String mac) {
        this.mac = mac;
    }

    public String getIp() {
        return ip;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public String getLinkLogin() {
        return linkLogin;
    }

    public void setLinkLogin(String linkLogin) {
        this.linkLogin = linkLogin;
    }

    public String getLinkOrig() {
        return linkOrig;
    }

    public void setLinkOrig(String linkOrig) {
        this.linkOrig = linkOrig;
    }

    public String getError() {
        return error;
    }

    public void setError(String error) {
        this.error = error;
    }

    public String getLinkLogout() {
        return linkLogout;
    }

    public void setLinkLogout(String linkLogout) {
        this.linkLogout = linkLogout;
    }

    public String getLinkStatus() {
        return linkStatus;
    }

    public void setLinkStatus(String linkStatus) {
        this.linkStatus = linkStatus;
    }

    public Long getUptimeSecs() {
        return uptimeSecs;
    }

    public void setUptimeSecs(Long uptimeSecs) {
        this.uptimeSecs = uptimeSecs;
    }

    public Long getBytesIn() {
        return bytesIn;
    }

    public void setBytesIn(Long bytesIn) {
        this.bytesIn = bytesIn;
    }

    public Long getBytesOut() {
        return bytesOut;
    }

    public void setBytesOut(Long bytesOut) {
        this.bytesOut = bytesOut;
    }

    @Override
    public String toString() {
        return "RouterOSParam{" +
                "ssid='" + ssid + '\'' +
                ", mac='" + mac + '\'' +
                ", ip='" + ip + '\'' +
                ", username='" + username + '\'' +
                ", linkLogin='" + linkLogin + '\'' +
                ", linkLogout='" + linkLogout + '\'' +
                ", linkStatus='" + linkStatus + '\'' +
                ", linkOrig='" + linkOrig + '\'' +
                ", uptimeSecs=" + uptimeSecs +
                ", bytesIn=" + bytesIn +
                ", bytesOut=" + bytesOut +
                ", error='" + error + '\'' +
                '}';
    }
}
