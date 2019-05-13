package org.toughradius.handler;

import org.toughradius.common.ValidateUtil;
import org.toughradius.component.Memarylogger;
import org.toughradius.entity.Bras;
import org.tinyradius.packet.RadiusPacket;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.regex.Matcher;
import java.util.regex.Pattern;


@Component
public class RadiusParseFilter implements RadiusConstant{

    @Autowired
    public Memarylogger logger;

    private final static Pattern pattern = Pattern.compile("\\w?\\s?\\d+/\\d+/\\d+:(\\d+).(\\d+)\\s?");

    /**
     * 请求预处理
     * @param request
     * @param nas
     * @return
     */
    public RadiusPacket doFilter(RadiusPacket request, Bras nas){
        try{
            parseCiscoVlan(request);
            switch (nas.getVendorId()) {
                case VENDOR_MIKROTIK:
                    return filterMikrotik(request);
                case VENDOR_HUAWEI:
                    return filterHuawei(request);
                case VENDOR_H3C:
                    return filterH3c(request);
                case VENDOR_ZTE:
                    return filterZTE(request);
                case VENDOR_RADBACK:
                    return filterRadback(request);
                default:
                    return filterDefault(request);
            }
        } catch(Exception e){
            logger.error(e.getMessage(), Memarylogger.RADIUSD);
            return request;
        }
    }

    /**
     * 默认解析处理
     * @param request
     * @return
     */
    private RadiusPacket filterDefault(RadiusPacket request){
        request.setMacAddr(request.getCallingStationId().replaceAll("-",":"));
        request.setInVlanId(getStdInVlanId(request));
        request.setOutVlanId(getStdOutVlanId(request));
        return request;
    }

    /**
     * 解析Mikrotik私有属性
     * @param request
     * @return
     */
    private RadiusPacket filterMikrotik(RadiusPacket request){
        request.setMacAddr(request.getCallingStationId().replaceAll("-",":"));
        int[] vlans = parseCiscoVlan(request);
        request.setInVlanId(vlans[0]);
        request.setOutVlanId(vlans[1]);
        return request;
    }

    /**
     * 解析huawei私有属性
     * @param request
     * @return
     */
    private RadiusPacket filterHuawei(RadiusPacket request){
        request.setMacAddr(request.getCallingStationId().replaceAll("-",":"));
        request.setInVlanId(getStdInVlanId(request));
        request.setOutVlanId(getStdOutVlanId(request));
        return request;
    }

    /**
     * 解析H3C私有属性
     * @param request
     * @return
     */
    private RadiusPacket filterH3c(RadiusPacket request){
        String ipHostAddr = null;
        try {
            ipHostAddr = request.getAttribute("H3C-Ip-Host-Addr ").getStringValue();
        }catch (Exception e){
            logger.error("H3C MacAddr 解析失败", Memarylogger.RADIUSD);
        }

        if(ValidateUtil.isNotEmpty(ipHostAddr)){
            if (ipHostAddr.length() > 17)
                request.setMacAddr(ipHostAddr.substring(ipHostAddr.length() - 17));
            else
                request.setMacAddr(ipHostAddr);
        }
        request.setInVlanId(getStdInVlanId(request));
        request.setOutVlanId(getStdOutVlanId(request));
        return request;
    }

    private RadiusPacket filterZTE(RadiusPacket request){
        String callingStationId31 = request.getCallingStationId();
        if (callingStationId31 != null && callingStationId31.length()>12)
        {
            String mac = callingStationId31.substring(callingStationId31.length()-12);
            String macAddress = mac.substring(0, 2) + ":" + mac.substring(2, 4) + ":"
                    + mac.substring(4, 6) + ":" + mac.substring(6, 8) + ":"
                    + mac.substring(8, 10) + ":" + mac.substring(10);
            request.setMacAddr(macAddress);
        }
        int[] vlans = parseCiscoVlan(request);
        request.setInVlanId(vlans[0]);
        request.setOutVlanId(vlans[1]);
        return request;
    }

    private RadiusPacket filterRadback(RadiusPacket request){
        String macAddr = null;
        try {
            macAddr = request.getAttribute("Mac-Addr").getStringValue();
            request.setMacAddr(macAddr.replaceAll("-",":"));
        }catch (Exception e){
            logger.error("Radback MacAddr 解析失败", Memarylogger.RADIUSD);
        }
        int[] vlans = parseCiscoVlan(request);
        request.setInVlanId(vlans[0]);
        request.setOutVlanId(vlans[1]);
        return request;
    }






    /**
     * 解析标准Vlan
     * @param request
     * @return
     */
    private int getStdInVlanId(RadiusPacket request)
    {
        String nasPortId87 = request.getNasPortId();
        if(ValidateUtil.isEmpty(nasPortId87))
            return 0;

        String str = nasPortId87.toLowerCase();
        int ind1 = str.indexOf("vlanid=");
        if (ind1 == -1)
            return 0;

        String vlanId = "0";
        int ind2 = str.indexOf(";", ind1);
        if (ind2 == -1)
            vlanId = str.substring(ind1+7);
        else
            vlanId = str.substring(ind1+7, ind2);
        return Integer.parseInt(vlanId);
    }

    /**
     * 解析标准Vlan
     * @param request
     * @return
     */
    private int getStdOutVlanId(RadiusPacket request)
    {
        String nasPortId87 = request.getNasPortId();
        if(ValidateUtil.isEmpty(nasPortId87))
            return 0;

        String str = nasPortId87.toLowerCase();
        int ind1 = str.indexOf("vlanid2=");
        if (ind1 == -1)
            return 0;

        String vlanId2 = "0";
        int ind2 = str.indexOf(";", ind1);
        if (ind2 == -1)
            vlanId2 = str.substring(ind1+8);
        else
            vlanId2 = str.substring(ind1+8, ind2);
        return Integer.parseInt(vlanId2);
    }


    private int[] parseCiscoVlan(RadiusPacket requset){
        int [] result = new int[]{0,0};
        try{
            Matcher m = pattern.matcher(requset.getNasPortId());
            if(m.find()){
                result[0] = Integer.valueOf(m.group(2));
                result[1] = Integer.valueOf(m.group(1));
            }
        } catch(Exception e){
            logger.error(String.format("VLAN 解析失败 nasPortId=%s", requset.getNasPortId()), Memarylogger.RADIUSD);
        }
        return result;
    }


}
