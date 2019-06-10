package org.toughradius.common;
import org.toughradius.common.coder.Encypt;

import javax.servlet.http.Cookie;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

public class CookieUtils {

    public static String getCookie(HttpServletRequest request,String cookieName){
        Cookie[] cookies =  request.getCookies();
        String ename = Encypt.encrypt(cookieName);
        if(cookies != null){
            for(Cookie cookie : cookies){
                if(cookie.getName().equals(ename)){
                    return Encypt.decrypt(cookie.getValue());
                }
            }
        }
        return null;
    }



    public static void writeCookie(HttpServletResponse response, String cookieName,String value){
        Cookie cookie = new Cookie(Encypt.encrypt(cookieName),Encypt.encrypt(value));
        cookie.setPath("/");
        cookie.setMaxAge(86400*30);
        response.addCookie(cookie);
    }



}
