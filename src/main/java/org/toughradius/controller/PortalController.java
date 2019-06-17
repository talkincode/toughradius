package org.toughradius.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.ResponseBody;
import org.springframework.web.servlet.ModelAndView;
import org.toughradius.common.*;
import org.toughradius.common.coder.Encypt;
import org.toughradius.component.*;
import org.toughradius.config.Constant;
import org.toughradius.config.PortalConfig;
import org.toughradius.entity.Bras;
import org.toughradius.entity.Subscribe;
import org.toughradius.form.WlanParam;
import org.toughradius.entity.WlanSession;
import org.toughradius.portal.PortalClient;
import org.toughradius.portal.PortalException;
import org.toughradius.portal.packet.PortalPacket;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import javax.servlet.http.HttpSession;
import java.io.IOException;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;
import java.util.Random;
import java.util.concurrent.ConcurrentHashMap;


@Controller
public class PortalController implements Constant {

    private final static String MODEL_OK = "ok";
    private final static String MODEL_FAIL = "fail";
    public final static String WLAN_SESSION_KEY = "TOUGHRADIUS_WLAN_SESSION_KEY";

    private final static ConcurrentHashMap<String,WlanParam> weifiParamCache = new ConcurrentHashMap<>();
    private final static ConcurrentHashMap<String, SmsSender.SmscodeCounter> smscodeCache = new ConcurrentHashMap<>();

    private Random random = new Random();


    @Autowired
    protected SmsSender smsSender;

    @Autowired
    protected Memarylogger logger;

    @Autowired
    private BrasService brasService;

    @Autowired
    private PortalConfig portalConfig;

    @Autowired
    private PortalClient client;

    @Autowired
    private ConfigService configService;

    @Autowired
    private SubscribeCache subscribeCache;

    /**
     * 模拟 ac 强制跳转
     * @param response
     * @throws IOException
     */
    @GetMapping("/wlandemo")
    public void wlandemo(HttpServletResponse response)throws IOException {
        response.sendRedirect("/wlan/index?wlanuserip=127.0.0.1&wlanusername=test01&" +
                "wlanusermac=00:00:00:00:00:00&wlanacname=default&wlanacip=127.0.0.1&wlanapmac=00:00:00:00:00:00&" +
                "ssid=toughwifi&wlanuserfirsturl=baidu.com&error=&v="+ DateTimeUtil.getDateTimeString());
    }

    private void setConfigModel(ModelAndView modelAndView){
        modelAndView.addObject(WLAN_USERAUTH_ENABLED, configService.getStringValue(WLAN_MODULE,Constant.WLAN_USERAUTH_ENABLED));
        modelAndView.addObject(WLAN_PWDAUTH_ENABLED, configService.getStringValue(WLAN_MODULE,Constant.WLAN_PWDAUTH_ENABLED));
        modelAndView.addObject(WLAN_SMSAUTH_ENABLED, configService.getStringValue(WLAN_MODULE,Constant.WLAN_SMSAUTH_ENABLED));
        modelAndView.addObject(WLAN_WXAUTH_ENABLED, configService.getStringValue(WLAN_MODULE,Constant.WLAN_WXAUTH_ENABLED));
    }

    private String getWlanemplate(){
        String template = configService.getStringValue(WLAN_MODULE,Constant.WLAN_TEMPLATE);
        if(ValidateUtil.isEmpty(template)){
            return "default";
        }
        return template;
    }


    private void setResultUrl(WlanParam param){
        String resultUrl = configService.getStringValue(WLAN_MODULE,WLAN_RESULT_URL);
        if(ValidateUtil.isNotEmpty(resultUrl)){
            param.setWlanuserfirsturl(resultUrl);
        }
    }


