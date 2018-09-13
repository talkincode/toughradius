package org.toughradius.entity;

import org.toughradius.common.DateTimeUtil;

import java.util.Date;

public class Ticket {
    private Integer id;

    private Integer groupId;

    private String username;

    private String nasId;

    private String nasAddr;

    private String nasPaddr;

    private Integer sessionTimeout;

    private String framedIpaddr;

    private String framedNetmask;

    private String macAddr;

    private Long nasPort;

    private String nasClass;

    private String nasPortId;

    private Integer nasPortType;

    private Integer serviceType;

    private String acctSessionId;

    private Integer acctSessionTime;

    private Long acctInputTotal;

    private Long acctOutputTotal;

    private Integer acctInputPackets;

    private Integer acctOutputPackets;

    private Date acctStartTime;

    private Date acctStopTime;

    public Integer getId() {
        return id;
    }

    public void setId(Integer id) {
        this.id = id;
    }

    public Integer getGroupId() {
        return groupId;
    }

    public void setGroupId(Integer groupId) {
        this.groupId = groupId;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username == null ? null : username.trim();
    }

    public String getNasId() {
        return nasId;
    }

    public void setNasId(String nasId) {
        this.nasId = nasId == null ? null : nasId.trim();
    }

    public String getNasAddr() {
        return nasAddr;
    }

    public void setNasAddr(String nasAddr) {
        this.nasAddr = nasAddr == null ? null : nasAddr.trim();
    }

    public String getNasPaddr() {
        return nasPaddr;
    }

    public void setNasPaddr(String nasPaddr) {
        this.nasPaddr = nasPaddr == null ? null : nasPaddr.trim();
    }

    public Integer getSessionTimeout() {
        return sessionTimeout;
    }

    public void setSessionTimeout(Integer sessionTimeout) {
        this.sessionTimeout = sessionTimeout;
    }

    public String getFramedIpaddr() {
        return framedIpaddr;
    }

    public void setFramedIpaddr(String framedIpaddr) {
        this.framedIpaddr = framedIpaddr == null ? null : framedIpaddr.trim();
    }

    public String getFramedNetmask() {
        return framedNetmask;
    }

    public void setFramedNetmask(String framedNetmask) {
        this.framedNetmask = framedNetmask == null ? null : framedNetmask.trim();
    }

    public String getMacAddr() {
        return macAddr;
    }

    public void setMacAddr(String macAddr) {
        this.macAddr = macAddr == null ? null : macAddr.trim();
    }

    public Long getNasPort() {
        return nasPort;
    }

    public void setNasPort(Long nasPort) {
        this.nasPort = nasPort;
    }

    public String getNasClass() {
        return nasClass;
    }

    public void setNasClass(String nasClass) {
        this.nasClass = nasClass == null ? null : nasClass.trim();
    }

    public String getNasPortId() {
        return nasPortId;
    }

    public void setNasPortId(String nasPortId) {
        this.nasPortId = nasPortId == null ? null : nasPortId.trim();
    }

    public Integer getNasPortType() {
        return nasPortType;
    }

    public void setNasPortType(Integer nasPortType) {
        this.nasPortType = nasPortType;
    }

    public Integer getServiceType() {
        return serviceType;
    }

    public void setServiceType(Integer serviceType) {
        this.serviceType = serviceType;
    }

    public String getAcctSessionId() {
        return acctSessionId;
    }

    public void setAcctSessionId(String acctSessionId) {
        this.acctSessionId = acctSessionId == null ? null : acctSessionId.trim();
    }

    public Integer getAcctSessionTime() {
        return acctSessionTime;
    }

    public void setAcctSessionTime(Integer acctSessionTime) {
        this.acctSessionTime = acctSessionTime;
    }

    public Long getAcctInputTotal() {
        return acctInputTotal;
    }

    public void setAcctInputTotal(Long acctInputTotal) {
        this.acctInputTotal = acctInputTotal;
    }

    public Long getAcctOutputTotal() {
        return acctOutputTotal;
    }

