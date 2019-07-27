package org.toughradius.controller;

import com.google.code.kaptcha.impl.DefaultKaptcha;
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
import org.toughradius.component.Memarylogger;
import org.toughradius.config.ApplicationConfig;
import org.toughradius.config.Constant;
import org.toughradius.entity.Config;
import org.toughradius.entity.MenuItem;
import org.toughradius.entity.SessionUser;

import javax.imageio.ImageIO;
import javax.servlet.ServletOutputStream;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import javax.servlet.http.HttpSession;
import java.awt.image.BufferedImage;
import java.io.ByteArrayOutputStream;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Controller
public class MainController implements Constant {

    @Autowired
    protected Memarylogger logger;

    @Autowired
    private ApplicationConfig appconfig;

    @Autowired
    private ConfigService configService;

    @Autowired
    DefaultKaptcha defaultKaptcha;

    /**
     * 构造界面菜单数据
     */
    public static List<MenuItem> getMenuData() {
        ArrayList<MenuItem> menuItems = new ArrayList<>();

        MenuItem dashboardItem = new MenuItem("dashboard", "dashboard", "控制面板");
        MenuItem cfgItem = new MenuItem("config", "cogs", "系统设置");
        MenuItem nasItem = new MenuItem("bras", "desktop", "接入设备");
        MenuItem userItem = new MenuItem("subscribe", "users", "用户管理");
        MenuItem onlineItem = new MenuItem("online", "user-circle", "在线查询");
        MenuItem ticketItem = new MenuItem("ticket", "table", "上网日志");
        MenuItem syslogItem = new MenuItem("syslog", "hdd-o", "系统日志");
        menuItems.add(dashboardItem);
        menuItems.add(cfgItem);
        menuItems.add(nasItem);
        menuItems.add(userItem);
        menuItems.add(onlineItem);
        menuItems.add(ticketItem);
        menuItems.add(syslogItem);
        return menuItems;
    }

    @GetMapping(value = {"/admin/login"})
    public String loginPage(){
        return "/static/login.html";
    }

    @GetMapping(value = {"/admin","/"})
    public String indexPage(){
        return "/static/index.html";
    }

    @GetMapping(value = "/admin/session")
    @ResponseBody
    public RestResult sessionHandeler(HttpSession session, HttpServletRequest request){
        SessionUser user = (SessionUser) session.getAttribute(SESSION_USER_KEY);
        if(user==null){
            return new RestResult(1, "用户未登录或登录已经过期");
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
    public RestResult loginHandler(String username, String password, String verifyCode,HttpSession session) {
        try {
            String sysUserName = configService.getStringValue(SYSTEM_MODULE,SYSTEM_USERNAME);
            String sysUserPwd = configService.getStringValue(SYSTEM_MODULE,SYSTEM_USERPWD);
            String vcode = (String) session.getAttribute(SESSION_VCODE_KEY);
            if(ValidateUtil.isEmpty(verifyCode)){
                return  new RestResult(1,"验证码不能为空");
            }
            if(!verifyCode.equals(vcode)){
                return  new RestResult(1,"验证码不正确");
            }

            if(ValidateUtil.isEmpty(sysUserName)){
                sysUserName = "admin";
                configService.updateConfig(new Config(SYSTEM_MODULE,SYSTEM_USERNAME,sysUserName,""));
            }
            if(ValidateUtil.isEmpty(sysUserPwd)){
                sysUserPwd = CoderUtil.md5Salt("root");
                configService.updateConfig(new Config(SYSTEM_MODULE,SYSTEM_USERPWD,sysUserPwd,""));
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
            logger.error("登录失败",e, Memarylogger.SYSTEM);
            return new RestResult(1,"login failure");
        }
    }


    @GetMapping("/admin/logout")
    public String LogoutHandler(HttpSession session) {
        session.invalidate();
        return "/static/login.html";
    }


    @GetMapping("/admin/verify-img.jpg")
    public void defaultKaptcha(HttpServletRequest httpServletRequest, HttpServletResponse httpServletResponse) throws Exception{
        byte[] captchaChallengeAsJpeg = null;
        ByteArrayOutputStream jpegOutputStream = new ByteArrayOutputStream();
        try {
            //生产验证码字符串并保存到session中
            String createText = defaultKaptcha.createText();
            httpServletRequest.getSession().setAttribute(SESSION_VCODE_KEY, createText);
            //使用生产的验证码字符串返回一个BufferedImage对象并转为byte写入到byte数组中
            BufferedImage challenge = defaultKaptcha.createImage(createText);
            ImageIO.write(challenge, "jpg", jpegOutputStream);
        } catch (IllegalArgumentException e) {
            httpServletResponse.sendError(HttpServletResponse.SC_NOT_FOUND);
            return;
        }

        //定义response输出类型为image/jpeg类型，使用response输出流输出图片的byte数组
        captchaChallengeAsJpeg = jpegOutputStream.toByteArray();
        httpServletResponse.setHeader("Cache-Control", "no-store");
        httpServletResponse.setHeader("Pragma", "no-cache");
        httpServletResponse.setDateHeader("Expires", 0);
        httpServletResponse.setContentType("image/jpeg");
        ServletOutputStream responseOutputStream = httpServletResponse.getOutputStream();
        responseOutputStream.write(captchaChallengeAsJpeg);
        responseOutputStream.flush();
        responseOutputStream.close();
    }


}