    /**
     * portal 首页
     * @param wlanParam
     * @return
     */
    @GetMapping("/wlan/index")
    public ModelAndView wlanIndexHandler(WlanParam wlanParam,HttpServletRequest request){
        String joinUrl = configService.getStringValue(WLAN_MODULE,Constant.WLAN_JOIN_URL);
        String template = getWlanemplate();
        ModelAndView modelAndView = new ModelAndView(template+"/index");
        wlanParam.setTemplate(template);
        wlanParam.setSrcacip(request.getRemoteAddr());
        setResultUrl(wlanParam);
        modelAndView.addObject("param", wlanParam);
        modelAndView.addObject("joinUrl", joinUrl);
        setConfigModel(modelAndView);
        return modelAndView;
    }


    /**
     * 认证页面
     * @param wlanParam
     * @return
     */
    @GetMapping("/wlan/login")
    public ModelAndView wlanLoginHandler(WlanParam wlanParam,HttpSession session,HttpServletRequest request,HttpServletResponse response){
        String autoAuthUsername = CookieUtils.getCookie(request,wlanParam.getWlanusermac());
        if(ValidateUtil.isNotEmpty(autoAuthUsername)){
            Subscribe subs = subscribeCache.findSubscribe(autoAuthUsername);
            if(subs!=null){
                wlanParam.setUsername(subs.getSubscriber());
                return authAuthAction(session,request,response,wlanParam,subs);
            }
        }

        ModelAndView modelAndView = new ModelAndView(getWlanemplate()+"/"+wlanParam.getAuthmode());
        wlanParam.setSrcacip(request.getRemoteAddr());
        modelAndView.addObject("param", wlanParam);
        setResultUrl(wlanParam);
        return modelAndView;
    }


    /**
     * 自动认证
     * @param session
     * @param request
     * @param param
     * @return
     */
    public ModelAndView authAuthAction(HttpSession session, HttpServletRequest request, HttpServletResponse response, WlanParam param,
                                       Subscribe subscribe){
        ModelAndView modelAndView = new ModelAndView(getWlanemplate()+"/result");
        modelAndView.addObject("param",param);
        // 预处理参数
        if(ValidateUtil.isEmpty(param.getWlanuserfirsturl())){
            if(ValidateUtil.isNotEmpty(param.getUrl())){
                param.setWlanuserfirsturl(param.getUrl());
            }
        }
        setResultUrl(param);

        // 查找 AC 设备
        Bras nas = null;
        try {
            nas = brasService.findBras(param.getWlanacip(),param.getSrcacip(),param.getWlanacname());
        } catch (ServiceException e) {
            return  processModel(modelAndView,param.getUsername(), MODEL_FAIL,"接入设备不存在");
        }
        return userPwdAuth(session, request, response, param, nas, subscribe.getPassword());
    }


    private ModelAndView processModel(ModelAndView mv, String username, String code, String message){
        mv.addObject("code",code);
        mv.addObject("username",username);
        mv.addObject("message",message);
        if(code.equals(MODEL_OK)){
            logger.error(username,message,Memarylogger.PORTAL);
        }else{
            logger.info(username,message,Memarylogger.PORTAL);
        }
        return  mv;
    }

    /**
     * 认证POST提交
     * @param session
     * @param request
     * @param param
     * @param password
     * @return
     */
    @PostMapping("/wlan/login")
    public ModelAndView wlanLoginPostHandler(HttpSession session,HttpServletRequest request,HttpServletResponse response, WlanParam param,
                                             String password){
        ModelAndView modelAndView = new ModelAndView(getWlanemplate()+"/result");
        modelAndView.addObject("param",param);
        // 预处理参数
        if(ValidateUtil.isEmpty(param.getWlanuserfirsturl())){
            if(ValidateUtil.isNotEmpty(param.getUrl())){
                param.setWlanuserfirsturl(param.getUrl());
            }
        }

        setResultUrl(param);

        // 查找 AC 设备
        Bras nas = null;
        try {
            nas = brasService.findBras(param.getWlanacip(),param.getSrcacip(),param.getWlanacname());
        } catch (ServiceException e) {
            return  processModel(modelAndView,param.getUsername(), MODEL_FAIL,"接入设备不存在");
        }

        String authMode = param.getAuthmode();

        //用户密码认证
        if(PORTAL_AUTH_USERPWD.equals(authMode)){
            return userPwdAuth(session, request,response, param, nas, password);
        }

        //固定密码认证
        if(PORTAL_AUTH_PASSWORD.equals(authMode)){
            return passwordAuth(session, request,response,  param,nas, password);
        }

        //微信认证
        if(PORTAL_AUTH_WEIXIN.equals(authMode)){
            return weixinAuth(session, request,response,  param, nas);
        }

        //短信认证
        if(PORTAL_AUTH_SMS.equals(authMode)){
            return smsAuth(session, request, response, param, nas);
        }
        return  processModel(modelAndView,param.getUsername(), MODEL_FAIL,"不支持的认证模式");
    }

