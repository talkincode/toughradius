package org.toughradius.entity;

import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.StringUtil;

import java.util.Date;

public class RadiusTicket {
    private Long id;

    private Long nodeId;

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

    private Integer inVlan;

    private Integer outVlan;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getNodeId() {
        return nodeId;
    }

    public void setNodeId(Long nodeId) {
        this.nodeId = nodeId;
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

    public Integer getInVlan() {
        return inVlan;
    }

    public void setInVlan(Integer inVlan) {
        this.inVlan = inVlan;
    }

    public Integer getOutVlan() {
        return outVlan;
    }

    public void setOutVlan(Integer outVlan) {
        this.outVlan = outVlan;
    }

    private static String safestr(Object src){
        if(src==null){
            return "";
        }
        return String.valueOf(src);
    }


    public static String getHeaderString(){
        StringBuilder buff = new StringBuilder();
        buff.append("id").append(",");
        buff.append("nodeId").append(",");
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
        buff.append("acctStopTime").append(",");
        buff.append("inVlan").append(",");
        buff.append("outVlan");
        return buff.toString();
    }

    public String toString(){
        StringBuilder buff = new StringBuilder();
        buff.append(safestr(id)).append(",");
        buff.append(safestr(nodeId)).append(",");
        buff.append(safestr(username)).append(",");
        buff.append(safestr(nasId)).append(",");
        buff.append(safestr(nasAddr)).append(",");
        buff.append(safestr(nasPaddr)).append(",");
        buff.append(String.valueOf(sessionTimeout)).append(",");
        buff.append(safestr(framedIpaddr)).append(",");
        buff.append(safestr(framedNetmask)).append(",");
        buff.append(safestr(macAddr)).append(",");
        buff.append(safestr(nasPort)).append(",");
        buff.append(safestr(nasClass)).append(",");
        buff.append(safestr(nasPortId)).append(",");
        buff.append(safestr(nasPortType)).append(",");
        buff.append(safestr(serviceType)).append(",");
        buff.append(safestr(acctSessionId)).append(",");
        buff.append(safestr(acctSessionTime)).append(",");
        buff.append(safestr(acctInputTotal)).append(",");
        buff.append(safestr(acctOutputTotal)).append(",");
        buff.append(safestr(acctInputPackets)).append(",");
        buff.append(safestr(acctOutputPackets)).append(",");
        buff.append(DateTimeUtil.toDateTimeString(acctStartTime)).append(",");
        buff.append(DateTimeUtil.toDateTimeString(acctStopTime)).append(",");
        buff.append(safestr(inVlan)).append(",");
        buff.append(safestr(outVlan));
        return buff.toString();
    }

    public static RadiusTicket fromString(String line) {
        try{
            String [] strs = line.trim().split(",");
            if(strs.length!=25){
                return null;
            }
            RadiusTicket log = new RadiusTicket();
            log.setId(Long.valueOf(strs[0]));
            log.setNodeId(Long.valueOf(strs[1]));
            log.setUsername(strs[2]);
            log.setNasId(strs[3]);
            log.setNasAddr(strs[4]);
            log.setNasPaddr(strs[5]);
            log.setSessionTimeout(Integer.valueOf(strs[6]));
            log.setFramedIpaddr(strs[7]);
            log.setFramedNetmask(strs[8]);
            log.setMacAddr(strs[9]);
            log.setNasPort(Long.valueOf(strs[10]));
            log.setNasClass(strs[11]);
            log.setNasPortId(strs[12]);
            log.setNasPortType(Integer.valueOf(strs[13]));
            log.setServiceType(Integer.valueOf(strs[14]));
            log.setAcctSessionId(strs[15]);
            log.setAcctSessionTime(Integer.valueOf(strs[16]));
            log.setAcctInputTotal(Long.valueOf(strs[17]));
            log.setAcctOutputTotal(Long.valueOf(strs[18]));
            log.setAcctInputPackets(Integer.valueOf(strs[19]));
            log.setAcctOutputPackets(Integer.valueOf(strs[20]));
            log.setAcctStartTime(DateTimeUtil.toDate(strs[21]));
            log.setAcctStopTime(DateTimeUtil.toDate(strs[22]));
            log.setInVlan(Integer.valueOf(strs[23]));
            log.setOutVlan(Integer.valueOf(strs[24]));
            return log;
        } catch(Exception e){
            return null;
        }
    }
}