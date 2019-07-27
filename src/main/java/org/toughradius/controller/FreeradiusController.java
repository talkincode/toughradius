package org.toughradius.controller;


import com.google.gson.Gson;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.PostMapping;
import org.tinyradius.util.RadiusException;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.ValidateUtil;
import org.toughradius.component.*;
import org.toughradius.entity.Bras;
import org.toughradius.entity.Subscribe;
import org.toughradius.form.FreeradiusAcctRequest;
import org.toughradius.form.FreeradiusAuthRequest;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;

import static org.toughradius.handler.RadiusConstant.VENDOR_MIKROTIK;

@Controller
public class FreeradiusController {

    @Autowired
    protected RadiusAuthStat radiusAuthStat;

    @Autowired
    protected RadiusStat radiusStat;

    @Autowired
    protected BrasService brasService;

    @Autowired
    protected SubscribeCache subscribeCache;

    @Autowired
    private ConfigService configService;

    @Autowired
    private FreeradiusService freeradiusService;

    @Autowired
    private Gson gson;

    @Autowired
    protected Memarylogger logger;

    private final static Map<String,Object> EMPTY_MAP = new HashMap<>();

    private Bras getNas(String nasip, String nasid) throws RadiusException {
        try {
            return brasService.findBras(nasip,null,nasid);
        } catch (ServiceException e) {
            throw  new RadiusException(e.getMessage());
        }
    }

    private void sendReject(HttpServletResponse response, String message) throws IOException {
        Map<String,Object> result = new HashMap<>();
        result.put("Reply-Message",message);
        response.setHeader("Content-Type","application/json;charset=UTF-8");
        response.setStatus(501);
        response.getWriter().write(gson.toJson(result));
    }

    private void sendAccept(HttpServletResponse response,Map<String,Object> result) throws IOException {
        response.setStatus(200);
        response.setHeader("Content-Type","application/json;charset=UTF-8");
        response.getWriter().write(gson.toJson(result));
    }

    private void sendAcctResp(HttpServletResponse response,Map<String,Object> result) throws IOException {
        response.setHeader("Content-Type","application/json;charset=UTF-8");
        response.setStatus(200);
        response.getWriter().write(gson.toJson(result));
    }

    /**
     * 默认授权
     * @param user
     * @param nas
     */
    private void filterDefault(Map<String,Object> result, Subscribe user, Bras nas){
        int sessionTimeout  = DateTimeUtil.compareSecond(user.getExpireTime(),new Date());
        Integer interimTimes = configService.getInterimTimes();
        if(interimTimes==null){
            interimTimes = 120;
        }

        if(ValidateUtil.isNotEmpty(user.getAddrPool())){
            result.put("reply:Framed-Pool",user.getAddrPool());
        }

        String ipaddr = user.getIpAddr();
        if(ValidateUtil.isNotEmpty(ipaddr) && ValidateUtil.isIP(ipaddr)){
            result.put("reply:Framed-IP-Address", ipaddr);
        }
        result.put("reply:Session-Timeout", String.valueOf(sessionTimeout));
        result.put("reply:Acct-Interim-Interval", String.valueOf(interimTimes));
    }

    /**
     * Mikrotik 授权
     * @param result
     * @param user
     */
    private void filterMikrotik(Map<String,Object> result, Subscribe user){
        long up = user.getUpRate() * 1024;
        long down = user.getDownRate() * 1024;
        result.put("reply:Mikrotik-Rate-Limit", String.format("%sk/%sk", up,down));
    }

    public void doFilter(Map<String,Object> result, Bras nas, Subscribe user){
        filterDefault(result,user, nas);
        if (VENDOR_MIKROTIK.equals(nas.getVendorId())) {
            filterMikrotik(result, user);
        }
    }


    /**
     * 验证处理，如果用户存在，就把密码响应回去做下一步验证
     *
     *    Authorize/Authenticate
     *
     *    Code   Meaning       Process body  Module code
     *    404    not found     no            notfound
     *    410    gone          no            notfound
     *    403    forbidden     no            userlock
     *    401    unauthorized  yes           reject
     *    204    no content    no            ok
     *    2xx    successful    yes           ok/updated
     *    5xx    server error  no            fail
     *    xxx    -             no            invalid
     * @param request
     * @param response
     * @throws Exception
     */
    @PostMapping(value = {"/api/freeradius/authorize"})
    public void authorize(FreeradiusAuthRequest request, HttpServletResponse response) throws Exception {
        String username = request.getUsername();
        String nasip = request.getNasip();
        String nasid = request.getNasid();
        try {
            radiusStat.incrAuthReq();
            final Bras nas = getNas(nasip, nasid);
            if (nas == null) {
                radiusStat.incrAuthDrop();
                radiusAuthStat.update(RadiusAuthStat.DROP);
                logger.error(username,"未授权的接入设备<认证> <" + nasip + ">", Memarylogger.RADIUSD);
                sendReject(response,"Unauthorized access devices");
                return;
            }

            Subscribe user = subscribeCache.findSubscribe(username);
            if(user == null){
                radiusStat.incrAuthDrop();
                radiusAuthStat.update(RadiusAuthStat.NOT_EXIST);
                logger.error(username,"用户 " + username + " 不存在", Memarylogger.RADIUSD);
                sendReject(response,"User not exist");
                return;
            }

            long timeout = (user.getExpireTime().getTime() - new Date().getTime())/1000;
            if (timeout <= 0 ) {
                radiusStat.incrAuthDrop();
                radiusAuthStat.update(RadiusAuthStat.STATUS_ERR);
                logger.error(username,"用户 " + username + " 已经过期", Memarylogger.RADIUSD);
                sendReject(response,"User expire");
                return;
            }
            Map<String,Object> result = new HashMap<>();
            result.put("control:Cleartext-Password",user.getPassword());
            doFilter(result,nas,user);
            sendAccept(response, result);
            radiusStat.incrAuthAccept();
            radiusAuthStat.update(RadiusAuthStat.ACCEPT);
        } catch (Exception e) {
            radiusStat.incrAuthDrop();
            radiusAuthStat.update(RadiusAuthStat.OTHER_ERR);
            logger.error(username,"用户 " + username + " 认证失败",e, Memarylogger.RADIUSD);
            sendReject(response,"User authorize failure");
        }
    }

    @PostMapping(value = {"/api/freeradius/authenticate"})
    public void authenticate(HttpServletRequest request, HttpServletResponse response) throws Exception {
        sendAccept(response, EMPTY_MAP);
    }

    @PostMapping(value = {"/api/freeradius/accounting"})
    public void accounting(FreeradiusAcctRequest request, HttpServletResponse response) throws Exception {
        try {
            radiusStat.incrAcctReq();
            final Bras nas = getNas(request.getNasip(), request.getNasid());
            if (nas == null) {
                logger.error(request.getUsername(),"未授权的接入设备<认证> <" + request.getNasip() + ">", Memarylogger.RADIUSD);
                return;
            }

            Subscribe user = subscribeCache.findSubscribe(request.getUsername());
            if(user == null){
                logger.error(request.getUsername(),"用户 " + request.getUsername() + " 不存在", Memarylogger.RADIUSD);
                return;
            }
            freeradiusService.doFilter(request,nas,user);
        } catch (Exception e) {
            logger.error(request.getUsername(),"用户 " + request.getUsername() + " 记账失败",e, Memarylogger.RADIUSD);
        }finally {
            radiusStat.incrAcctResp();
            sendAcctResp(response, EMPTY_MAP);
        }

    }

}
