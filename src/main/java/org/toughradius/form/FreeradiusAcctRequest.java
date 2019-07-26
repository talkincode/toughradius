package org.toughradius.form;

public class FreeradiusAcctRequest {

    private String username;
    private String nasip;
    private String nasid;
    private String macAddr;
    private String nasPortId;
    private String framedIPAddress;
    private String acctSessionId;
    private String acctSessionTime;
    private Long acctInputOctets;
    private Long acctOutputOctets;
    private Long acctInputGigawords;
    private Long acctOutputGigawords;
    private Integer acctInputPackets;
    private Integer acctOutputPackets;

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

    public String getMacAddr() {
        return macAddr;
    }

    public void setMacAddr(String macAddr) {
        this.macAddr = macAddr;
    }

    public String getNasPortId() {
        return nasPortId;
    }

    public void setNasPortId(String nasPortId) {
        this.nasPortId = nasPortId;
    }

    public String getFramedIPAddress() {
        return framedIPAddress;
    }

    public void setFramedIPAddress(String framedIPAddress) {
        this.framedIPAddress = framedIPAddress;
    }

    public String getAcctSessionId() {
        return acctSessionId;
    }

    public void setAcctSessionId(String acctSessionId) {
        this.acctSessionId = acctSessionId;
    }

    public String getAcctSessionTime() {
        return acctSessionTime;
    }

    public void setAcctSessionTime(String acctSessionTime) {
        this.acctSessionTime = acctSessionTime;
    }

    public Long getAcctInputOctets() {
        return acctInputOctets;
    }

    public void setAcctInputOctets(Long acctInputOctets) {
        this.acctInputOctets = acctInputOctets;
    }

    public Long getAcctOutputOctets() {
        return acctOutputOctets;
    }

    public void setAcctOutputOctets(Long acctOutputOctets) {
        this.acctOutputOctets = acctOutputOctets;
    }

    public Long getAcctInputGigawords() {
        return acctInputGigawords;
    }

    public void setAcctInputGigawords(Long acctInputGigawords) {
        this.acctInputGigawords = acctInputGigawords;
    }

    public Long getAcctOutputGigawords() {
        return acctOutputGigawords;
    }

    public void setAcctOutputGigawords(Long acctOutputGigawords) {
        this.acctOutputGigawords = acctOutputGigawords;
    }

    public Integer getAcctInputPackets() {
        return acctInputPackets;
    }

    public void setAcctInputPackets(Integer acctInputPackets) {
        this.acctInputPackets = acctInputPackets;
    }

    public Integer getAcctOutputPackets() {
        return acctOutputPackets;
    }

    public void setAcctOutputPackets(Integer acctOutputPackets) {
        this.acctOutputPackets = acctOutputPackets;
    }
}
