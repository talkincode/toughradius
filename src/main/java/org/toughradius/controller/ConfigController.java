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
import org.toughradius.entity.Config;
import org.toughradius.form.RadiusConfigForm;
import org.toughradius.form.SmsConfigForm;
import org.toughradius.form.WlanCongigForm;

import java.util.HashMap;
import java.util.List;
import java.util.Map;


@Controller
public class ConfigController {

    @Autowired
    protected Memarylogger logger;

    @Autowired
    private ConfigService configService;

    @GetMapping(value = {"/admin/config/load/{module}"})
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

    @PostMapping(value = {"/admin/config/radius/update"})
    @ResponseBody
    public RestResult updateRadiusConfig(RadiusConfigForm form){
        try{
            configService.updateConfig(new Config(ConfigService.RADIUS_MODULE,"RADIUS_INTERIM_INTELVAL",form.getRADIUS_INTERIM_INTELVAL()));
            configService.updateConfig(new Config(ConfigService.RADIUS_MODULE,"RADIUS_MAX_SESSION_TIMEOUT",form.getRADIUS_MAX_SESSION_TIMEOUT()));
            configService.updateConfig(new Config(ConfigService.RADIUS_MODULE,"RADIUS_TICKET_HISTORY_DAYS",form.getRADIUS_TICKET_HISTORY_DAYS()));
            configService.updateConfig(new Config(ConfigService.RADIUS_MODULE,"RADIUS_IGNORE_PASSWORD",form.getRADIUS_IGNORE_PASSWORD()));
            configService.updateConfig(new Config(ConfigService.RADIUS_MODULE,"RADIUS_EXPORE_ADDR_POOL",form.getRADIUS_EXPORE_ADDR_POOL()));
            configService.updateConfig(new Config(ConfigService.RADIUS_MODULE,"RADIUS_ONLINE_EXPIRE_CHECK",form.getRADIUS_ONLINE_EXPIRE_CHECK()));
        }catch(Exception e){
            logger.error("update config error",e, Memarylogger.SYSTEM);
        }
        return new RestResult(0,"update radius config done");
    }

    @PostMapping(value = {"/admin/config/sms/update"})
    @ResponseBody
    public RestResult updateSmsConfig(SmsConfigForm form){
        try{
            configService.updateConfig(new Config(ConfigService.SMS_MODULE,"SMS_GATEWAY",form.getSMS_GATEWAY()));
            configService.updateConfig(new Config(ConfigService.SMS_MODULE,"SMS_APPID",form.getSMS_APPID()));
            configService.updateConfig(new Config(ConfigService.SMS_MODULE,"SMS_APPKEY",form.getSMS_APPKEY()));
        }catch(Exception e){
            logger.error("update config error",e, Memarylogger.SYSTEM);
        }
        return new RestResult(0,"update sms config done");
    }

    @PostMapping(value = {"/admin/config/wlan/update"})
    @ResponseBody
    public RestResult updateWlanConfig(WlanCongigForm form){
        try{
            configService.updateConfig(new Config(ConfigService.WLAN_MODULE,"WLAN_WECHAT_SSID",form.getWLAN_WECHAT_SSID()));
            configService.updateConfig(new Config(ConfigService.WLAN_MODULE,"WLAN_WECHAT_SHOPID",form.getWLAN_WECHAT_SHOPID()));
            configService.updateConfig(new Config(ConfigService.WLAN_MODULE,"WLAN_WECHAT_APPID",form.getWLAN_WECHAT_APPID()));
            configService.updateConfig(new Config(ConfigService.WLAN_MODULE,"WLAN_WECHAT_SECRETKEY",form.getWLAN_WECHAT_SECRETKEY()));
        }catch(Exception e){
            logger.error("update config error",e, Memarylogger.SYSTEM);
        }
        return new RestResult(0,"update sms config done");
    }


    @PostMapping(value = {"/admin/password"})
    @ResponseBody
    public RestResult updatePasswordConfig(String oldpassword,String password1,String password2 ){
        if(ValidateUtil.isEmpty(password1)||password1.length() < 6){
            return new RestResult(1, String.format("密码长度至少%s位", 6));
        }

        if(!password1.equals(password2)){
            return new RestResult(1,"确认密码不符");
        }

        String sysUserPwd = configService.getStringValue(ConfigService.SYSTEM_MODULE,ConfigService.SYSTEM_USERPWD);

        if(!sysUserPwd.equals(CoderUtil.md5Salt(oldpassword))){
            return new RestResult(1,"旧密码错误");
        }

        configService.updateConfig(new Config(ConfigService.SYSTEM_MODULE,ConfigService.SYSTEM_USERPWD,CoderUtil.md5Salt(password1),""));

        return new RestResult(0,"密码修改成功");
    }

}

