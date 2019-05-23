package org.toughradius.controller;

import com.github.pagehelper.Page;
import com.github.pagehelper.PageHelper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.*;
import org.toughradius.common.*;
import org.toughradius.component.SubscribeService;
import org.toughradius.component.Memarylogger;
import org.toughradius.entity.Subscribe;
import org.toughradius.form.SubscribeForm;
import org.toughradius.form.SubscribeQuery;

import java.util.List;

@Controller
public class SubsribeController {

    @Autowired
    protected Memarylogger logger;

    @Autowired
    protected SubscribeService subscribeService;

    @GetMapping(value = {"/api/v6/subscribe/query","/admin/subscribe/query"})
    @ResponseBody
    public PageResult<Subscribe> querySubscribe(@RequestParam(defaultValue = "0") int start,
                                                @RequestParam(defaultValue = "40") int count,
                                                String  createTime, String expireTime, String status,String subscriber, String keyword){
        if(ValidateUtil.isNotEmpty(expireTime)&&expireTime.length() == 16){
            expireTime += ":00";
        }
        if(ValidateUtil.isNotEmpty(createTime)&&createTime.length() == 16){
            createTime += ":59";
        }
        int page = start / count;
        Page<Object> objects = PageHelper.startPage(page + 1, count);
        PageResult<Subscribe> result = new PageResult<>(0,0,null);
        try{
            SubscribeQuery query = new SubscribeQuery();
            if(ValidateUtil.isNotEmpty(expireTime))
                query.setExpireTime(DateTimeUtil.toTimestamp(expireTime));
            if(ValidateUtil.isNotEmpty(createTime))
                query.setCreateTime(DateTimeUtil.toTimestamp(createTime));
            query.setStatus(status);
            query.setKeyword(keyword);
            query.setSubscriber(subscriber);
            List<Subscribe> data = subscribeService.queryForList(query);
            return new PageResult<>(start,(int) objects.getTotal(), data);

        }catch(Exception e){
            logger.error("query subscribe error",e, Memarylogger.SYSTEM);
        }
        return result;
    }

    @GetMapping(value = {"/api/v6/subscribe/detail","/admin/subscribe/detail"})
    @ResponseBody
    public RestResult<Subscribe> querySubscribeDetail(Long id){
        try{
            return new RestResult<Subscribe>(0,"ok",subscribeService.findById(id));
        }catch(Exception e){
            logger.error("查询用户详情失败",e, Memarylogger.SYSTEM);
            return new RestResult(1,"查询用户详情失败");
        }
    }

    @PostMapping(value = {"/api/v6/subscribe/create","/admin/subscribe/create"})
    @ResponseBody
    public RestResult addSubscribe(SubscribeForm form){
        try{
            if(subscribeService.findSubscribe(form.getSubscriber())!=null){
                return new RestResult(1,"用户已经存在");
            }
            Subscribe subscribe = form.getSubscribeData();
            subscribe.setBeginTime(DateTimeUtil.nowTimestamp());
            subscribe.setCreateTime(DateTimeUtil.nowTimestamp());
            subscribe.setUpdateTime(DateTimeUtil.nowTimestamp());
            subscribe.setBeginTime(DateTimeUtil.nowTimestamp());
            subscribe.setStatus("enabled");
            subscribe.setUpPeakRate(subscribe.getUpRate());
            subscribe.setDownPeakRate(subscribe.getDownPeakRate());
            subscribeService.insertSubscribe(subscribe);
            return RestResult.SUCCESS;
        }catch(Exception e){
            logger.error("创建用户失败",e, Memarylogger.SYSTEM);
            return new RestResult(1,"创建用户失败");
        }
    }


    @PostMapping(value = {"/api/v6/subscribe/batchcreate","/admin/subscribe/batchcreate"})
    @ResponseBody
    public RestResult batchAddSubscribe(SubscribeForm form){
        try{
            int width = String.valueOf(form.getOpenNum()).length();
            for(int i = 0;i<form.getOpenNum();i++){
                Subscribe subscribe = form.getSubscribeData();
                subscribe.setSubscriber(form.getUserPrefix()+String.format("%0"+width+"d",i+1));
                if(form.getRandPasswd()==1||ValidateUtil.isEmpty(form.getPassword())){
                    subscribe.setPassword(StringUtil.getRandomDigits(6));
                }
                subscribe.setBeginTime(DateTimeUtil.nowTimestamp());
                subscribe.setCreateTime(DateTimeUtil.nowTimestamp());
                subscribe.setUpdateTime(DateTimeUtil.nowTimestamp());
                subscribe.setBeginTime(DateTimeUtil.nowTimestamp());
                subscribe.setStatus("enabled");
                subscribe.setUpPeakRate(subscribe.getUpRate());
                subscribe.setDownPeakRate(subscribe.getDownPeakRate());
                subscribeService.insertSubscribe(subscribe);
            }
            return RestResult.SUCCESS;
        }catch(Exception e){
            logger.error("批量创建用户失败",e, Memarylogger.SYSTEM);
            return new RestResult(1,"批量创建用户失败");
        }
    }

    @PostMapping(value = {"/api/v6/subscribe/uppwd","/admin/subscribe/uppwd"})
    @ResponseBody
    public RestResult updateSubscribe(SubscribeForm form){
        try{
            if(subscribeService.findById(form.getId())==null){
                return new RestResult(1,"用户不存在");
            }
            if(!form.getPassword().equals(form.getCpassword())){
                return new RestResult(1,"确认密码不符");
            }
            subscribeService.updatePassword(form.getId(),form.getPassword());
            return RestResult.SUCCESS;
        }catch(Exception e){
            logger.error("更新用户失败",e, Memarylogger.SYSTEM);
            return new RestResult(1,"更新用户失败");
        }
    }

    @GetMapping(value = {"/api/v6/subscribe/release","/admin/subscribe/release"})
    @ResponseBody
    public RestResult releaseSubscribe(String ids){
        try{
            for(String id : ids.split(",")){
                subscribeService.release(id);
            }
            return RestResult.SUCCESS;
        }catch(Exception e){
            logger.error("释放用户绑定失败",e, Memarylogger.SYSTEM);
            return new RestResult(1,"释放用户绑定失败");
        }
    }

    @PostMapping(value = {"/api/v6/subscribe/update","/admin/subscribe/update"})
    @ResponseBody
    public RestResult updatePassword(SubscribeForm form){
        try{
            if(subscribeService.findById(form.getId())==null){
                return new RestResult(1,"用户不存在");
            }
            Subscribe subscribe = form.getSubscribeData();
            subscribe.setUpdateTime(DateTimeUtil.nowTimestamp());
            subscribeService.updateSubscribe(subscribe);
            return RestResult.SUCCESS;
        }catch(Exception e){
            logger.error("更新用户失败",e, Memarylogger.SYSTEM);
            return new RestResult(1,"更新用户失败");
        }
    }

    @GetMapping(value = {"/api/v6/subscribe/delete","/admin/subscribe/delete"})
    @ResponseBody
    public RestResult delete(String ids){
        try{
            for (String id : ids.split(",") ) {
                subscribeService.deleteById(Long.valueOf(id));
            }
            return RestResult.SUCCESS;
        }catch(Exception e){
            logger.error("删除用户失败",e, Memarylogger.SYSTEM);
            return new RestResult(1,"删除用户失败");
        }
    }


}
