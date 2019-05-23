package org.toughradius.component;

import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;
import org.springframework.stereotype.Component;
import org.toughradius.common.ValidateUtil;
import org.toughradius.config.LangElement;

import javax.annotation.PostConstruct;
import java.io.*;
import java.lang.reflect.Type;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
public class LangResources {

    private Map<String,LangElement> langMap = new HashMap<String, LangElement>();

    @PostConstruct
    public void LangResources() {
        try {
            InputStream fis = LangResources.class.getClassLoader().getResourceAsStream("lang_resource.json");
            BufferedReader reader = new BufferedReader(new InputStreamReader(fis));
            Type listType = new TypeToken<ArrayList<LangElement>>(){}.getType();
            List<LangElement> langs =  new Gson().fromJson(reader,listType);
            for ( LangElement lang: langs) {
                langMap.put(lang.getSrc(),lang);
            }
        } catch (Exception ie) {
            ie.printStackTrace();
        }
    }

    public String tr(String src, String lang){
        if(ValidateUtil.isEmpty(src) || ValidateUtil.isEmpty(lang)){
            return src;
        }
        LangElement le = langMap.get(src);
        switch (lang){
            case "zh":{
                String result = le.getZh();
                if(ValidateUtil.isNotEmpty(result)){
                    return result;
                }
            }
            case "en":{
                String result = le.getEn();
                if(ValidateUtil.isNotEmpty(result)){
                    return result;
                }
            }
            default:
                return src;
        }
    }

}
