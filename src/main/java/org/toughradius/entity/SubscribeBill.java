package org.toughradius.entity;

import java.math.BigInteger;

public class SubscribeBill {
    private String billType;
    private BigInteger flowAmount;

    public String getBillType() {
        return billType;
    }

    public void setBillType(String billType) {
        this.billType = billType;
    }

    public BigInteger getFlowAmount() {
        return flowAmount;
    }

    public void setFlowAmount(BigInteger flowAmount) {
        this.flowAmount = flowAmount;
    }
}
