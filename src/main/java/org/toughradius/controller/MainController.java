package org.toughradius.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.ResponseBody;
import org.toughradius.common.CoderUtil;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.RestResult;
import org.toughradius.common.ValidateUtil;
import org.toughradius.component.ConfigService;
import org.toughradius.component.Syslogger;
import org.toughradius.config.ApplicationConfig;
import org.toughradius.entity.Config;
import org.toughradius.entity.MenuItem;
import org.toughradius.entity.SessionUser;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpSession;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Controller
public class MainController {

    @Autowired
    protected Syslogger logger;

    @Autowired
    private ApplicationConfig appconfig;

    @Autowired
    private ConfigService configService;

    private static final String SESSION_USER_KEY = "SESSION_USER_KEY";

    public static List<MenuItem> getMenuData() {
        ArrayList<MenuItem> menuItems = new ArrayList<>();

        MenuItem dashboardItem = new MenuItem("dashboard", "dashboard", "控制面板");
        MenuItem cfgItem = new MenuItem("config", "cogs", "系统设置");
        MenuItem nasItem = new MenuItem("bras", "desktop", "NAS 管理");
        MenuItem userItem = new MenuItem("subscribe", "users", "用户管理");
        MenuItem onlineItem = new MenuItem("online", "user-circle", "在线查询");
        MenuItem ticketItem = new MenuItem("ticket", "table", "上网日志");
        menuItems.add(dashboardItem);
        menuItems.add(cfgItem);
        menuItems.add(nasItem);
        menuItems.add(userItem);
        menuItems.add(onlineItem);
        menuItems.add(ticketItem);
        return menuItems;
    }

    @GetMapping(value = {"/admin/login"})
    public String loginPage(){
        return "/static/login.html";
    }

    @GetMapping(value = {"/admin","/"})
    public String indexPage(HttpSession session){
        SessionUser user = (SessionUser) session.getAttribute(SESSION_USER_KEY);
        if(user==null){
            return "/static/login.html";
        }else {
            return "/static/index.html";
        }
    }

    @GetMapping(value = "/admin/session")
    @ResponseBody
    public RestResult sessionHandeler(HttpSession session, HttpServletRequest request){
        SessionUser user = (SessionUser) session.getAttribute(SESSION_USER_KEY);
        if(user==null){
            return new RestResult(1, "not login");
        }
        RestResult result = new RestResult(0,"ok");
        Map<String, Object> child = new HashMap<>();
        String localAddr = request.getLocalAddr();
        child.put("menudata", getMenuData());
        child.put("username", user.getUsername());
        child.put("lastLogin", user.getLastLogin());
        child.put("level", "super");
        child.put("system_name","ToughRADIUS");
        child.put("version", appconfig.getVersion());
        child.put("ipaddr", localAddr);
        result.setData(child);
        return  result;
    }



    @PostMapping("/admin/login")
    @ResponseBody
    public RestResult loginHandler(String username, String password, HttpSession session) {
        try {
            String sysUserName = configService.getStringValue(ConfigService.SYSTEM_MODULE,ConfigService.SYSTEM_USERNAME);
            String sysUserPwd = configService.getStringValue(ConfigService.SYSTEM_MODULE,ConfigService.SYSTEM_USERPWD);
            if(ValidateUtil.isEmpty(sysUserName)){
                sysUserName = "admin";
                configService.updateConfig(new Config(ConfigService.SYSTEM_MODULE,ConfigService.SYSTEM_USERNAME,sysUserName,""));
            }
            if(ValidateUtil.isEmpty(sysUserPwd)){
                sysUserPwd = CoderUtil.md5Salt("root");
                configService.updateConfig(new Config(ConfigService.SYSTEM_MODULE,ConfigService.SYSTEM_USERPWD,sysUserPwd,""));
            }

            if(username.equals(sysUserName) && CoderUtil.md5Salt(password).equals(sysUserPwd)){
                SessionUser suser = new SessionUser(sysUserName);
                suser.setLastLogin(DateTimeUtil.getDateTimeString());
                session.setAttribute(SESSION_USER_KEY, suser);
                return RestResult.SUCCESS;
            }else{
                return  new RestResult(1,"用户名密码错误");
            }
        } catch (Exception e) {
            logger.error("登录失败",e,Syslogger.SYSTEM);
            return new RestResult(1,"login failure");
        }
    }


    @GetMapping("/admin/logout")
    public String LogoutHandler(HttpSession session) {
        session.invalidate();
        return "/static/login.html";
    }


}
