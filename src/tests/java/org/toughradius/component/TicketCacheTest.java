package org.toughradius.component;

import org.toughradius.entity.RadiusTicket;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

import java.util.Date;

@RunWith(SpringRunner.class)
@SpringBootTest
public class TicketCacheTest {

    @Autowired
    private TicketCache ticketCache;

    @Test
    public void testWriteTicket(){
        for (int i=0;i<=10000;i++){
            RadiusTicket radiusTicket = new RadiusTicket();
            radiusTicket.setNodeId(1);
            radiusTicket.setAreaId(1);
            radiusTicket.setUsername("test");
            radiusTicket.setNasId("radius-tester");
            radiusTicket.setNasAddr("10.10.10.10");
            radiusTicket.setNasPaddr("10.10.10.10");
            radiusTicket.setAcctSessionTime(123123141);
            radiusTicket.setSessionTimeout(324234234);
            radiusTicket.setFramedIpaddr("192.134.3.3");
            radiusTicket.setFramedNetmask("255.255.255.0");
            radiusTicket.setMacAddr("24:23:00:0:00:00");
            radiusTicket.setNasPort(0L);
            radiusTicket.setNasClass("");
            radiusTicket.setNasPortId("");
            radiusTicket.setNasPortType(0);
            radiusTicket.setServiceType(0);
            radiusTicket.setAcctSessionId("sdfsdfdsf");
            radiusTicket.setAcctInputTotal(12112L);
            radiusTicket.setAcctOutputTotal(345345L);
            radiusTicket.setAcctInputPackets(345);
            radiusTicket.setAcctOutputPackets(435435);
            radiusTicket.setAcctStopTime(new Date());
            radiusTicket.setAcctStartTime(new Date());
            ticketCache.addTicket(radiusTicket);
        }
        ticketCache.syncData();
    }
}