    public void setAcctOutputTotal(Long acctOutputTotal) {
        this.acctOutputTotal = acctOutputTotal;
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

    public Date getAcctStartTime() {
        return acctStartTime;
    }

    public void setAcctStartTime(Date acctStartTime) {
        this.acctStartTime = acctStartTime;
    }

    public Date getAcctStopTime() {
        return acctStopTime;
    }

    public void setAcctStopTime(Date acctStopTime) {
        this.acctStopTime = acctStopTime;
    }


    public static String getHeaderString(){
        StringBuilder buff = new StringBuilder();
        buff.append("groupId").append(",");
        buff.append("username").append(",");
        buff.append("nasId").append(",");
        buff.append("nasAddr").append(",");
        buff.append("nasPaddr").append(",");
        buff.append("sessionTimeout").append(",");
        buff.append("framedIpaddr").append(",");
        buff.append("framedNetmask").append(",");
        buff.append("macAddr").append(",");
        buff.append("nasPort").append(",");
        buff.append("nasClass").append(",");
        buff.append("nasPortId").append(",");
        buff.append("nasPortType").append(",");
        buff.append("serviceType").append(",");
        buff.append("acctSessionId").append(",");
        buff.append("acctSessionTime").append(",");
        buff.append("acctInputTotal").append(",");
        buff.append("acctOutputTotal").append(",");
        buff.append("acctInputPackets").append(",");
        buff.append("acctOutputPackets").append(",");
        buff.append("acctStartTime").append(",");
        buff.append("acctStopTime");
        return buff.toString();
    }

    public String toString(){
        StringBuilder buff = new StringBuilder();
        buff.append(groupId).append(",");
        buff.append(username).append(",");
        buff.append(nasId).append(",");
        buff.append(nasAddr).append(",");
        buff.append(nasPaddr).append(",");
        buff.append(sessionTimeout).append(",");
        buff.append(framedIpaddr).append(",");
        buff.append(framedNetmask).append(",");
        buff.append(macAddr).append(",");
        buff.append(nasPort).append(",");
        buff.append(nasClass).append(",");
        buff.append(nasPortId).append(",");
        buff.append(nasPortType).append(",");
        buff.append(serviceType).append(",");
        buff.append(acctSessionId).append(",");
        buff.append(acctSessionTime).append(",");
        buff.append(acctInputTotal).append(",");
        buff.append(acctOutputTotal).append(",");
        buff.append(acctInputPackets).append(",");
        buff.append(acctOutputPackets).append(",");
        buff.append(DateTimeUtil.toDateTimeString(acctStartTime)).append(",");
        buff.append(DateTimeUtil.toDateTimeString(acctStopTime));
        return buff.toString();
    }

    public static Ticket fromString(String line) {
        try{
            String [] strs = line.trim().split(",");
            if(strs.length!=22){
                return null;
            }
            Ticket log = new Ticket();
            log.setGroupId(Integer.valueOf(strs[0]));
            log.setUsername(strs[1]);
            log.setNasId(strs[2]);
            log.setNasAddr(strs[3]);
            log.setNasPaddr(strs[4]);
            log.setSessionTimeout(Integer.valueOf(strs[5]));
            log.setFramedIpaddr(strs[6]);
            log.setFramedNetmask(strs[7]);
            log.setMacAddr(strs[8]);
            log.setNasPort(Long.valueOf(strs[9]));
            log.setNasClass(strs[10]);
            log.setNasPortId(strs[11]);
            log.setNasPortType(Integer.valueOf(strs[12]));
            log.setServiceType(Integer.valueOf(strs[13]));
            log.setAcctSessionId(strs[14]);
            log.setAcctSessionTime(Integer.valueOf(strs[15]));
            log.setAcctInputTotal(Long.valueOf(strs[16]));
            log.setAcctOutputTotal(Long.valueOf(strs[17]));
            log.setAcctInputPackets(Integer.valueOf(strs[18]));
            log.setAcctOutputPackets(Integer.valueOf(strs[19]));
            log.setAcctStartTime(DateTimeUtil.toDate(strs[20]));
            log.setAcctStopTime(DateTimeUtil.toDate(strs[21]));
            return log;
        } catch(Exception e){
            return null;
        }
    }
}