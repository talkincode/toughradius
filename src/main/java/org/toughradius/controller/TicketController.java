package org.toughradius.controller;

import io.swagger.annotations.Api;
import io.swagger.annotations.ApiImplicitParam;
import io.swagger.annotations.ApiImplicitParams;
import io.swagger.annotations.ApiOperation;
import org.springframework.web.bind.annotation.ResponseBody;
import org.toughradius.common.PageResult;
import org.toughradius.component.TicketCache;
import org.toughradius.entity.RadiusTicket;
import org.toughradius.component.Memarylogger;
import org.toughradius.component.ServiceException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.ArrayList;

@RestController
@Api(value = "用户流量日志API")
public class TicketController {

    @Autowired
    private TicketCache ticketCache;

    @Autowired
    private Memarylogger logger;

    @GetMapping("/api/v6/ticket/query")
    @ResponseBody
    @ApiOperation(value = "用户流量日志查询API, 支持分页")
    @ApiImplicitParams({
            @ApiImplicitParam(paramType = "query", dataType = "int", name = "start", value = "分页起始记录", defaultValue = "0", required = true),
            @ApiImplicitParam(paramType = "query", dataType = "int", name = "count", value = "分页大小", defaultValue = "40", required = true),
            @ApiImplicitParam(paramType = "query", dataType = "String", name = "startDate", value = "上线时间起始点", required = false),
            @ApiImplicitParam(paramType = "query", dataType = "String", name = "endDate", value = "上线时间结束点", required = false),
            @ApiImplicitParam(paramType = "query", dataType = "String", name = "nasid", value = "NAS 标识", required = false),
            @ApiImplicitParam(paramType = "query", dataType = "String", name = "nasaddr", value = "NAS 地址", required = false),
            @ApiImplicitParam(paramType = "query", dataType = "String", name = "nodeId", value = "节点ID(扩展)", required = false),
            @ApiImplicitParam(paramType = "query", dataType = "String", name = "username", value = "用户账号", required = false),
            @ApiImplicitParam(paramType = "query", dataType = "String", name = "keyword", value = "模糊匹配", required = false)
    })
    public PageResult<RadiusTicket> queryTicketApi(@RequestParam(defaultValue = "0") int start,
                                                   @RequestParam(defaultValue = "40") int count,
                                                   String startDate,
                                                   String endDate,
                                                   String nasid,
                                                   String nasaddr,
                                                   Integer nodeId,
                                                   String username,
                                                   String keyword) {
        return queryTicket(start, count, startDate, endDate, nasid, nasaddr, nodeId, username, keyword);
    }

    @GetMapping("/admin/ticket/query")
    @ApiOperation(value = "", hidden = true)
    public PageResult<RadiusTicket> queryTicket(@RequestParam(defaultValue = "0") int start,
                                                @RequestParam(defaultValue = "40") int count,
                                                String startDate,
                                                String endDate,
                                                String nasid,
                                                String nasaddr,
                                                Integer nodeId,
                                                String username,
                                                String keyword) {

        try {
            return ticketCache.queryTicket(start, count, startDate, endDate, nasid, nasaddr, nodeId, username, keyword);
        } catch (ServiceException e) {
            logger.error(String.format("/admin查询上网日志发生错误, %s", e.getMessage()), e, Memarylogger.SYSTEM);
            return new PageResult<RadiusTicket>(start, 0, new ArrayList<RadiusTicket>());
        }
    }
}
