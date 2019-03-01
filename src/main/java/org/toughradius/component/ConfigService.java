package org.toughradius.component;

import org.toughradius.entity.Config;
import org.toughradius.mapper.ConfigMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class ConfigService {

    public final static String RADIUS_MODULE = "radius";
    public final static String RADIUS_IGNORE_PASSWORD = "RADIUS_IGNORE_PASSWORD";

    @Autowired
    private ConfigMapper tcConfigMapper;

    public Config findConfig(String module, String name){
        return tcConfigMapper.findConfig(module,name);
    }

    public String getStringValue(String module, String name){
        Config cfg = tcConfigMapper.findConfig(module,name);
        if(cfg!=null){
            return cfg.getValue();
        }
        return null;
    }

    public Integer getInterimTimes(){
        return tcConfigMapper.getInterimTimes();
    }

    public Integer getIsCheckPwd(){
        return tcConfigMapper.getIsCheckPwd();
    }

}
