package org.toughradius.component;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.toughradius.entity.Option;
import org.toughradius.mapper.OptionMapper;

@Service
public class OptionService {

    public final static String RADIUS_IGNORE_PASSWORD = "RADIUS_IGNORE_PASSWORD";
    public final static String RADIUS_INTERIM_INTELVAL = "RADIUS_INTERIM_INTELVAL";

    @Autowired
    private OptionMapper optionMapper;

    public Option findOption(String name){
        return optionMapper.findOption(name);
    }

    public String getStringValue(String name){
        return optionMapper.getStringValue(name);
    }

    public Integer getIntegerValue(String name){
        return optionMapper.getIntegerValue(name);
    }

    public Integer getInterimTimes(){
        return optionMapper.getIntegerValue(RADIUS_INTERIM_INTELVAL);
    }

    public Integer getIsCheckPwd(){
        return optionMapper.getIntegerValue(RADIUS_IGNORE_PASSWORD);
    }

}
