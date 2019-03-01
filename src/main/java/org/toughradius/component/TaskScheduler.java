package org.toughradius.component;


import com.google.gson.Gson;
import org.toughradius.common.FileUtil;
import org.toughradius.config.RadiusConfig;
import org.toughradius.entity.Config;
import org.toughradius.entity.RadiusOnline;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.scheduling.concurrent.ThreadPoolTaskExecutor;
import org.springframework.stereotype.Component;

/**
 * Radius 定时任务设计
 */
@Component
public class TaskScheduler {

    @Autowired
    private OnlineCache onlineCache;

    @Autowired
    private TicketCache ticketCache;

    @Autowired
    private SubscribeCache subscribeCache;

    @Autowired
    private SubscribeService subscribeService;

    @Autowired
    private ConfigService configService;

    @Autowired
    private RadiusConfig radiusConfig;

    @Autowired
    private RadiusStat radiusStat;

    @Autowired
    private Gson gson;

    @Autowired
    private ThreadPoolTaskExecutor taskExecutor;

    @Autowired
    private Syslogger logger;



    /**
     * 消息统计任务
     */
    @Scheduled(fixedRate = 5 * 1000, initialDelay = 5)
    public void syncStatFile(){
        taskExecutor.execute(()->{
            radiusStat.runStat();
            FileUtil.writeFile(radiusConfig.getStatfile(),gson.toJson(radiusStat.getData()));
        });
    }

    /**
     * 在线用户清理
     */
    @Scheduled(fixedRate =60 * 1000)
    public void  checkOnlineExpire(){
        Config config = configService.findConfig("radius","RADIUS_INTERIM_INTELVAL");
        if(config!=null){
            int interim_times = Integer.valueOf(config.getValue());
            onlineCache.clearOvertimeTcRadiusOnline(interim_times);
        }
    }

    /**
     * 过期用户清理
     */
    @Scheduled(fixedRate =300 * 1000)
    public void  checkOnlineUserExpire(){
        onlineCache.unlockExpireTcRadiusOnline();
    }


    /**
     * 清理余额不足的在线用户, 8小时一次
     */
    @Scheduled(fixedRate = 4 * 3600 * 1000)
    public void unlockNoAmountOnline( )
    {
        for (RadiusOnline online : onlineCache.getCacheData().values()) {
            if (online.getUnLockFlag() == RadiusOnline.AMOUNT_NOT_ENOUGH){
                onlineCache.asyncUnlockOnline(online.getAcctSessionId());
            }
        }
    }

    /**
     * 更新用户缓存
     */
    @Scheduled(fixedRate = 10 * 1000)
    public void  updateSubscribeCache(){
        taskExecutor.execute(() -> subscribeCache.updateSubscribeCache());
    }

    /**
     * 同步上网日志
     */
    @Scheduled(fixedRate = 10 * 1000)
    public void syncTcRadiusTicket() {
        taskExecutor.execute(()->ticketCache.syncData());
    }

}
