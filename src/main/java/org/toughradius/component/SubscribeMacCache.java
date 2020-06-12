package org.toughradius.component;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.Subscribe;

import java.util.concurrent.ConcurrentHashMap;

@Service
public class SubscribeMacCache {

    private final static ConcurrentHashMap<String, MacCacheObject> cacheData = new ConcurrentHashMap<String, MacCacheObject>();

    @Autowired
    private SubscribeCache subscribeCache;


    public int size()
    {
        return cacheData.size();
    }

    /**
     *  获取 MAC 缓存用户
     * @return
     */
    public Subscribe findSubscribe(String macAddr){
        macAddr = macAddr.toLowerCase();
        if(ValidateUtil.isNotEmpty(macAddr) && cacheData.containsKey(macAddr)){
            SubscribeMacCache.MacCacheObject co = cacheData.get(macAddr);
            if(co.isExpire()){
                cacheData.remove(macAddr);
                return null;
            }
            //从缓存重新
            return subscribeCache.findSubscribe(co.getUsername());
        }
        return null;
    }

    public void update(String macAddr, String username, long expire){
        macAddr = macAddr.toLowerCase();
        cacheData.put(macAddr, new MacCacheObject(macAddr,username,expire));
    }

    public void remove(String macAddr){
        macAddr = macAddr.toLowerCase();
        cacheData.remove(macAddr);
    }

    /**
     * MAC 缓存对象
     */
    class MacCacheObject {

        private String macAddr;
        private String username;
        private long lastUpdate;
        private long expire = 86400 * 30;

        public MacCacheObject(String macAddr,String username, long expire) {
            this.macAddr = macAddr;
            this.username = username;
            this.expire = expire;
            this.lastUpdate = System.currentTimeMillis();
        }


        public String getMacAddr() {
            return macAddr;
        }

        public String getUsername() {
            return username;
        }

        public boolean isExpire(){
            return (System.currentTimeMillis() - lastUpdate) > this.expire;
        }

    }
}
