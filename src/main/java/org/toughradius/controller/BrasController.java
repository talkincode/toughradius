package org.toughradius.controller;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.ResponseBody;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.PageResult;
import org.toughradius.common.RestResult;
import org.toughradius.component.BrasService;
import org.toughradius.component.ConfigService;
import org.toughradius.component.Syslogger;
import org.toughradius.entity.Bras;
import org.toughradius.entity.Config;

import java.util.*;

@Controller
public class BrasController {

    @Autowired
    protected Syslogger logger;

    @Autowired
    private BrasService brasService;

    @GetMapping(value = {"/admin/bras/query"})
    @ResponseBody
    public List<Bras> queryBras(){
        List<Bras> result = new ArrayList<Bras>();
        try{
            result = brasService.queryForList(new Bras());
            return result;
        }catch(Exception e){
            logger.error("query bras error",e, Syslogger.SYSTEM);
        }
        return result;
    }

    @PostMapping(value = {"/admin/bras/create"})
    @ResponseBody
    public RestResult addBras(Bras bras){
        try{
            if( !"0.0.0.0".equals(bras.getIpaddr()) &&brasService.selectByIPAddr(bras.getIpaddr())!=null){
                return new RestResult(1,"BRAS IP已经存在");
            }
            if(brasService.selectByidentifier(bras.getIdentifier())!=null){
                return new RestResult(1,"BRAS标识已经存在");
            }
            bras.setRemark("");
            bras.setStatus("enable");
            bras.setCreateTime(DateTimeUtil.nowTimestamp());
            brasService.insertBras(bras);
            return RestResult.SUCCESS;
        }catch(Exception e){
            logger.error("创建BRAS失败",e, Syslogger.SYSTEM);
            return new RestResult(1,"创建BRAS失败");
        }
    }

    @PostMapping(value = {"/admin/bras/update"})
    @ResponseBody
    public RestResult updateBras(Bras bras){
        try{
            if(brasService.selectById(bras.getId())==null){
                return new RestResult(1,"BRAS不存在");
            }
            brasService.updateBras(bras);
            return RestResult.SUCCESS;
        }catch(Exception e){
            logger.error("更新BRAS失败",e, Syslogger.SYSTEM);
            return new RestResult(1,"更新BRAS失败");
        }
    }

    @PostMapping(value = {"/admin/bras/delete"})
    @ResponseBody
    public RestResult deleteBras(Integer id){
        List<Bras> result = new ArrayList<Bras>();
        try{
            brasService.deleteById(id);
            return RestResult.SUCCESS;
        }catch(Exception e){
            logger.error("删除BRAS失败",e, Syslogger.SYSTEM);
            return new RestResult(1,"删除BRAS失败");
        }
    }


}
