package org.toughradius.common;

import java.io.UnsupportedEncodingException;
import java.net.URLEncoder;
import java.util.HashMap;
import java.util.Iterator;
import java.util.Map;

public class URLUtil
{
    /**
     * URL编码
     * 
     * @param value 值
     * @param encoding 编码
     * @return 编码后的值
     */
    public static String toURLEncoding(String value, String encoding)
    {
        if (ValidateUtil.isEmpty(value))
            return value;
        
        try
        {
            return URLEncoder.encode(value, encoding);
        }
        catch (UnsupportedEncodingException e)
        {
            return value;
        }
    }

    /**
     * 生成URL MAP对象
     * @param url URL
     * @param map map
     */
    public static void toURLMap(String url, Map<String, String> map)
    {
        if(ValidateUtil.isEmpty(url) || map == null)
            return;
        
        int i0 = url.indexOf('?');
        if (i0 != -1)
            url = url.substring(i0+1);
        
        int i = url.indexOf('&');
        while (i != -1)
        {
            String keyValue = url.substring(0, i);
            int i1 = url.indexOf('=');
            if (i1 == -1)
                map.put(keyValue, null);
            else
            {
                String key = keyValue.substring(0, i1);
                String value = keyValue.substring(i1+1);
                map.put(key, value);
            }
            
            url = url.substring(i+1);
            i = url.indexOf('&');
        }
    }
    
    /**
     * 给定一个url,和key,取得value
     * 
     * @param url
     * @param key
     * @return String 给定一个url,和key,取得value
     */
    public static String getURLValue(String url,String key)
    {
        if(ValidateUtil.isEmpty(url) || ValidateUtil.isEmpty(key))
            return "";

        int p = url.indexOf("?");
        if (p != -1)
            url = url.substring(p);
        
        if (!url.startsWith("?"))
            url = "?" + url;
        
        int index = url.indexOf(key);
        if(index == -1)
            return "";
        
        if ((url.charAt(index-1) != '?' && url.charAt(index-1) != '&') 
            || index+key.length()>=url.length() || url.charAt(index+key.length()) != '=')
            return "";

        int point = index + key.length();
        int valueIndex = url.indexOf('&',point);
        if (valueIndex == -1)
            return url.substring(point+1,url.length());
        
        return url.substring(point+1,valueIndex);
    }
    
    /**
     * 给定一组key,value增加到url中
     * 
     * @param url
     * @param key
     * @param value
     * @return String 增加后的url
     */
    public static String urlAdd(String url,String key,String value)
    {
        if (ValidateUtil.isEmpty(url) || ValidateUtil.isEmpty(key))
            return url;
        
        if (ValidateUtil.isEmpty(value))
            value = "";
            
        if(url.indexOf("?")==-1)
            url += "?" + key + "=" + value;
        else
            url += "&" + key + "=" + value;
            
        return url;
    }
    
    /**
     * 给定一组key,value修改到url中,注：该方法仅修改第一个，多个不处理
     * 
     * @param url
     * @param key
     * @param value
     * @return String 修改后的url,注：该方法仅修改第一个，多个不处理
     */
    public static String urlModify(String url,String key,String value)
    {
        if (ValidateUtil.isEmpty(url) || ValidateUtil.isEmpty(key))
            return url;
        
        if (ValidateUtil.isEmpty(value))
            value = "";
        
        int index = url.indexOf(key + "=");
        if (index == -1)
            return urlAdd(url,key,value);
        
        if (index + key.length() == url.length())
            return url += value;
        
        int valueIndex = url.indexOf('&',index + key.length());
        if (valueIndex == -1)
            return url.substring(0,index + key.length() + 1) + value;

        return url.substring(0,index + key.length() + 1) + value + url.substring(valueIndex);
    }

    /**
     * 给定一组key删除到url中,注：该方法仅删除第一个，多个不处理
     * 
     * @param url
     * @param key
     * @return String 删除后的url,注：该方法仅删除第一个，多个不处理
     */
    public static String urlDelete(String url,String key)
    {
        if (ValidateUtil.isEmpty(url) || ValidateUtil.isEmpty(key))
            return url;
        
        int index = url.indexOf(key + "=");
        if (index == -1)
            return url;
        
        //如果是最后一个，则该KEY的VALUE为空,或最后一个KEY,则删除前一个连接符
        if (index + key.length() == url.length())
            return url.substring(0,index);
        
        int valueIndex = url.indexOf('&',index + key.length());
        if (valueIndex == -1)
            return url.substring(0,index - 1);
        
        //如果不是最后一个,则删除后一个连接符
        return url.substring(0,index) + url.substring(valueIndex + 1);
    }
    
    public static void main(String[] args)
    {
        String qs = "abc?fs=1&fd&adc=2&";
        HashMap<String, String> map = new HashMap<String, String>();
        toURLMap(qs, map);
        
        for (Iterator<Map.Entry<String, String>> it=map.entrySet().iterator();it.hasNext();)
        {
            Map.Entry<String, String> entry = it.next();
            System.out.print(entry.getKey());
            System.out.print(" = ");
            System.out.println(entry.getValue());
            
        }
    }
}
