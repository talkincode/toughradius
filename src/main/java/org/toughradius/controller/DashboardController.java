package org.toughradius.controller;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ResponseBody;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.RestResult;
import org.toughradius.common.SystemUtil;

@Controller
public class DashboardController {

    @GetMapping(value = {"/admin/dashboard/cpuuse"})
    @ResponseBody
    public RestResult cpuuse(){
        return new RestResult(0,"ok", SystemUtil.getCpuUsage());
    }

    @GetMapping(value = {"/admin/dashboard/memuse"})
    @ResponseBody
    public RestResult memuse(){
        return new RestResult(0,"ok", SystemUtil.getMemUsage());
    }

    @GetMapping(value = {"/admin/dashboard/diskuse"})
    @ResponseBody
    public RestResult diskuse(){
        try {
            return new RestResult(0,"ok", SystemUtil.getDiskUsage());
        } catch (Exception e) {
            e.printStackTrace();
            return new RestResult(0,"ok", 0);
        }
    }

    @GetMapping(value = {"/admin/dashboard/uptime"})
    @ResponseBody
    public String uptime(){
        return String.format("<i class='fa fa-bar-chart'></i> 应用系统运行时长 %s ", DateTimeUtil.formatSecond(SystemUtil.getUptime()/1000));
    }
}
