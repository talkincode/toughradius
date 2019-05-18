package org.toughradius.form;

public class RadiusConfigForm {

    private String radiusInterimIntelval;
    private String radiusIgnorePassword;
    private String radiusTicketHistoryDays;
    private String radiusExpireAddrPool;


    public String getRadiusInterimIntelval() {
        return radiusInterimIntelval;
    }

    public void setRadiusInterimIntelval(String radiusInterimIntelval) {
        this.radiusInterimIntelval = radiusInterimIntelval;
    }

    public String getRadiusIgnorePassword() {
        return radiusIgnorePassword;
    }

    public void setRadiusIgnorePassword(String radiusIgnorePassword) {
        this.radiusIgnorePassword = radiusIgnorePassword;
    }

    public String getRadiusTicketHistoryDays() {
        return radiusTicketHistoryDays;
    }

    public void setRadiusTicketHistoryDays(String radiusTicketHistoryDays) {
        this.radiusTicketHistoryDays = radiusTicketHistoryDays;
    }

    public String getRadiusExpireAddrPool() {
        return radiusExpireAddrPool;
    }

    public void setRadiusExpireAddrPool(String radiusExpireAddrPool) {
        this.radiusExpireAddrPool = radiusExpireAddrPool;
    }
}
