package org.toughradius.controller;

import org.toughradius.common.PageResult;
import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.TraceMessage;
import org.toughradius.component.Memarylogger;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class SyslogController {

    @Autowired
    private Memarylogger logger;

    @GetMapping("/admin/syslog/query")
    public PageResult<TraceMessage> queryTraceMessage(@RequestParam(defaultValue = "0") int start, @RequestParam(defaultValue = "40") int count,
                                                      String startDate, String endDate, String type, String username, String keyword){
        if(ValidateUtil.isNotEmpty(startDate)&&startDate.length() == 16){
            startDate += ":00";
        }
        if(ValidateUtil.isNotEmpty(endDate)&&endDate.length() == 16){
            endDate += ":59";
        }
        return logger.queryMessage(start,count,startDate,endDate,type, username,keyword);
    }

}
