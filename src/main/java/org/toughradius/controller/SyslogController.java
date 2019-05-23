package org.toughradius.controller;

import org.toughradius.common.PageResult;
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

    @GetMapping({"/api/v6/syslog/query","/admin/syslog/query"})
    public PageResult<TraceMessage> queryTraceMessage(@RequestParam(defaultValue = "0") int start, @RequestParam(defaultValue = "40") int count,
                                                      String startDate, String endDate, String type, String username, String keyword){
        return logger.queryMessage(start,count,startDate,endDate,type, username,keyword);
    }

}
