package org.toughradius.config;


import com.google.gson.Gson;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.ModelAndView;
import org.springframework.web.servlet.handler.HandlerInterceptorAdapter;
import org.toughradius.common.RestResult;
import org.toughradius.common.ValidateUtil;
import org.toughradius.common.coder.Base64;
import org.toughradius.component.ConfigService;
import org.toughradius.component.LangResources;
import org.toughradius.component.Memarylogger;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.net.InetAddress;
import java.util.Objects;

@Configuration
public class AccessInterceptor extends HandlerInterceptorAdapter implements Constant {

    @Autowired
    protected Memarylogger logger;

    @Autowired
    protected LangResources langs;

    @Autowired
    protected Gson gson;

    @Autowired
    protected ApplicationConfig appConfig;

    @Autowired
    protected ConfigService cfgService;

    private String getIpAddr(HttpServletRequest request) {
        String ip = request.getHeader("x-forwarded-for");
        if(ip == null || ip.length() == 0 || "unknown".equalsIgnoreCase(ip)) {
            ip = request.getHeader("Proxy-Client-IP");
        }
        if(ip == null || ip.length() == 0 || "unknown".equalsIgnoreCase(ip)) {
            ip = request.getHeader("WL-Proxy-Client-IP");
        }
        if(ip == null || ip.length() == 0 || "unknown".equalsIgnoreCase(ip)) {
            ip = request.getRemoteAddr();
            if(ip.equals("127.0.0.1")){
                //根据网卡取本机配置的IP
                InetAddress inet=null;
                try {
                    inet = InetAddress.getLocalHost();
                } catch (Exception e) {
                    e.printStackTrace();
                }
                ip= Objects.requireNonNull(inet).getHostAddress();
            }
        }
        // 多个代理的情况，第一个IP为客户端真实IP,多个IP按照','分割
        if(ip != null && ip.length() > 15){
            if(ip.indexOf(",")>0){
                ip = ip.substring(0,ip.indexOf(","));
            }
        }
        return ip;
    }


    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
        String ip = getIpAddr(request);

        // 白名单检测
        String allows = cfgService.getStringValue(API_MODULE,API_ALLOW_IPLIST);
        if(ValidateUtil.isNotEmpty(allows) && allows.contains(ip)){
            return true;
        }

        // 黑名单检测
        String blacks = cfgService.getStringValue(API_MODULE,API_BLACK_IPLIST);
        if(ValidateUtil.isNotEmpty(blacks) && blacks.contains(ip)){
            response.setStatus(HttpServletResponse.SC_FORBIDDEN);
            response.getWriter().print(gson.toJson(new RestResult(1,"Forbidden, black ip " + ip)));
            return false;
        }

        String header = request.getHeader("Authorization");
        response.setContentType("application/json;charset=UTF-8");
        if(ValidateUtil.isEmpty(header)){
            response.setCharacterEncoding("UTF-8");
            response.setHeader("Authorization","Required");
            response.setHeader("WWW-Authentication ","Basic");
            response.setStatus(HttpServletResponse.SC_FORBIDDEN);
            response.getWriter().print(gson.toJson(new RestResult(1,"Forbidden, unauthorized user")));
            return false;
        }
        if(ValidateUtil.isNotEmpty(header) && !header.substring(0, 6).equals("Basic ")){
            response.setStatus(HttpServletResponse.SC_FORBIDDEN);
            response.getWriter().print(gson.toJson(new RestResult(1,"Unsupported authentication methods")));
            return false;
        }
        String basicAuthEncoded = header.substring(6);
        //will contain "bob:secret"
        String basicAuthAsString = new String(new Base64().decode(basicAuthEncoded.getBytes()));
        if(!basicAuthAsString.trim().equals(String.format("%s:%s",
                cfgService.getStringValue(API_MODULE,API_USERNAME),
                cfgService.getStringValue(API_MODULE,API_PASSWD)))){
            response.getWriter().print(gson.toJson(new RestResult(1,"Authentication failure")));
            return false;
        }else{
            return true;
        }
    }

    @Override
    public void postHandle(HttpServletRequest request, HttpServletResponse response, Object handler, ModelAndView modelAndView) throws Exception {
        super.postHandle(request, response, handler, modelAndView);
    }

    @Override
    public void afterCompletion(HttpServletRequest request, HttpServletResponse response, Object handler, Exception ex) throws Exception {
        super.afterCompletion(request, response, handler, ex);
    }
}
