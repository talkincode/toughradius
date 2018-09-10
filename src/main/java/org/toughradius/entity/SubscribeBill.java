package org.toughradius.entity;

public class SubscribeBill {
    private String billType;
    private Integer flowAmount;

    public String getBillType() {
        return billType;
    }

    public void setBillType(String billType) {
        this.billType = billType;
    }

    public Integer getFlowAmount() {
        return flowAmount;
    }

    public void setFlowAmount(Integer flowAmount) {
        this.flowAmount = flowAmount;
    }
}
