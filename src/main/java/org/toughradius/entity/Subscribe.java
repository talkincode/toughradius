package org.toughradius.entity;

import java.sql.Timestamp;

public class Subscribe {
    private Long id;

    private Long nodeId;

    private String subscriber;

    private String realname;

    private String password;

    private String domain;

    private String addrPool;

    private String policy;

    private Integer activeNum;

    private Integer isOnline;

    private Integer bindMac;

    private Integer bindVlan;

    private String ipAddr;

    private String macAddr;

    private Integer inVlan;

    private Integer outVlan;

    private Long upRate;

    private Long downRate;

    private Long upPeakRate;

    private Long downPeakRate;

    private String upRateCode;

    private String downRateCode;

    private String status;

    private Timestamp beginTime;

    private Timestamp expireTime;

    private Timestamp createTime;

    private Timestamp updateTime;

    private String remark;

    public String getRealname() {
        return realname;
    }

    public void setRealname(String realname) {
        this.realname = realname;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getSubscriber() {
        return subscriber;
    }

    public void setSubscriber(String subscriber) {
        this.subscriber = subscriber == null ? null : subscriber.trim();
    }

    public String getPassword() {
        return password;
    }

    public void setPassword(String password) {
        this.password = password == null ? null : password.trim();
    }

    public String getDomain() {
        return domain;
    }

    public void setDomain(String domain) {
        this.domain = domain == null ? null : domain.trim();
    }

    public String getAddrPool() {
        return addrPool;
    }

    public void setAddrPool(String addrPool) {
        this.addrPool = addrPool == null ? null : addrPool.trim();
    }

    public String getPolicy() {
        return policy;
    }

    public void setPolicy(String policy) {
        this.policy = policy == null ? null : policy.trim();
    }

    public Integer getActiveNum() {
        return activeNum;
    }

    public void setActiveNum(Integer activeNum) {
        this.activeNum = activeNum;
    }

    public String getIpAddr() {
        return ipAddr;
    }

    public void setIpAddr(String ipAddr) {
        this.ipAddr = ipAddr == null ? null : ipAddr.trim();
    }

    public String getMacAddr() {
        return macAddr;
    }

    public void setMacAddr(String macAddr) {
        this.macAddr = macAddr == null ? null : macAddr.trim();
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

    public Long getUpRate() {
        return upRate;
    }

    public void setUpRate(Long upRate) {
        this.upRate = upRate;
    }

    public Long getDownRate() {
        return downRate;
    }

    public void setDownRate(Long downRate) {
        this.downRate = downRate;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status == null ? null : status.trim();
    }

    public Timestamp getBeginTime() {
        return beginTime;
    }

    public void setBeginTime(Timestamp beginTime) {
        this.beginTime = beginTime;
    }

    public Timestamp getExpireTime() {
        return expireTime;
    }

    public void setExpireTime(Timestamp expireTime) {
        this.expireTime = expireTime;
    }

    public Timestamp getCreateTime() {
        return createTime;
    }

    public void setCreateTime(Timestamp createTime) {
        this.createTime = createTime;
    }

    public Timestamp getUpdateTime() {
        return updateTime;
    }

    public void setUpdateTime(Timestamp updateTime) {
        this.updateTime = updateTime;
    }

    public String getUpRateCode() {
        return upRateCode;
    }

    public void setUpRateCode(String upRateCode) {
        this.upRateCode = upRateCode;
    }

    public String getDownRateCode() {
        return downRateCode;
    }

    public void setDownRateCode(String downRateCode) {
        this.downRateCode = downRateCode;
    }

    public Long getNodeId() {
        return nodeId;
    }

    public void setNodeId(Long nodeId) {
        this.nodeId = nodeId;
    }

    public Long getUpPeakRate() {
        return upPeakRate;
    }

    public void setUpPeakRate(Long upPeakRate) {
        this.upPeakRate = upPeakRate;
    }

    public Long getDownPeakRate() {
        return downPeakRate;
    }

    public void setDownPeakRate(Long downPeakRate) {
        this.downPeakRate = downPeakRate;
    }

    public Integer getIsOnline() {
        return isOnline;
    }

    public void setIsOnline(Integer isOnline) {
        this.isOnline = isOnline;
    }

    public String getRemark() {
        return remark;
    }

    public void setRemark(String remark) {
        this.remark = remark;
    }

    public Integer getBindMac() {
        return bindMac;
    }

    public void setBindMac(Integer bindMac) {
        this.bindMac = bindMac;
    }

    public Integer getBindVlan() {
        return bindVlan;
    }

    public void setBindVlan(Integer bindVlan) {
        this.bindVlan = bindVlan;
    }
}