    /**
     * 用户名密码认证模式
     * @param session
     * @param request
     * @param param
     * @param nas
     * @param password
     * @return
     */
    private ModelAndView userPwdAuth(HttpSession session, HttpServletRequest request,HttpServletResponse response,  WlanParam param, Bras nas,
                                     String password){
        ModelAndView mv = new ModelAndView(getWlanemplate()+"/result");
        mv.addObject("param",param);
        if(ValidateUtil.isEmpty(param.getUsername())){
            return  processModel(mv,param.getUsername(), MODEL_FAIL,"帐号不能为空");
        }
        if(ValidateUtil.isEmpty(password)){
            return  processModel(mv,param.getUsername(), MODEL_FAIL,"密码不能为空");
        }
        setResultUrl(param);
        return doPortalAuth(session, request, response, mv, param, nas, param.getUsername(), password);
    }

    /**
     * 密码认证模式
     * @param session
     * @param request
     * @param param
     * @param nas
     * @param password
     * @return
     */
    private ModelAndView passwordAuth(HttpSession session,HttpServletRequest request,HttpServletResponse response, WlanParam param,Bras nas,
                                      String password){
        ModelAndView mv = new ModelAndView(getWlanemplate()+"/result");
        mv.addObject("param",param);
        if(ValidateUtil.isEmpty(password)){
            return  processModel(mv,param.getWlanusername(), MODEL_FAIL,"密码不能为空");
        }
        String username = "pu_"+ CoderUtil.random16str();
        subscribeCache.createTempSubscribe(username,password,1);
        param.setUsername(username);
        setResultUrl(param);
        return doPortalAuth(session, request, response, mv, param, nas, username, password);
    }

    /**
     * 微信认证模式
     * @param session
     * @param request
     * @param param
     * @param nas
     * @return
     */
    private ModelAndView weixinAuth(HttpSession session, HttpServletRequest request,HttpServletResponse response, WlanParam param, Bras nas){
        ModelAndView mv = new ModelAndView(getWlanemplate()+"/weixin");
        mv.addObject("param",param);
        String username = "wxu_"+ CoderUtil.random16str();
        String password = CoderUtil.random16str();
        subscribeCache.createTempSubscribe(username,password,1);
        param.setUsername(username);

        String ssid= configService.getStringValue(WLAN_MODULE,WLAN_WECHAT_SSID);
        String shopid = configService.getStringValue(WLAN_MODULE,WLAN_WECHAT_SHOPID);
        String appid = configService.getStringValue(WLAN_MODULE,WLAN_WECHAT_APPID);
        String secretKey = configService.getStringValue(WLAN_MODULE,WLAN_WECHAT_SECRETKEY);
        String extend = param.getWlanusermac();
        String timestamp = String.valueOf(DateTimeUtil.nowTimestamp().getTime());
        String authurl = "http://"+request.getHeader("host")+"/wlan/wxcallback";
        String mac = param.getWlanusermac();
        String bssid = param.getWlanapmac();
        String sign = CoderUtil.md5Encoder(appid+extend+timestamp+shopid+authurl+mac+ssid+bssid+secretKey,"UTF-8").toLowerCase();
        mv.addObject("ssid",ssid);
        mv.addObject("shopid",shopid);
        mv.addObject("appid",appid);
        mv.addObject("extend",param.getWlanusermac());
        mv.addObject("timestamp",timestamp);
        mv.addObject("authurl",authurl);
        mv.addObject("mac",param.getWlanusermac());
        mv.addObject("bssid",param.getWlanapmac());
        mv.addObject("sign",sign);
        setResultUrl(param);
        return doPortalAuth(session, request, response, mv, param, nas, username, password);
    }

