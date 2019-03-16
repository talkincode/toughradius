package org.toughradius.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.ResponseBody;
import org.toughradius.common.RestResult;
import org.toughradius.common.SystemUtil;
import org.toughradius.component.ConfigService;
import org.toughradius.component.Syslogger;
import org.toughradius.entity.Config;
import org.toughradius.entity.RadiusConfigForm;

import java.util.HashMap;
import java.util.List;
import java.util.Map;


@Controller
public class ConfigController {

    @Autowired
    protected Syslogger logger;

    @Autowired
    private ConfigService configService;

    @GetMapping(value = {"/admin/config/load/radius"})
    @ResponseBody
    public Map loadRadiusConfig(){
        Map result = new HashMap();
        try{
            List<Config> cfgs = configService.queryForList(ConfigService.RADIUS_MODULE);
            for (Config cfg : cfgs){
                result.put(cfg.getName(),cfg.getValue());
            }
        }catch(Exception e){
            logger.error("query config error",e,Syslogger.SYSTEM);
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
            logger.error("update config error",e,Syslogger.SYSTEM);
        }
        return new RestResult(0,"update radius config done");
    }
}

