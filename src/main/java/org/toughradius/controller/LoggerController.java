package org.toughradius.controller;

import org.toughradius.common.PageResult;
import org.toughradius.common.RestResult;
import org.toughradius.common.ValidateUtil;
import org.toughradius.component.TicketCache;
import org.toughradius.entity.RadiusTicket;
import org.toughradius.entity.TraceMessage;
import org.toughradius.component.Syslogger;
import org.toughradius.component.ServiceException;
import org.toughradius.component.RadiusStat;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.ArrayList;
import java.util.Map;

@RestController
public class LoggerController {

    @Autowired
    private TicketCache ticketCache;

    @Autowired
    private RadiusStat radiusStat;

    @Autowired
    private Syslogger logger;

    @GetMapping("/admin/logger/query")
    public PageResult<TraceMessage> queryTraceMessage(@RequestParam(defaultValue = "0") int start,
                                                      @RequestParam(defaultValue = "40") int count,
                                                      String startDate, String endDate, String type, String username, String keyword){
        if(ValidateUtil.isNotEmpty(startDate)&&startDate.length() == 16){
            startDate += ":00";
        }
        if(ValidateUtil.isNotEmpty(endDate)&&endDate.length() == 16){
            endDate += ":59";
        }
        return logger.queryMessage(start,count,startDate,endDate,type, username,keyword);
    }

    @GetMapping("/admin/radius/stat")
    public Map queryRadiusStat(){
        return radiusStat.getData();
    }

    @PostMapping("/admin/logger/add")
    public RestResult addTraceMessage(String name, String msg, String type){
        if(ValidateUtil.isNotEmpty(msg)){
            logger.info(name, msg,type);
        }
        return RestResult.SUCCESS;
    }

    @GetMapping("/admin/ticket/query")
    public PageResult<RadiusTicket> queryTicket(@RequestParam(defaultValue = "0") int start,
                                                @RequestParam(defaultValue = "40") int count,
                                                String startDate,
                                                String endDate,
                                                String nasid,
                                                String nasaddr,
                                                Integer nodeId,
                                                Integer areaId,
                                                String username,
                                                String keyword){

        try {
            return ticketCache.queryTicket(start,count,startDate,endDate, nasid, nasaddr, nodeId, areaId, username,keyword);
        } catch (ServiceException e) {
            logger.error(String.format("查询上网日志发生错误, %s", e.getMessage()),e,Syslogger.SYSTEM);
            return new PageResult<RadiusTicket>(start,0, new ArrayList<RadiusTicket>());
        }
    }
}
