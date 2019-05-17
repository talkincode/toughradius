package org.toughradius.component;

import org.apache.ibatis.annotations.Param;
import org.toughradius.common.CoderUtil;
import org.toughradius.entity.Config;
import org.toughradius.mapper.ConfigMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class ConfigService {

    public final static String RADIUS_MODULE = "radius";
    public final static String RADIUS_IGNORE_PASSWORD = "RADIUS_IGNORE_PASSWORD";

    public final static String SYSTEM_MODULE = "system";
    public final static String SYSTEM_USERNAME = "SYSTEM_USERNAME";
    public final static String SYSTEM_USERPWD = "SYSTEM_USERPWD";

    public final static String SMS_MODULE = "sms";
    public final static String SMS_GATEWAY = "SMS_GATEWAY";
    public final static String SMS_APPID = "SMS_APPID";
    public final static String SMS_APPKEY = "SMS_APPKEY";

    public final static String WLAN_MODULE = "wlan";
    public final static String WLAN_WECHAT_SSID = "wlan_wechat_ssid";
    public final static String WLAN_WECHAT_SHOPID = "wlan_wechat_shopid";
    public final static String WLAN_WECHAT_APPID = "wlan_wechat_appid";
    public final static String WLAN_WECHAT_SECRETKEY = "wlan_wechat_secretkey";




    @Autowired
    private ConfigMapper configMapper;

    public Config findConfig(String module, String name){
        return configMapper.findConfig(module,name);
    }

    public String getStringValue(String module, String name){
        Config cfg = configMapper.findConfig(module,name);
        if(cfg!=null){
            return cfg.getValue();
        }
        return null;
    }

    public Integer getInterimTimes(){
        return configMapper.getInterimTimes();
    }

    public Integer getIsCheckPwd(){
        return configMapper.getIsCheckPwd();
    }

    public void insertConfig(Config config){
        configMapper.insertConfig(config);
    }

    public void updateConfig(Config config){
        Config cfg = configMapper.findConfig(config.getType(),config.getName());
        if(cfg==null){
            config.setId(CoderUtil.randomLongId());
            configMapper.insertConfig(config);
        }else{
            configMapper.updateConfig(config);
        }
    }

    public void deleteById(Integer id){
        configMapper.deleteById(id);
    }

    public void deleteConfig(@Param(value = "type") String type, @Param(value = "name") String name){
        configMapper.deleteConfig(type,name);
    }

    public List<Config> queryForList(@Param(value = "type") String type){
        return configMapper.queryForList(type);
    }

}
