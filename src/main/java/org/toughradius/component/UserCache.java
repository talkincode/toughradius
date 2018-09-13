package org.toughradius.component;

import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.User;

import java.math.BigInteger;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class UserCache {

    private final static ConcurrentHashMap<String,CacheObject> cacheData = new ConcurrentHashMap<String,CacheObject>();

    @Autowired
    private UserService userService;

    private final static Log logger = LogFactory.getLog(UserCache.class);

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
    public User findUser(String username){
        username = username.toLowerCase();
        String srcUsername = username.contains("@")? username.substring(0,username.indexOf("@")):null;
        if(ValidateUtil.isNotEmpty(srcUsername) && cacheData.containsKey(srcUsername)){
            CacheObject co = cacheData.get(srcUsername);
            return co.getUser();
        }

        if(cacheData.containsKey(username)){
            CacheObject co = cacheData.get(username);
            return co.getUser();
        }

        User subs = null;
        if(ValidateUtil.isNotEmpty(srcUsername)){
            subs = userService.findUser(srcUsername);
            cacheData.put(username, new CacheObject(srcUsername, subs));
        }

        if(subs==null){
            subs = userService.findUser(username);
            cacheData.put(username, new CacheObject(username, subs));
        }
        return subs;
    }

    protected void reloadSubscribe(String username){
        username = username.toLowerCase();
        User subs = userService.findUser(username);
        if(subs!=null){
            synchronized (cacheData)
            {
                if(cacheData.containsKey(username)){
                    CacheObject co = cacheData.get(username);
                    co.setUser(subs);
                }else{
                    cacheData.put(username, new CacheObject(username, subs));
                }
            }
        }
    }

    public void  updateSubscribeCache(){
        long start = System.currentTimeMillis();
        List<User> subslist = userService.findLastUpdateUser(DateTimeUtil.getPreviousDateTimeBySecondString(300));
        int count = 0;
        for(User subs : subslist){
            String username = subs.getUsername().toLowerCase();
            UserCache.CacheObject co = getCacheData().get(username);
            User cacheUser = co!=null?co.getUser():null;
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
        logger.info(String.format("update component total = %s, cast %s ms ", count, System.currentTimeMillis()-start));
    }

    public BigInteger getFlowAmount(String username){
        return userService.getFlowAmount(username);
    }

    class CacheObject {

        private String key;
        private User user;
        private long lastUpdate;

        public CacheObject(String key, User subscribe) {
            this.key = key;
            this.user = subscribe;
            this.setLastUpdate(System.currentTimeMillis());
        }

        public String getKey() {
            return key;
        }

        public void setKey(String key) {
            this.key = key;
        }

        public User getUser() {
            return user;
        }

        public void setUser(User user) {
            this.user = user;
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
