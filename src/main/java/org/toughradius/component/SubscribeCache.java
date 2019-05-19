package org.toughradius.component;
import org.toughradius.common.CoderUtil;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.Subscribe;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.sql.Timestamp;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class SubscribeCache {

    private final static ConcurrentHashMap<String,CacheObject> cacheData = new ConcurrentHashMap<String,CacheObject>();

    @Autowired
    private SubscribeService subscribeService;

    @Autowired
    private Memarylogger logger;

    public ConcurrentHashMap<String,CacheObject> getCacheData(){
        return cacheData;
    }

    public int size()
    {
        return cacheData.size();
    }

    /**
     *  获取缓存用户
     * @param username
     * @return
     */
    public Subscribe findSubscribe(String username){
        username = username.toLowerCase();
        String srcUsername = username.contains("@")? username.substring(0,username.indexOf("@")):null;
        if(ValidateUtil.isNotEmpty(srcUsername) && cacheData.containsKey(srcUsername)){
            CacheObject co = cacheData.get(srcUsername);
            return co.getSubscribe();
        }

        if(cacheData.containsKey(username)){
            CacheObject co = cacheData.get(username);
            return co.getSubscribe();
        }

        Subscribe subs = null;
        if(ValidateUtil.isNotEmpty(srcUsername)){
            subs = subscribeService.findSubscribe(srcUsername);
            cacheData.put(username, new CacheObject(srcUsername, subs));
        }

        if(subs==null){
            subs = subscribeService.findSubscribe(username);
            cacheData.put(username, new CacheObject(username, subs));
        }
        return subs;
    }

    /**
     *  是否存在缓存用户
     * @param username
     * @return
     */
    public boolean exists(String username){
        username = username.toLowerCase();
        String srcUsername = username.contains("@")? username.substring(0,username.indexOf("@")):null;
        return ValidateUtil.isNotEmpty(srcUsername) && cacheData.containsKey(srcUsername);
    }

    public void remove(String username){
        username = username.toLowerCase();
        String srcUsername = username.contains("@")? username.substring(0,username.indexOf("@")):"";
        cacheData.remove(srcUsername);
    }

    protected void reloadSubscribe(String username){
        username = username.toLowerCase();
        Subscribe subs = subscribeService.findSubscribe(username);
        if(subs!=null){
            synchronized (cacheData)
            {
                if(cacheData.containsKey(username)){
                    CacheObject co = cacheData.get(username);
                    co.setSubscribe(subs);
                }else{
                    cacheData.put(username, new CacheObject(username, subs));
                }
            }
        }
    }

    public void  updateSubscribeCache(){
        long start = System.currentTimeMillis();
        List<Subscribe> subslist = subscribeService.findLastUpdateUser(DateTimeUtil.getPreviousDateTimeBySecondString(300));
        int count = 0;
        for(Subscribe subs : subslist){
            String username = subs.getSubscriber().toLowerCase();
            SubscribeCache.CacheObject co = getCacheData().get(username);
            Subscribe cacheUser = co!=null?co.getSubscribe():null;
            if(cacheUser!=null && DateTimeUtil.compareSecond(cacheUser.getUpdateTime(), subs.getUpdateTime()) == 0 ){
                continue;
            }
            count ++;
            reloadSubscribe(username);
            if(count % 1000 == 0){
                try {
                    Thread.sleep(10);
                } catch (InterruptedException ignored) {
                }
            }
        }
        logger.print(String.format("update user total = %s, cast %s ms ", count, System.currentTimeMillis()-start));
    }

    public Subscribe createTempSubscribe(String username, String password, int days){
        Subscribe subscribe = new Subscribe();
        subscribe.setId(CoderUtil.randomLongId());
        subscribe.setSubscriber(username);
        subscribe.setPassword(password);
        subscribe.setNodeId(0L);
        subscribe.setRealname("临时用户");
        subscribe.setDomain("wlan");
        subscribe.setAddrPool("");
        subscribe.setPolicy("");
        subscribe.setIsOnline(0);
        subscribe.setActiveNum(1);
        subscribe.setBindMac(0);
        subscribe.setBindVlan(0);
        subscribe.setIpAddr("");
        subscribe.setMacAddr("");
        subscribe.setInVlan(0);
        subscribe.setOutVlan(0);
        subscribe.setUpRate(10485760L);
        subscribe.setDownRate(10485760L);
        subscribe.setUpRateCode("");
        subscribe.setDownRateCode("");
        subscribe.setStatus("enabled");
        subscribe.setBeginTime(DateTimeUtil.nowTimestamp());
        subscribe.setExpireTime(DateTimeUtil.getDayEnd(DateTimeUtil.nowTimestamp(),days));
        subscribe.setCreateTime(DateTimeUtil.nowTimestamp());
        subscribe.setUpdateTime(DateTimeUtil.nowTimestamp());
        subscribe.setRemark("临时用户");
        return subscribe;
    }


    class CacheObject {

        private String key;
        private Subscribe subscribe;
        private long lastUpdate;

        public CacheObject(String key, Subscribe subscribe) {
            this.key = key;
            this.subscribe = subscribe;
            this.setLastUpdate(System.currentTimeMillis());
        }

        public String getKey() {
            return key;
        }

        public void setKey(String key) {
            this.key = key;
        }

        public Subscribe getSubscribe() {
            return subscribe;
        }

        public void setSubscribe(Subscribe subscribe) {
            this.subscribe = subscribe;
            this.setLastUpdate(System.currentTimeMillis());
        }

        public long getLastUpdate() {
            return lastUpdate;
        }

        public void setLastUpdate(long lastUpdate) {
            this.lastUpdate = lastUpdate;
        }
    }

}