    /**
     * 微信回调
     * @param request
     * @param response
     * @param openId
     * @param tid
     * @param extend
     * @throws ServiceException
     * @throws IOException
     */
    @GetMapping("/wlan/wxcallback")
    public void wechatCallback(HttpServletRequest request,HttpServletResponse response,String openId,String tid,String extend) throws ServiceException, IOException {
        logger.info("收到微信回调: "+request.getQueryString(), Memarylogger.PORTAL);
        WlanParam param =  weifiParamCache.get(extend);
        if(param==null){
            response.setStatus(500);
        }else{
            param.setRmflag(true);
            response.setStatus(200);
        }
    }

    /**
     * 短信认证模式
     * @param session
     * @param request
     * @param param
     * @param nas
     * @return
     */
    private ModelAndView smsAuth(HttpSession session,HttpServletRequest request, HttpServletResponse response,WlanParam param,Bras nas){
        ModelAndView mv = new ModelAndView(getWlanemplate()+"/result");
        mv.addObject("param",param);
        String phone = param.getPhone();
        String smscode = param.getSmscode();
        if(ValidateUtil.isEmpty(phone)){
            return processModel(mv,param.getPhone(),MODEL_FAIL,"手机号码不能为空");
        }
        if(ValidateUtil.isEmpty(smscode)){
            return processModel(mv,param.getPhone(),MODEL_FAIL,"验证码不能为空");
        }

        SmsSender.SmscodeCounter smscounter =smscodeCache.get(phone);
        if(smscounter==null){
            return processModel(mv,param.getPhone(),MODEL_FAIL,"验证码无效");
        }
        if((new Date().getTime()-smscounter.getSendtime())>300000){
            smscodeCache.remove(phone);
            return processModel(mv,param.getPhone(),MODEL_FAIL,"验证码已经过期");
        }
        if(!smscode.equals(smscounter.getSmscode())){
            return processModel(mv,param.getPhone(),MODEL_FAIL,"验证码无效");
        }

        String username = "smsu_"+ param.getPhone();
        String password = CoderUtil.random16str();
        if(!subscribeCache.exists(username)){
            subscribeCache.createTempSubscribe(username,password,1);
        }else{
            Subscribe u = subscribeCache.findSubscribe(username);
            password = u.getPassword();
        }
        param.setUsername(username);
        setResultUrl(param);
        return doPortalAuth(session, request,response, mv, param, nas, username, password);
    }

    /**
     * 发送短信
     * @param phone
     * @return
     * @throws ServiceException
     */
    @GetMapping("/wlan/sendsms")
    @ResponseBody
    public RestResult sendSms(String phone)throws ServiceException{
        if(!ValidateUtil.isPhone(phone)){
            return new RestResult(1,"无效的手机号码");
        }
        Map<String,Object> map=new HashMap<>();
        if(smscodeCache.values().stream().anyMatch(x->x.getPhone().equals(phone)&&(new Date().getTime()-x.getSendtime())<60000)){
            return new RestResult(1,"同一手机号60秒只能发送一次请求");
        }
        String vocdeTemplate = configService.getStringValue(SMS_MODULE,SMS_VCODE_TEMPLATE);
        if(ValidateUtil.isEmpty(vocdeTemplate)){
            return new RestResult(1,"没有配置短信验证码模板");
        }
        String vcode =  String.valueOf(random.nextInt((999999 - 111111) + 1) + 111111);
        int r = smsSender.sendQcloudSms(phone,vocdeTemplate.replaceAll("<vcode>",vcode));
        if(r==0){
            smscodeCache.put(phone,new SmsSender.SmscodeCounter(phone,vcode));
            logger.info(phone,"短信验证码发送成功: "+phone+" "+vcode);
            return RestResult.SUCCESS;
        }
        logger.error(phone,"短信验证码发送成失败: "+phone+" "+vcode);
        return RestResult.UNKNOW;
    }

