package org.toughradius.controller;

import org.toughradius.common.PageResult;
import org.toughradius.common.RestResult;
import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.RadiusOnline;
import org.toughradius.component.OnlineCache;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
public class OnlineController {

    @Autowired
    private OnlineCache onlineCache;

    //在线查询
    @GetMapping("/online/query")
    public PageResult<RadiusOnline> queryOnlineHandler(@RequestParam(defaultValue = "0") int start, @RequestParam(defaultValue = "40") int count,
                                                       String nodeId, String areaId, Integer invlan, Integer outVlan, String nasAddr, String nasId, String beginTime, String endTime, String keyword, String sort){
        return onlineCache.queryOnlinePage(start,count,nodeId,areaId,invlan,outVlan,nasAddr,nasId,beginTime,endTime,keyword,sort);
    }


    @GetMapping("/online/getlast")
    public RadiusOnline getLastOnlineHandler(String username){
        return onlineCache.getLastOnline(username);
    }

    @GetMapping("/online/unlock")
    public RestResult unlockOnlineHandler(@RequestParam(name = "ids")String ids,
                                          @RequestParam(name = "sessionId")String sessionId){
        if(ValidateUtil.isNotEmpty(ids)){
            onlineCache.unlockOnlines(ids);
            return new RestResult(0,"批量下线执行中，请等待");
        }else if(ValidateUtil.isNotEmpty(sessionId)){
            boolean r= onlineCache.unlockOnline(sessionId);
            return new RestResult(r?0:1,"下线执行完成");
        }else{
            return new RestResult(1,"无效参数");
        }
    }
    //清理在线
    @GetMapping("/online/clear")
    public RestResult clearOnlineHandler( String nodeId, String areaId,Integer invlan, Integer outVlan,  String nasAddr, String nasId, String beginTime, String endTime,  String keyword){
        onlineCache.clearOnlineByFilter(nodeId,areaId,invlan, outVlan,nasAddr,nasId,beginTime,endTime,keyword);
        return new RestResult(0,"success");
    }
    //一个下线
    @GetMapping("/online/delete")
    public RestResult DeleteOnlineHandler(String ids){
        for(String oid : ids.split(",")){
            onlineCache.removeOnline(oid);
        }
        return new RestResult(0,"success");
    }

    @GetMapping("/online/query/byids")
    public List<RadiusOnline> queryOnlineByIds(String ids){
        return onlineCache.queryOnlineByIds(ids);
    }

    @GetMapping("/online/query/noamount")
    public List<RadiusOnline> queryNoAmountOnline(){
        return onlineCache.queryNoAmountOnline();
    }

    //清理在线用户（超时的）
    @GetMapping("/online/autoclear")
    public void autoClearHandler(int interim_times){
        onlineCache.clearOvertimeTcRadiusOnline(interim_times);
    }


    @GetMapping("/")
    public String indexHandler(){
        return "ok";
    }
}

