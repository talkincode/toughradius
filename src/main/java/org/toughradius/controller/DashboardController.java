package org.toughradius.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.ResponseBody;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.component.RadiusStat;
import org.toughradius.entity.OprSession;
import org.toughradius.entity.RestResult;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpSession;
import java.lang.management.ManagementFactory;
import java.util.Map;

@Controller
public class DashboardController extends BasicController {

    @Autowired
    private RadiusStat radiusStat;

    @GetMapping("/admin/dashboard/uptime")
    @ResponseBody
    public String uptimeHandler(String username, String password, HttpSession session, HttpServletRequest request, String token) {
        long uptimeMs = ManagementFactory.getRuntimeMXBean().getUptime();
        return String.format("已运行时长 %s", DateTimeUtil.toTimeDesc(uptimeMs/1000));
    }


    @GetMapping("/admin/dashboard/msgstat")
    @ResponseBody
    public RestResult<Map> msgstatHandler(String username, String password, HttpSession session, HttpServletRequest request, String token) {
        return new RestResult<Map>(0,"success",radiusStat.getData());
    }


}
