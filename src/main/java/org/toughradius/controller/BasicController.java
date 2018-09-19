package org.toughradius.controller;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.context.request.RequestContextHolder;
import org.springframework.web.context.request.ServletRequestAttributes;
import org.toughradius.component.OptionService;
import org.toughradius.config.SystemConfig;
import org.toughradius.entity.MenuItem;
import org.toughradius.entity.OprSession;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpSession;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;


public class BasicController {

    public final static String OPR_SESSION_KEY = "OPR_SESSION_KEY";

    private Log logger = LogFactory.getLog(BasicController.class);

    public OprSession getOprSession(String username, String loginIp) {
        OprSession oss =  new OprSession(username, loginIp );
        oss.setMenus(getMenus());
        return oss;
    }


    private List<MenuItem> getMenus(){
        List<MenuItem> menus = new ArrayList<MenuItem>();

        MenuItem dashboard= new MenuItem("dashboard","dashboard","控制面板");
        menus.add(dashboard);

        MenuItem option= new MenuItem("option","cogs","配置管理");
        menus.add(option);

        MenuItem nas = new MenuItem("nas","server","NAS 设备管理");
        menus.add(nas);

        MenuItem group = new MenuItem("group","users","用户组管理");
        menus.add(group);

        MenuItem user = new MenuItem("user","user","用户管理");
        menus.add(user);

        return menus;
    }
}
