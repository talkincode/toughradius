package org.toughradius.component;

import org.toughradius.common.CoderUtil;
import org.toughradius.config.RadiusConfig;
import org.toughradius.entity.Subscribe;
import org.tinyradius.packet.AccessRequest;
import org.tinyradius.packet.AccountingRequest;
import org.tinyradius.packet.RadiusPacket;
import org.tinyradius.util.RadiusClient;
import org.tinyradius.util.RadiusException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.util.HashMap;
import java.util.Random;

@Component
public class RadiusTester {

    private final static HashMap<String,String> macMap = new HashMap<String,String>();
    private final static HashMap<String,String> vlanMap = new HashMap<String,String>();
    private final static HashMap<String,String> ipmap = new HashMap<String,String>();

    @Autowired
    protected RadiusConfig radiusConfig;

    @Autowired
    private Memarylogger logger;

    @Autowired
    private BrasService brasService;

    @Autowired
    private SubscribeCache subscribeCache;

    public String randomMac() {
        Random random = new Random();
        String[] mac = {
                String.format("%02x", 0x52),
                String.format("%02x", 0x54),
                String.format("%02x", 0x00),
                String.format("%02x", random.nextInt(0xff)),
                String.format("%02x", random.nextInt(0xff)),
                String.format("%02x", random.nextInt(0xff))
        };
        return String.join(":", mac);
    }

    public String randomNasPortId() {
        Random random = new Random();
        return String.format("3/0/1:%s.%s", random.nextInt(0x1000),random.nextInt(0x1000));
    }

    public String randIpaddr(){
        Random random = new Random();
        return String.format("192.%s.%s.%s",random.nextInt(0xfe), random.nextInt(0xfe),random.nextInt(0xfe));
    }




    public String sendAuth(String username,String papchap){
        StringBuilder result = new StringBuilder();
        try {
            RadiusClient cli = new RadiusClient("127.0.0.1","secret");
            AccessRequest request = new AccessRequest();
            Subscribe user = subscribeCache.findSubscribe(username);
            if(user==null){
                return "用户不存在或已经停机";
            }

            String randmac = macMap.get(username);
            if(randmac==null){
                randmac = this.randomMac();
                macMap.put(username,randmac);
            }

            String randvlan = vlanMap.get(username);
            if(randvlan==null){
                randvlan = this.randomNasPortId();
                vlanMap.put(username,randvlan);
            }

            request.setUserName(username);
            request.setUserPassword(user.getPassword());
            request.setAuthProtocol(papchap);
            request.addAttribute("Service-Type","Framed-User");
            request.addAttribute("Framed-Protocol","PPP");
            request.addAttribute("NAS-IP-Address","127.0.0.1");
            request.addAttribute("Calling-Station-Id",randmac);
            request.addAttribute("NAS-Identifier","inner-tester");
            request.addAttribute("NAS-Port-Id",randvlan);

            result.append(String.format("发送测试认证请求 %s", request.toString())).append("\n\n");
            RadiusPacket dmrep = cli.communicate(request,radiusConfig.getAuthport());
            result.append(String.format("接收到测试认证响应 %s", dmrep.toString()));
            return result.toString();
        } catch (IOException | RadiusException e) {
            result.append(String.format("发送测试认证失败 %s", e.getMessage()));
            return result.toString();
        }
    }


    public String sendAcct(String username, int type){
        StringBuilder result = new StringBuilder();
        try {
            RadiusClient cli = new RadiusClient("127.0.0.1","secret");
            AccountingRequest request = new AccountingRequest(username,type);
            String randmac = macMap.get(username);
            if(randmac==null){
                randmac = this.randomMac();
                macMap.put(username,randmac);
            }

            String randvlan = vlanMap.get(username);
            if(randvlan==null){
                randvlan = this.randomNasPortId();
                vlanMap.put(username,randvlan);
            }

            String randIpaddr = ipmap.get(username);
            if(randIpaddr==null){
                randIpaddr = this.randIpaddr();
                ipmap.put(username,randIpaddr);
            }

            request.setUserName(username);
            request.addAttribute("Service-Type","Framed-User");
            request.addAttribute("Framed-Protocol","PPP");
            request.addAttribute("Acct-Session-Id", CoderUtil.md5Encoder(randmac));
            request.addAttribute("NAS-IP-Address","127.0.0.1");
            request.addAttribute("Calling-Station-Id",randmac);
            request.addAttribute("Called-Station-Id","00-04-5F-00-0F-D1");
            request.addAttribute("NAS-Identifier","inner-tester");
            request.addAttribute("NAS-Port-Id",randvlan);
            request.addAttribute("Framed-IP-Address",randIpaddr);
            request.addAttribute("NAS-Port","0");
            if(type == AccountingRequest.ACCT_STATUS_TYPE_START){
                request.addAttribute("Acct-Input-Octets","0");
                request.addAttribute("Acct-Output-Octets","0");
                request.addAttribute("Acct-Input-Packets","0");
                request.addAttribute("Acct-Output-Packets","0");
                request.addAttribute("Acct-Session-Time","0");
            }else if(type == AccountingRequest.ACCT_STATUS_TYPE_INTERIM_UPDATE){
                request.addAttribute("Acct-Input-Octets","1048576");
                request.addAttribute("Acct-Output-Octets","8388608");
                request.addAttribute("Acct-Input-Packets","1048576");
                request.addAttribute("Acct-Output-Packets","8388608");
                request.addAttribute("Acct-Session-Time","60");
            }else if(type == AccountingRequest.ACCT_STATUS_TYPE_STOP){
                request.addAttribute("Acct-Input-Octets","2097152");
                request.addAttribute("Acct-Output-Octets","16777216");
                request.addAttribute("Acct-Input-Packets","2097152");
                request.addAttribute("Acct-Output-Packets","16777216");
                request.addAttribute("Acct-Session-Time","120");
            }

            result.append(String.format("发送测试记账请求 %s", request.toString())).append("\n\n");
            RadiusPacket dmrep = cli.communicate(request,radiusConfig.getAcctport());
            result.append(String.format("接收到测试记账响应 %s", dmrep.toString()));
            return result.toString();
        } catch (IOException | RadiusException e) {
            result.append(String.format("发送测试记账失败 %s", e.getMessage()));
            return result.toString();
        }
    }
}
