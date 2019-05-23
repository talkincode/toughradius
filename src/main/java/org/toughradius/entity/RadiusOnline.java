package org.toughradius.entity;

public class RadiusOnline implements Cloneable{

    public final static int AMOUNT_NOT_ENOUGH = 1000;
    public final static int USER_NOT_EXIST = 1001;

    private Long id;

    private Long nodeId;

    private String realname;

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

    private String acctStartTime;

    private int unLockFlag;

    private Integer inVlan;

    private Integer outVlan;

    private boolean radsec;


    @Override
    public RadiusOnline clone() throws CloneNotSupportedException {
        return (RadiusOnline) super.clone();
    }

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

    public String getRealname() {
        return realname;
    }

    public void setRealname(String realname) {
        this.realname = realname;
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

    public String getAcctStartTime() {
        return acctStartTime;
    }

    public void setAcctStartTime(String acctStartTime) {
        this.acctStartTime = acctStartTime;
    }
    public int getUnLockFlag() {
        return unLockFlag;
    }

    public void setUnLockFlag(int unLockFlag) {
        this.unLockFlag = unLockFlag;
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

    public boolean isRadsec() {
        return radsec;
    }

    public void setRadsec(boolean radsec) {
        this.radsec = radsec;
    }
}