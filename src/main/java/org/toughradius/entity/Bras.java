package org.toughradius.entity;

import sun.rmi.runtime.Log;

import java.io.Serializable;
import java.sql.Timestamp;
import java.util.Date;

public class Bras implements Serializable{

    private Long id;

    private String identifier;

    private String name;

    private String ipaddr;

    private String vendorId;

    private String secret;

    private Integer coaPort;

    private String status;

    private Integer acPort;

    private Integer authLimit;

    private Integer acctLimit;

    private String portalVendor;

    private String remark;

    private Timestamp createTime;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
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

    public Timestamp getCreateTime() {
        return createTime;
    }

    public void setCreateTime(Timestamp createTime) {
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

    public Integer getAcPort() {
        return acPort;
    }

    public void setAcPort(Integer acPort) {
        this.acPort = acPort;
    }

    public String getPortalVendor() {
        return portalVendor;
    }

    public void setPortalVendor(String portalVendor) {
        this.portalVendor = portalVendor;
    }
}