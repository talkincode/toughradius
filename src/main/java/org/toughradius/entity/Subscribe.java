package org.toughradius.entity;

import java.math.BigDecimal;
import java.math.BigInteger;
import java.sql.Timestamp;
import java.util.Date;

public class Subscribe {
    private Integer id;

    private Integer nodeId;

    private Integer areaId;

    private Integer productId;

    private String subscriber;

    private String realname;

    private String password;

    private String billType;

    private String domain;

    private String addrPool;

    private String policy;

    private Integer activeNum;

    private Integer isOnline;

    private BigInteger flowAmount;

    private Boolean bindMac;

    private Boolean bindVlan;

    private String ipAddr;

    private String macAddr;

    private Integer inVlan;

    private Integer outVlan;

    private BigDecimal upRate;

    private BigDecimal downRate;

    private BigDecimal upPeakRate;

    private BigDecimal downPeakRate;

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

    public Integer getId() {
        return id;
    }

    public void setId(Integer id) {
        this.id = id;
    }

    public Integer getProductId() {
        return productId;
    }

    public void setProductId(Integer productId) {
        this.productId = productId;
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

    public String getBillType() {
        return billType;
    }

    public void setBillType(String billType) {
        this.billType = billType == null ? null : billType.trim();
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

    public BigInteger getFlowAmount() {
        return flowAmount;
    }

    public void setFlowAmount(BigInteger flowAmount) {
        this.flowAmount = flowAmount;
    }

    public Boolean getBindMac() {
        return bindMac;
    }

    public void setBindMac(Boolean bindMac) {
        this.bindMac = bindMac;
    }

    public Boolean getBindVlan() {
        return bindVlan;
    }

    public void setBindVlan(Boolean bindVlan) {
        this.bindVlan = bindVlan;
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

    public BigDecimal getUpRate() {
        return upRate;
    }

    public void setUpRate(BigDecimal upRate) {
        this.upRate = upRate;
    }

    public BigDecimal getDownRate() {
        return downRate;
    }

    public void setDownRate(BigDecimal downRate) {
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

    public Integer getNodeId() {
        return nodeId;
    }

    public void setNodeId(Integer nodeId) {
        this.nodeId = nodeId;
    }

    public BigDecimal getUpPeakRate() {
        return upPeakRate;
    }

    public void setUpPeakRate(BigDecimal upPeakRate) {
        this.upPeakRate = upPeakRate;
    }

    public BigDecimal getDownPeakRate() {
        return downPeakRate;
    }

    public void setDownPeakRate(BigDecimal downPeakRate) {
        this.downPeakRate = downPeakRate;
    }

    public Integer getAreaId() {
        return areaId;
    }

    public void setAreaId(Integer areaId) {
        this.areaId = areaId;
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
}