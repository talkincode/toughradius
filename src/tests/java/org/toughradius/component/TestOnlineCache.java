package org.toughradius.component;

import org.toughradius.component.OnlineCache;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.PageResult;
import org.toughradius.entity.RadiusOnline;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

@RunWith(SpringRunner.class)
@SpringBootTest
public class TestOnlineCache {

    @Autowired
    private OnlineCache onlineCache;


    @Test
    public void testAddOnline(){
        for(int i=0;i<1000; i++){
            RadiusOnline online = new RadiusOnline();
            online.setNodeId(1);
            online.setAreaId(1);
            online.setUsername("username"+i);
            online.setNasId("radius-tester");
            online.setNasAddr("10.10.10.10");
            online.setNasPaddr("10.10.10.10");
            online.setSessionTimeout(86400);
            online.setFramedIpaddr("10.10.10.10");
            online.setFramedNetmask("255.255.255.255");
            online.setMacAddr("1q:e3:d4");
            online.setNasPort(12L);
            online.setNasClass("class");
            online.setNasPortId("0/0");
            online.setServiceType(0);
            online.setAcctSessionId("100000"+i);
            online.setAcctSessionTime(86400);
            online.setAcctInputTotal(0L);
            online.setAcctOutputTotal(0L);
            online.setAcctInputPackets(0);
            online.setAcctOutputPackets(0);
            online.setAcctStartTime(DateTimeUtil.getDateTimeString());
            onlineCache.putOnline(online);
        }
        System.out.println(onlineCache.size());
        assert  onlineCache.size() == 1000;
        PageResult res = onlineCache.queryOnlinePage(0,10,"1","1", null,null,"","","","","" ,"");
        System.out.println(res.getData().size());
        assert res.getData().size() == 10;
        PageResult res2 = onlineCache.queryOnlinePage(80,40,"1","1", null,null,"","","","","" ,"");
        System.out.println(res2.getData().size());
        assert res2.getData().size() == 40;
        PageResult res3 = onlineCache.queryOnlinePage(0,10,"1","1", null,null,"","","","","username19","" );
        System.out.println(res3.getData().size());
        assert res3.getData().size() == 10;
    }




}