    /**
     * AC 认证
     * @param session
     * @param request
     * @param mv
     * @param param
     * @param nas
     * @param username
     * @param password
     * @return
     */
    private ModelAndView doPortalAuth(HttpSession session,HttpServletRequest request,HttpServletResponse response,ModelAndView mv, WlanParam param,
                                      Bras nas, String username, String password){
        //认证参数
        String userIp = param.getWlanuserip();
        String secret = nas.getSecret();
        String acip = param.getWlanacip();
        String basip = nas.getIpaddr();
        int acport = nas.getAcPort();
        String mac = param.getClientMac();
        short seriaNo = -1;
        short reqId = -1;
        int ver = PortalPacket.getVerbyName(nas.getPortalVendor());

        //Challenge请求
        PortalPacket challengeResp = null;
        byte [] challenge = null;
        if(portalConfig.getPapchap() == PortalPacket.AUTH_CHAP){
            try {
                PortalPacket challengeReq = PortalPacket.createReqChallenge(ver,userIp,nas.getSecret(),param.getClientMac());
                if(portalConfig.isTraceEnabled())
                    logger.info(username,String.format(">> 发送 REQ_CHALLENGE -> : %s", basip ,challengeReq.toString()),Memarylogger.PORTAL);
                challengeResp = client.sendToAc(challengeReq,basip,acport);
                if(portalConfig.isTraceEnabled())
                    logger.info(username,String.format("<< 接收到 ACK_CHALLENGE <- %s : %s", basip,challengeResp.toString()),Memarylogger.PORTAL);
                if(challengeResp.getErrCode()>0){
                    if(challengeResp.getErrCode()==2){
                        return  processModel(mv,username, MODEL_OK,challengeResp.getErrMessage());
                    }else{
                        return  processModel(mv,username, MODEL_FAIL,challengeResp.getErrMessage());
                    }
                }
                challenge = challengeReq.getChallenge();
                seriaNo = challengeReq.getSerialNo();
                reqId = challengeResp.getReqId();
            } catch (PortalException e) {
                return  processModel(mv,username, MODEL_FAIL,"Challenge请求失败");
            }
        }

        //Auth 请求
        try {
            PortalPacket authReq = PortalPacket.createReqAuth(ver,userIp,username,password,reqId,challenge,secret,acip,portalConfig.getPapchap(),mac);
            if(portalConfig.isTraceEnabled())
                logger.info(username, String.format(">> 发送 REQ_AUTH -> %s: %s", basip,authReq.toString()),Memarylogger.PORTAL);
            PortalPacket authResp = client.sendToAc(authReq,basip,acport);
            if(portalConfig.isTraceEnabled())
                logger.info(username, String.format("<< 接收到 ACK_AUTH <- %s: %s", basip, authResp.toString()),Memarylogger.PORTAL);
            if(authResp.getErrCode()>0){
                if(authResp.getErrCode()==2){
                    authSucess(session,response,param);
                    return  processModel(mv,username, MODEL_OK,"认证成功");
                }else{
                    return  processModel(mv,username, MODEL_FAIL,authResp.getErrMessage()+authResp.getTextInfo() );
                }
            }

            seriaNo = authReq.getSerialNo();
            reqId = authResp.getReqId();
        } catch (PortalException e) {
            return  processModel(mv,username, MODEL_FAIL,"认证请求失败" );
        }

        //认证确认
        try {
            PortalPacket affackReq = PortalPacket.createAffAckAuth(ver,userIp,secret,acip,seriaNo,reqId,portalConfig.getPapchap(),mac);
            if(portalConfig.isTraceEnabled())
                logger.info(username,String.format(">> 发送 AFF_ACK_AUTH -> %s: %s",basip, affackReq.toString()),Memarylogger.PORTAL);
            client.sendToAcNoReply(affackReq,basip,acport);
        } catch (PortalException e) {
            logger.error(username,"发送 AFF_ACK_AUTH 错误",e,Memarylogger.PORTAL);
        }

        authSucess(session,response,param);
        return  processModel(mv,username, MODEL_OK,"认证成功");
    }


