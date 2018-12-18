package org.toughradius.entity;

import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.Date;

public class User {

    private String code;

    private String groupCode;

    private String fullname;

    private String email;

    private String mobile;

    private String username;

    private String password;

    private String billType;

    private String domain;

    private String addrPool;

    private Integer onlineNum;

    private BigInteger flowAmount;

    private Integer bindMac;

    private Integer bindVlan;

    private String ipAddr;

    private String macAddr;

    private Integer inVlan;

    private Integer outVlan;

    private BigInteger upRate;

    private BigInteger downRate;


    private String rateCode;

    private String upRateCode;

    private String downRateCode;


    private String status;

    private String remark;

    private Integer isOnline;

    private Date lastOnline;
    private Date lastOffline;
    private Date expireTime;

    private Date createTime;

    private Date updateTime;

    public String getCode() {
        return code;
    }

    public void setCode(String code) {
        this.code = code;
    }

    public String getGroupCode() {
        return groupCode;
    }

    public void setGroupCode(String groupCode) {
        this.groupCode = groupCode;
    }

    public String getFullname() {
        return fullname;
    }

    public void setFullname(String fullname) {
        this.fullname = fullname;
    }

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public String getMobile() {
        return mobile;
    }

    public void setMobile(String mobile) {
        this.mobile = mobile;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public String getPassword() {
        return password;
    }

    public void setPassword(String password) {
        this.password = password;
    }

    public String getBillType() {
        return billType;
    }

    public void setBillType(String billType) {
        this.billType = billType;
    }

    public String getDomain() {
        return domain;
    }

    public void setDomain(String domain) {
        this.domain = domain;
    }

    public String getAddrPool() {
        return addrPool;
    }

    public void setAddrPool(String addrPool) {
        this.addrPool = addrPool;
    }

    public Integer getOnlineNum() {
        return onlineNum;
    }

    public void setOnlineNum(Integer onlineNum) {
        this.onlineNum = onlineNum;
    }

    public BigInteger getFlowAmount() {
        return flowAmount;
    }

    public void setFlowAmount(BigInteger flowAmount) {
        this.flowAmount = flowAmount;
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

    public String getIpAddr() {
        return ipAddr;
    }

    public void setIpAddr(String ipAddr) {
        this.ipAddr = ipAddr;
    }

    public String getMacAddr() {
        return macAddr;
    }

    public void setMacAddr(String macAddr) {
        this.macAddr = macAddr;
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

    public BigInteger getUpRate() {
        return upRate;
    }

    public void setUpRate(BigInteger upRate) {
        this.upRate = upRate;
    }

    public BigInteger getDownRate() {
        return downRate;
    }

    public void setDownRate(BigInteger downRate) {
        this.downRate = downRate;
    }

    public String getRateCode() {
        return rateCode;
    }

    public void setRateCode(String rateCode) {
        this.rateCode = rateCode;
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

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getRemark() {
        return remark;
    }

    public void setRemark(String remark) {
        this.remark = remark;
    }

    public Date getExpireTime() {
        return expireTime;
    }

    public void setExpireTime(Date expireTime) {
        this.expireTime = expireTime;
    }

    public Date getCreateTime() {
        return createTime;
    }

    public void setCreateTime(Date createTime) {
        this.createTime = createTime;
    }

    public Date getUpdateTime() {
        return updateTime;
    }

    public void setUpdateTime(Date updateTime) {
        this.updateTime = updateTime;
    }

    public Integer getIsOnline() {
        return isOnline;
    }

    public void setIsOnline(Integer isOnline) {
        this.isOnline = isOnline;
    }

    public Date getLastOnline() {
        return lastOnline;
    }

    public void setLastOnline(Date lastOnline) {
        this.lastOnline = lastOnline;
    }

    public Date getLastOffline() {
        return lastOffline;
    }

    public void setLastOffline(Date lastOffline) {
        this.lastOffline = lastOffline;
    }
}