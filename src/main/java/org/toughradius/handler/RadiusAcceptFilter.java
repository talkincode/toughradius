package org.toughradius.handler;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.tinyradius.packet.AccessAccept;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.ValidateUtil;
import org.toughradius.component.OptionService;
import org.toughradius.entity.Nas;
import org.toughradius.entity.User;

import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.Date;

@Component
public class RadiusAcceptFilter implements RadiusConstant{

    @Autowired
    private OptionService configService;

    /**
     * Radius 认证成功后下发属性处理
     * @param accept
     * @param nas
     * @param user
     * @return
     */
    public AccessAccept doFilter(AccessAccept accept, Nas nas, User user){
        accept = filterDefault(accept,user, nas);
        switch (nas.getVendorid()){
            case VENDOR_MIKROTIK:
                return filterMikrotik(accept, user);
            case VENDOR_HUAWEI:
                return filterHuawei(accept, user);
            case VENDOR_H3C:
                return filterH3c(accept, user);
            case VENDOR_ZTE:
                return filterZTE(accept, user);
            case VENDOR_RADBACK:
                return filterRadback(accept, user);
            default:
                return accept;
        }
    }

    /**
     * 默认属性下发
     * @param accept
     * @param user
     * @return
     */
    private AccessAccept filterDefault(AccessAccept accept, User user, Nas nas){
        int sessionTimeout  = DateTimeUtil.compareSecond(user.getExpireTime(),new Date());
        long preSessionTimeout = accept.getPreSessionTimeout();
        if(preSessionTimeout>Integer.MAX_VALUE){
            preSessionTimeout = Integer.MAX_VALUE;
        }
        if(preSessionTimeout!=0){
            sessionTimeout = (int)preSessionTimeout;
        }
        int interimTimes = accept.getPreInterim();
        Integer dbInterimTimes = configService.getInterimTimes();
        if(dbInterimTimes!=null){
            interimTimes = dbInterimTimes;
        }

        if(ValidateUtil.isNotEmpty(user.getAddrPool())){
            accept.addAttribute("Framed-Pool",user.getAddrPool());
        }

        String ipaddr = user.getIpAddr();
        if(ValidateUtil.isNotEmpty(ipaddr) && ValidateUtil.isIP(ipaddr)){
            accept.addAttribute("Framed-IP-Address", ipaddr);
        }

        accept.addAttribute("Session-Timeout", String.valueOf(sessionTimeout));
        accept.addAttribute("Acct-Interim-Interval", String.valueOf(interimTimes));

        return accept;
    }


    private AccessAccept filterMikrotik(AccessAccept accept, User user){
        int up = user.getUpRate().multiply(BigInteger.valueOf(1024)).intValue();
        int down = user.getDownRate().multiply(BigInteger.valueOf(1024)).intValue();
        accept.addAttribute("Mikrotik-Rate-Limit", String.format("%sk/%sk", up,down));
        return accept;
    }

    private AccessAccept filterHuawei(AccessAccept accept, User user){
        int up = user.getUpRate().multiply(BigInteger.valueOf(1024*1024)).intValue();
        int down = user.getDownRate().multiply(BigInteger.valueOf(1024*1024)).intValue();
        int peakUp = up * 4;
        int peakDown = down * 4;
        try{
            peakUp = user.getUpPeakRate().multiply(BigInteger.valueOf(1024*1024)).intValue();
            peakDown = user.getDownPeakRate().multiply(BigInteger.valueOf(1024*1024)).intValue();
        }catch(Exception e){
            e.printStackTrace();
        }
        accept.addAttribute("Huawei-Input-Average-Rate", String.valueOf(up));
        accept.addAttribute("Huawei-Input-Peak-Rate", String.valueOf(peakUp));
        accept.addAttribute("Huawei-Output-Average-Rate", String.valueOf(down));
        accept.addAttribute("Huawei-Output-Peak-Rate", String.valueOf(peakDown));

        String domain = user.getDomain();
        if(ValidateUtil.isNotEmpty(domain)){
            accept.addAttribute("Huawei-Domain-Name", domain);
        }

        return accept;
    }

    private AccessAccept filterH3c(AccessAccept accept, User user){
        int up = user.getUpRate().multiply(BigInteger.valueOf(1024*1024)).intValue();
        int down = user.getDownRate().multiply(BigInteger.valueOf(1024*1024)).intValue();
        int peakUp = up * 4;
        int peakDown = down * 4;
        try{
            peakUp = user.getUpPeakRate().multiply(BigInteger.valueOf(1024*1024)).intValue();
            peakDown = user.getDownPeakRate().multiply(BigInteger.valueOf(1024*1024)).intValue();
        }catch(Exception e){
            e.printStackTrace();
        }
        accept.addAttribute("H3C-Input-Average-Rate", String.valueOf(up));
        accept.addAttribute("H3C-Input-Peak-Rate", String.valueOf(peakUp));
        accept.addAttribute("H3C-Output-Average-Rate", String.valueOf(down));
        accept.addAttribute("H3C-Output-Peak-Rate", String.valueOf(peakDown));
        return accept;
    }

    private AccessAccept filterZTE(AccessAccept accept, User user){
        int up = user.getUpRate().multiply(BigInteger.valueOf(1024*1024)).intValue();
        int down = user.getDownRate().multiply(BigInteger.valueOf(1024*1024)).intValue();
        accept.addAttribute("ZTE-Rate-Ctrl-Scr-Up", String.valueOf(up));
        accept.addAttribute("ZTE-Rate-Ctrl-Scr-Down", String.valueOf(down));
        return accept;
    }

    private AccessAccept filterRadback(AccessAccept accept, User user){
        String policy = user.getPolicy();
        if(ValidateUtil.isNotEmpty(policy))
            accept.addAttribute("Sub-Profile-Name", policy);

        String domain = user.getDomain();
        if(ValidateUtil.isNotEmpty(domain)){
            accept.addAttribute("Context-Name", domain);
        }

        return accept;
    }

}
