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


    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
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
