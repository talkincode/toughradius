package org.toughradius.config;


import com.google.gson.Gson;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.method.HandlerMethod;
import org.springframework.web.servlet.ModelAndView;
import org.springframework.web.servlet.handler.HandlerInterceptorAdapter;
import org.toughradius.common.ValidateUtil;
import org.toughradius.common.coder.Base64;
import org.toughradius.component.LangResources;
import org.toughradius.component.Syslogger;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.lang.reflect.Method;

@Configuration
public class AccessInterceptor extends HandlerInterceptorAdapter {

    @Autowired
    protected Syslogger logger;

    @Autowired
    protected LangResources langs;

    @Autowired
    protected Gson gson;

    @Autowired
    protected ApplicationConfig appConfig;


    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
        response.setContentType("application/json;charset=UTF-8");
        if (handler instanceof HandlerMethod){
            String header = request.getHeader("Authorization");
            if(ValidateUtil.isNotEmpty(header) && !header.substring(0, 6).equals("Basic ")){
                response.getWriter().print(gson.toJson(langs.tr("未支持的验证方式",request.getHeader("Accept-Language"))));
                return false;
            }
            HandlerMethod handlerMethod = (HandlerMethod) handler;
            Method method = handlerMethod.getMethod();
            ApiAccess access = method.getAnnotation(ApiAccess.class);
            if(access!=null){
                String basicAuthEncoded = header.substring(6);
                //will contain "bob:secret"
                String basicAuthAsString = new String(new Base64().decode(basicAuthEncoded.getBytes()));
                if(!basicAuthAsString.trim().equals(String.format("%s:%s", appConfig.getApikey(),appConfig.getApisecret()))){
                    response.getWriter().print(gson.toJson(langs.tr("未授权的操作",request.getHeader("Accept-Language"))));
                    return false;
                }else{
                    return true;
                }
            }else {
                return true;
            }
        }else {
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
