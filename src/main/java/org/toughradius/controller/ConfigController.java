package org.toughradius.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.ResponseBody;
import org.toughradius.common.CoderUtil;
import org.toughradius.common.RestResult;
import org.toughradius.common.ValidateUtil;
import org.toughradius.component.ConfigService;
import org.toughradius.component.Memarylogger;
import org.toughradius.config.Constant;
import org.toughradius.entity.Config;
import org.toughradius.form.ApiConfigForm;
import org.toughradius.form.RadiusConfigForm;
import org.toughradius.form.SmsConfigForm;
import org.toughradius.form.WlanCongigForm;

import java.util.HashMap;
import java.util.List;
import java.util.Map;


@Controller
public class ConfigController implements Constant {

    @Autowired
    protected Memarylogger logger;

    @Autowired
    private ConfigService configService;

    @GetMapping(value = {"/api/v6/config/load/{module}","/admin/config/load/{module}"})
    @ResponseBody
    public Map loadRadiusConfig(@PathVariable(name = "module")String module){
        Map result = new HashMap();
        try{
            List<Config> cfgs = configService.queryForList(module);
            for (Config cfg : cfgs){
                result.put(cfg.getName(),cfg.getValue());
            }
        }catch(Exception e){
            logger.error("query config error",e, Memarylogger.SYSTEM);
        }
        return result;
    }

    /**
     * RADIUS 配置更新
     * @param form
     * @return
     */
    @PostMapping(value = {"/api/v6/radius/update","/admin/config/radius/update"})
    @ResponseBody
    public RestResult updateRadiusConfig(RadiusConfigForm form){
        try{
            configService.updateConfig(new Config(RADIUS_MODULE,RADIUS_INTERIM_INTELVAL,form.getRadiusInterimIntelval()));
            configService.updateConfig(new Config(RADIUS_MODULE,RADIUS_TICKET_HISTORY_DAYS,form.getRadiusTicketHistoryDays()));
            configService.updateConfig(new Config(RADIUS_MODULE,RADIUS_IGNORE_PASSWORD,form.getRadiusIgnorePassword()));
            configService.updateConfig(new Config(RADIUS_MODULE,RADIUS_EXPORE_ADDR_POOL,form.getRadiusExpireAddrPool()));
        }catch(Exception e){
            logger.error("update config error",e, Memarylogger.SYSTEM);
        }
        return new RestResult(0,"update radius config done");
    }

    /**
     * 短信配置更新呢
     * @param form
     * @return
     */
    @PostMapping(value = {"/api/v6/sms/update","/admin/config/sms/update"})
    @ResponseBody
    public RestResult updateSmsConfig(SmsConfigForm form){
        try{
            configService.updateConfig(new Config(SMS_MODULE,SMS_GATEWAY,form.getSmsGateway()));
            configService.updateConfig(new Config(SMS_MODULE,SMS_APPID,form.getSmsAppid()));
            configService.updateConfig(new Config(SMS_MODULE,SMS_APPKEY,form.getSmsAppkey()));
            configService.updateConfig(new Config(SMS_MODULE,SMS_VCODE_TEMPLATE,form.getSmsVcodeTemplate()));
        }catch(Exception e){
            logger.error("update config error",e, Memarylogger.SYSTEM);
        }
        return new RestResult(0,"update sms config done");
    }

    /**
     * API 配置更新呢
     * @param form
     * @return
     */
    @PostMapping(value = {"/admin/config/api/update"})
    @ResponseBody
    public RestResult updateApiConfig(ApiConfigForm form){
        try{
            configService.updateConfig(new Config(API_MODULE,API_TYPE,form.getApiType()));
            configService.updateConfig(new Config(API_MODULE,API_USERNAME,form.getApiUsername()));
            configService.updateConfig(new Config(API_MODULE,API_PASSWD,form.getApiPasswd()));
            configService.updateConfig(new Config(API_MODULE,API_ALLOW_IPLIST,form.getApiAllowIplist()));
            configService.updateConfig(new Config(API_MODULE,API_BLACK_IPLIST,form.getApiBlackIplist()));
        }catch(Exception e){
            logger.error("update config error",e, Memarylogger.SYSTEM);
        }
        return new RestResult(0,"update api config done");
    }


    /**
     * 无线认证配置更新
     * @param form
     * @return
     */
    @PostMapping(value = {"/api/v6/wlan/update","/admin/config/wlan/update"})
    @ResponseBody
    public RestResult updateWlanConfig(WlanCongigForm form){
        try{
            configService.updateConfig(new Config(WLAN_MODULE,WLAN_WECHAT_SSID,form.getWlanWechatSsid()));
            configService.updateConfig(new Config(WLAN_MODULE,WLAN_WECHAT_SHOPID,form.getWlanWechatShopid()));
            configService.updateConfig(new Config(WLAN_MODULE,WLAN_WECHAT_APPID,form.getWlanWechatAppid()));
            configService.updateConfig(new Config(WLAN_MODULE,WLAN_WECHAT_SECRETKEY,form.getWlanWechatSecretkey()));
            configService.updateConfig(new Config(WLAN_MODULE,WLAN_TEMPLATE,form.getWlanTemplate()));
            configService.updateConfig(new Config(WLAN_MODULE,WLAN_JOIN_URL,form.getWlanJoinUrl()));
            configService.updateConfig(new Config(WLAN_MODULE,WLAN_RESULT_URL,form.getWlanResultUrl()));
            configService.updateConfig(new Config(WLAN_MODULE,WLAN_USERAUTH_ENABLED,form.getWlanUserauthEnabled()));
            configService.updateConfig(new Config(WLAN_MODULE,WLAN_PWDAUTH_ENABLED,form.getWlanPwdauthEnabled()));
            configService.updateConfig(new Config(WLAN_MODULE,WLAN_SMSAUTH_ENABLED,form.getWlanSmsauthEnabled()));
            configService.updateConfig(new Config(WLAN_MODULE,WLAN_WXAUTH_ENABLED,form.getWlanWxauthEnabled()));
        }catch(Exception e){
            logger.error("update config error",e, Memarylogger.SYSTEM);
        }
        return new RestResult(0,"update sms config done");
    }


    /**
     * 管理密码更新
     * @param oldpassword
     * @param password1
     * @param password2
     * @return
     */
    @PostMapping(value = {"/admin/password"})
    @ResponseBody
    public RestResult updatePasswordConfig(String oldpassword,String password1,String password2 ){
        if(ValidateUtil.isEmpty(password1)||password1.length() < 6){
            return new RestResult(1, String.format("密码长度至少%s位", 6));
        }

        if(!password1.equals(password2)){
            return new RestResult(1,"确认密码不符");
        }

        String sysUserPwd = configService.getStringValue(SYSTEM_MODULE,SYSTEM_USERPWD);

        if(!sysUserPwd.equals(CoderUtil.md5Salt(oldpassword))){
            return new RestResult(1,"旧密码错误");
        }

        configService.updateConfig(new Config(SYSTEM_MODULE,SYSTEM_USERPWD,CoderUtil.md5Salt(password1),""));

        return new RestResult(0,"密码修改成功");
    }

}

