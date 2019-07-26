package org.toughradius.form;

public class FreeradiusAuthRequest {

    private String username;
    private String nasip;
    private String nasid;

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public String getNasip() {
        return nasip;
    }

    public void setNasip(String nasip) {
        this.nasip = nasip;
    }

    public String getNasid() {
        return nasid;
    }

    public void setNasid(String nasid) {
        this.nasid = nasid;
    }
}
