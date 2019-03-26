package org.toughradius.handler;

import org.toughradius.common.DateTimeUtil;
import org.toughradius.entity.Bras;
import org.toughradius.entity.Subscribe;
import org.tinyradius.packet.AccessAccept;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.junit4.SpringRunner;

import java.math.BigDecimal;
import java.sql.Timestamp;
import java.util.Date;

@RunWith(SpringRunner.class)
@SpringBootTest
public class RadiusAcceptFilterTest {

    @Autowired
    private RadiusAcceptFilter acceptFilter;

    @Test
    public void testFilterDefault(){
        AccessAccept accept = new AccessAccept(1);
        accept.setPreInterim(300);
        accept.setPreSessionTimeout(86400);
        Bras nas = new Bras();
        nas.setVendorId("0");
        Subscribe user = new Subscribe();
        user.setIpAddr("192.168.9.9");
        Timestamp expire = DateTimeUtil.toTimestamp("2019-12-12 00:00:00");
        user.setExpireTime(expire);
        AccessAccept aa = acceptFilter.doFilter(accept,nas,user);
        System.out.println(aa.toString());
        assert aa.getSessionTimeout()  == 86400;
    }

    @Test
    public void testFilterDefaultProxy(){
        AccessAccept accept = new AccessAccept(1);
        accept.setPreInterim(300);
        accept.setPreSessionTimeout(86400);
        Bras nas = new Bras();
        nas.setVendorId("0");
        Subscribe user = new Subscribe();
        user.setIpAddr("192.168.9.9");
        user.setSubscriber("tets01");
        user.setPassword("888888");
        Timestamp expire = DateTimeUtil.toTimestamp("2019-12-12 00:00:00");
        user.setExpireTime(expire);
        AccessAccept aa = acceptFilter.doFilter(accept,nas,user);
        System.out.println(aa.toString());
        assert aa.getSessionTimeout()  == 86400;
    }

    @Test
    public void testFilterMIKROTIK(){
        AccessAccept accept = new AccessAccept(1);
        accept.setPreInterim(300);
        accept.setPreSessionTimeout(86400);
        Bras nas = new Bras();
        nas.setVendorId("14988");
        Subscribe user = new Subscribe();
        user.setIpAddr("192.168.9.9");
        user.setUpRate(new BigDecimal(10));
        user.setDownRate(new BigDecimal(10));
        Timestamp expire = DateTimeUtil.toTimestamp("2019-12-12 00:00:00");
        user.setExpireTime(expire);
        AccessAccept aa = acceptFilter.doFilter(accept,nas,user);
        System.out.println(aa.toString());
        assert aa.getSessionTimeout()  == 86400;
        assert aa.getAttribute("Mikrotik-Rate-Limit").getStringValue().equalsIgnoreCase("10240k/10240k");
    }

    @Test
    public void testFilterHuawei(){
        AccessAccept accept = new AccessAccept(1);
        accept.setPreInterim(300);
        accept.setPreSessionTimeout(86400);
        Bras nas = new Bras();
        nas.setVendorId("2011");
        Subscribe user = new Subscribe();
        user.setDomain("hncatv");
        user.setIpAddr("192.168.9.9");
        user.setUpRate(new BigDecimal(10));
        user.setUpPeakRate(new BigDecimal(30));
        user.setDownRate(new BigDecimal(10));
        user.setDownPeakRate(new BigDecimal(30));
        Timestamp expire = DateTimeUtil.toTimestamp("2019-12-12 00:00:00");
        user.setExpireTime(expire);
        AccessAccept aa = acceptFilter.doFilter(accept,nas,user);
        System.out.println(aa.toString());
        assert aa.getSessionTimeout()  == 86400;
        assert aa.getAttribute("Huawei-Domain-Name").getStringValue().equalsIgnoreCase("hncatv");

    }

    @Test
    public void testFilterRadback(){
        AccessAccept accept = new AccessAccept(1);
        accept.setPreInterim(300);
        accept.setPreSessionTimeout(86400);
        Bras nas = new Bras();
        nas.setVendorId("2352");
        Subscribe user = new Subscribe();
        user.setDomain("hncatv");
        user.setIpAddr("192.168.9.9");
        user.setPolicy("dm10m");
        Timestamp expire = DateTimeUtil.toTimestamp("2019-12-12 00:00:00");
        user.setExpireTime(expire);
        AccessAccept aa = acceptFilter.doFilter(accept,nas,user);
        System.out.println(aa.toString());
        assert aa.getSessionTimeout()  == 86400;
        assert aa.getAttribute("Context-Name").getStringValue().equalsIgnoreCase("hncatv");

    }


}
