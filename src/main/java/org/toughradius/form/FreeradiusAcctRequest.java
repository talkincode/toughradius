package org.toughradius.form;
import org.toughradius.common.DateTimeUtil;

public class FreeradiusAcctRequest {

    private String username;
    private String nasip;
    private String nasid;
    private String macAddr;
    private String nasPortId;
    private String framedIPAddress;
    private String acctSessionId;
    private Integer acctSessionTime;
    private Integer sessionTimeout;
    private String framedIPNetmask;
    private Long acctInputOctets;
    private Long acctOutputOctets;
    private Long acctInputGigawords;
    private Long acctOutputGigawords;
    private Integer acctInputPackets;
    private Integer acctOutputPackets;
    private String acctStatusType;

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

    public Integer getAcctSessionTime() {
        return acctSessionTime;
    }

    public void setAcctSessionTime(Integer acctSessionTime) {
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

    public String getAcctStatusType() {
        return acctStatusType;
    }

    public void setAcctStatusType(String acctStatusType) {
        this.acctStatusType = acctStatusType;
    }

    public Integer getSessionTimeout() {
        return sessionTimeout;
    }

    public void setSessionTimeout(Integer sessionTimeout) {
        this.sessionTimeout = sessionTimeout;
    }

    public String getFramedIPNetmask() {
        return framedIPNetmask;
    }

    public void setFramedIPNetmask(String framedIPNetmask) {
        this.framedIPNetmask = framedIPNetmask;
    }

    public long getAcctInputTotal(){
        try{
            Long ba = getAcctInputOctets();
            Long ga = getAcctInputGigawords();
            long b1 = ba!=null? ba :0L;
            long gl = ga != null ? ga :0L;
            long gb = gl * (4*1024*1024*1024);
            return b1 + gb;
        } catch(Exception e){
            return 0;
        }
    }

    public long getAcctOutputTotal(){
        try{
            Long ba = getAcctOutputOctets();
            Long ga = getAcctOutputGigawords();
            long b1 = ba!=null? ba :0L;
            long gl = ga != null ? ga :0L;
            long gb = gl * (4*1024*1024*1024);
            return b1 + gb;
        } catch(Exception e){
            return 0;
        }
    }

    public String getAcctStartTime(){
        try{
            Integer stime = getAcctSessionTime();
            int sstime = stime!=null?stime:0;
            return DateTimeUtil.getPreviousDateTimeBySecondString(sstime);
        } catch(Exception e){
            return DateTimeUtil.getDateTimeString();
        }
    }
}