    /**
     * 认证结果封装
     * @param session
     * @param param
     * @return
     */
    private void authSucess(HttpSession session,HttpServletResponse response,WlanParam param){
        WlanSession wss = new WlanSession();
        wss.setWlanParam(param);
        wss.setLoginStatus(1);
        session.setAttribute(WLAN_SESSION_KEY,wss);
        if("enabled".equals(param.getRememberPwd())){
            CookieUtils.writeCookie(response,PORTAL_REMBERPWD_COOKIE, param.getUsername());
        }
    }

    /**
     * 断开连接
     * @param session
     * @param response
     * @return
     */
    @GetMapping("/wlan/disconnect")
    public ModelAndView wlanDisconnectHandler(HttpSession session,HttpServletResponse response){
        WlanSession wlanSession = (WlanSession) session.getAttribute(WLAN_SESSION_KEY);
        if(wlanSession!=null){
            WlanParam param = wlanSession.getWlanParam();
            sendLogout(param);
            session.removeAttribute(WLAN_SESSION_KEY);
            ModelAndView modelAndView = new ModelAndView(getWlanemplate()+"/index");
            modelAndView.addObject("param", param);
            setConfigModel(modelAndView);
            return modelAndView;
        }
        ModelAndView modelAndView = new ModelAndView(getWlanemplate()+"/index");
        modelAndView.addObject("param", new WlanParam());
        setConfigModel(modelAndView);
        return modelAndView;
    }

    /**
     * 向Ac发送断开请求
     * @param param
     */
    private void sendLogout(WlanParam param){
        if(param!=null){
            try {
                Bras nas = null;
                try {
                    nas = brasService.findBras(param.getWlanacip(),param.getSrcacip(),param.getWlanacname());
                } catch (ServiceException e) {
                    logger.error("bras "+param.getWlanacip()+" not exists",Memarylogger.PORTAL);
                    return;
                }

                String userIp = param.getWlanuserip();
                String secret = nas.getSecret();
                String acip = param.getWlanacip();
                String basip = nas.getIpaddr();
                int acport = nas.getAcPort();
                String mac = param.getClientMac();
                short seriaNo = (short)-1;
                int ver = PortalPacket.getVerbyName(nas.getPortalVendor());
                PortalPacket affackReq = PortalPacket.createReqLogout(ver,userIp,secret,acip,seriaNo,portalConfig.getPapchap(),mac);
                if(portalConfig.isTraceEnabled())
                    logger.info(String.format(">> 发送 REQ_LOGOUT -> %s: %s", basip, affackReq.toString()),Memarylogger.PORTAL);
                client.sendToAcNoReply(affackReq,basip,acport);
            } catch (PortalException e) {
                logger.error("发送 REQ_LOGOUT 错误",e,Memarylogger.PORTAL);
            } catch (Exception ee){
                logger.error("下线错误",ee,Memarylogger.PORTAL);
            }finally {
                if(!param.getAuthmode().equals(PORTAL_AUTH_USERPWD)){
                    subscribeCache.remove(param.getUsername());
                }
                param.setRmflag(true);
            }
        }
    }

    /**
     *微信临时认证用户认证超时处理
     */
    @Scheduled(fixedRate = 30 * 1000)
    public void checkWeChat(){
        weifiParamCache.values().removeIf(WlanParam::isRmflag);
        weifiParamCache.values().forEach(param->{
            long starttime  = param.getStarttime();
            long currtTime  = new Date().getTime();
            if(currtTime-starttime > 60 * 1000){
                sendLogout(param);
            }
        });
    }



}
