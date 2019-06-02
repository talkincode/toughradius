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
    @GetMapping({"/api/v6/online/query","/admin/online/query"})
    public PageResult<RadiusOnline> queryOnlineHandler(@RequestParam(defaultValue = "0") int start, @RequestParam(defaultValue = "40") int count,
                                                       String nodeId, Integer invlan, Integer outVlan, String nasAddr, String nasId, String beginTime, String endTime, String keyword, String sort){
        return onlineCache.queryOnlinePage(start,count,nodeId,invlan,outVlan,nasAddr,nasId,beginTime,endTime,keyword,sort);
    }


    @GetMapping({"/api/v6/online/unlock", "/admin/online/unlock"})
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
    @GetMapping({"/api/v6/online/clear","/admin/online/clear"})
    public RestResult clearOnlineHandler( String nodeId,Integer invlan, Integer outVlan,  String nasAddr, String nasId, String beginTime, String endTime,  String keyword){
        onlineCache.clearOnlineByFilter(nodeId,invlan, outVlan,nasAddr,nasId,beginTime,endTime,keyword);
        return new RestResult(0,"success");
    }


    @GetMapping("/api/v6/online/fc")
    public RestResult forceClearOnlineHandler(String username){
        onlineCache.unlockOnlineByUser(username);
        return new RestResult(0,"success");
    }


    //一个下线
    @GetMapping({"/api/v6/online/delete","/admin/online/delete"})
    public RestResult DeleteOnlineHandler(String ids){
        for(String oid : ids.split("/admin,")){
            onlineCache.removeOnline(oid);
        }
        return new RestResult(0,"success");
    }

    @GetMapping({"/api/v6/online/query/byids","/admin/online/query/byids"})
    public List<RadiusOnline> queryOnlineByIds(String ids){
        return onlineCache.queryOnlineByIds(ids);
    }


}

