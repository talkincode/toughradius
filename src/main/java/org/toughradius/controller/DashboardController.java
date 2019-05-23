package org.toughradius.controller;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ResponseBody;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.RestResult;
import org.toughradius.common.SystemUtil;
import org.toughradius.component.RadiusAuthStat;
import org.toughradius.component.RadiusCastStat;
import org.toughradius.component.RadiusOnlineStat;
import org.toughradius.component.RadiusStat;

import java.util.Map;

/**
 * 控制面板
 */
@Controller
public class DashboardController {

    @Autowired
    private RadiusStat radiusStat;

    @Autowired
    private RadiusCastStat radiusCastStat;

    @Autowired
    private RadiusAuthStat radiusAuthStat;

    @Autowired
    private RadiusOnlineStat radiusOnlineStat;

    @GetMapping({"/api/v6/radius/stat","/admin/radius/stat"})
    @ResponseBody
    public Map queryRadiusStat(){
        return radiusStat.getData();
    }

    @GetMapping({"/api/v6/radius/caststat","/admin/radius/caststat"})
    @ResponseBody
    public Map queryRadiusCastStat(){
        return radiusCastStat.getData();
    }

    @GetMapping({"/api/v6/radius/authstat","/admin/radius/authstat"})
    @ResponseBody
    public Map queryRadiusAuthStat(){
        return radiusAuthStat.getData();
    }

    @GetMapping({"/api/v6/radius/onlinestat","/admin/radius/onlinestat"})
    @ResponseBody
    public Map queryRadiusOnlineStat(){
        return radiusOnlineStat.getData();
    }

    @GetMapping({"/api/v6/cpuuse","/admin/dashboard/cpuuse"})
    @ResponseBody
    public RestResult cpuuse(){
        return new RestResult(0,"ok", SystemUtil.getCpuUsage());
    }

    @GetMapping(value = {"/api/v6/memuse","/admin/dashboard/memuse"})
    @ResponseBody
    public RestResult memuse(){
        return new RestResult(0,"ok", SystemUtil.getMemUsage());
    }

    @GetMapping({"/api/v6/diskuse","/admin/dashboard/diskuse"})
    @ResponseBody
    public RestResult diskuse(){
        try {
            return new RestResult(0,"ok", SystemUtil.getDiskUsage());
        } catch (Exception e) {
            e.printStackTrace();
            return new RestResult(0,"ok", 0);
        }
    }

    @GetMapping({"/admin/dashboard/uptime"})
    @ResponseBody
    public String uptime(){
        return String.format("<i class='fa fa-bar-chart'></i> 应用系统运行时长 %s ", DateTimeUtil.formatSecond(SystemUtil.getUptime()/1000));
    }
}
