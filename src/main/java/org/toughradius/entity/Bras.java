package org.toughradius.entity;

import java.io.Serializable;
import java.util.Date;

public class Bras implements Serializable{

    private Integer id;

    private String identifier;

    private String name;

    private String ipaddr;

    private String vendorId;

    private String secret;

    private Integer coaPort;

    private String status;

    private Integer authLimit;

    private Integer acctLimit;

    private String remark;

    private Date createTime;

    public Integer getId() {
        return id;
    }

    public void setId(Integer id) {
        this.id = id;
    }

    public String getIdentifier() {
        return identifier;
    }

    public void setIdentifier(String identifier) {
        this.identifier = identifier == null ? null : identifier.trim();
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name == null ? null : name.trim();
    }

    public String getIpaddr() {
        return ipaddr;
    }

    public void setIpaddr(String ipaddr) {
        this.ipaddr = ipaddr == null ? null : ipaddr.trim();
    }

    public String getVendorId() {
        return vendorId;
    }

    public void setVendorId(String vendorId) {
        this.vendorId = vendorId == null ? null : vendorId.trim();
    }

    public String getSecret() {
        return secret;
    }

    public void setSecret(String secret) {
        this.secret = secret == null ? null : secret.trim();
    }

    public Integer getCoaPort() {
        return coaPort;
    }

    public void setCoaPort(Integer coaPort) {
        this.coaPort = coaPort;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status == null ? null : status.trim();
    }

    public String getRemark() {
        return remark;
    }

    public void setRemark(String remark) {
        this.remark = remark == null ? null : remark.trim();
    }

    public Date getCreateTime() {
        return createTime;
    }

    public void setCreateTime(Date createTime) {
        this.createTime = createTime;
    }

    public Integer getAuthLimit() {
        return authLimit;
    }

    public void setAuthLimit(Integer authLimit) {
        this.authLimit = authLimit;
    }

    public Integer getAcctLimit() {
        return acctLimit;
    }

    public void setAcctLimit(Integer acctLimit) {
        this.acctLimit = acctLimit;
    }
}