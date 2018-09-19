package org.toughradius.controller;


import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.ResponseBody;
import org.toughradius.entity.OprSession;
import org.toughradius.entity.RestResult;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpSession;
import java.util.Date;

import static org.toughradius.entity.RestResult.SUCCESS;

@Controller
@Component
public class MainController extends BasicController{

    private Log logger = LogFactory.getLog(MainController.class);

    private String getLoginTemplate(){
        return "/static/login.html?v="+System.currentTimeMillis();
    }


    @GetMapping({"/admin","/"})
    public String showMain(HttpSession session){
        OprSession oss = (OprSession) session.getAttribute(OPR_SESSION_KEY);
        if(oss==null){
            return getLoginTemplate();
        }else {
            return "/static/index.html";
        }
    }


    @GetMapping(value={"/admin/login"})
    public String index() {
        return getLoginTemplate();
    }


    @PostMapping("/admin/login")
    @ResponseBody
    public RestResult loginHandler(String username, String password, HttpSession session, HttpServletRequest request, String token) {
        try {
            OprSession oprSession = getOprSession("admin", request.getRemoteAddr());
            oprSession.setLoginIp(request.getRemoteAddr());
            session.setAttribute(OPR_SESSION_KEY,oprSession);
            return RestResult.LoginSuccess;
        } catch (Exception e) {
            logger.error("用户登录错误",e);
            return RestResult.LoginError;
        }
    }

    @GetMapping("/admin/logout")
    public String LogoutHandler(HttpSession session){
        session.invalidate();
        return getLoginTemplate();
    }


    @GetMapping(value = "/admin/session")
    @ResponseBody
    public RestResult sessionHandeler(HttpSession session)
    {
        OprSession oss = (OprSession) session.getAttribute(OPR_SESSION_KEY);
        return new RestResult(SUCCESS,"OK",oss);
    }




}
