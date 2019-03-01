package org.toughradius.common;

import java.util.Iterator;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;


public class ValidateCache {

    private ConcurrentHashMap<String,Counter> cacheData = new ConcurrentHashMap<String,Counter>();
    private int timems = 60 * 1000;
    private int maxTimes = 5;

    public int getMaxTimes() {
        return maxTimes;
    }

    public void setMaxTimes(int maxTimes) {
        this.maxTimes = maxTimes;
    }

    public ValidateCache(int timems, int maxTimes) {
        this.timems = timems;
        this.maxTimes = maxTimes;
    }

    public int size(){
        return cacheData.size();
    }

    public void incr(String key){
        if(cacheData.containsKey(key)){
            cacheData.get(key).incr();
        }else{
            cacheData.put(key, new Counter());
        }
    }

    public int errors(String key){
        if(cacheData.containsKey(key)){
            return cacheData.get(key).getTotal();
        }else{
            return 0;
        }
    }

    public boolean isOver(String key){
        if(!cacheData.containsKey(key)){
            return false;
        }
        Counter count = cacheData.get(key);
        long ctimes = System.currentTimeMillis() - count.getStartTime();
        if(ctimes > this.timems){
            synchronized (cacheData){
                cacheData.remove(key);
            }
            return false;
        }else{
            return count.getTotal() > this.maxTimes;
        }

    }

    public void clearExpire(){
        for (Iterator<Counter> it = cacheData.values().iterator(); it.hasNext();)
        {
            Counter count = it.next();
            long ctimes = System.currentTimeMillis() - count.getStartTime();
            if(ctimes > this.timems){
                synchronized (cacheData){
                   it.remove();
                }
            }
        }
    }

    class Counter{
        private AtomicInteger errors ;
        private long startTime;
        public Counter() {
            this.errors = new AtomicInteger(1);
            this.startTime = System.currentTimeMillis();
        }

        public int incr(){
            return this.errors.incrementAndGet();
        }

        public int getTotal() {
            return errors.intValue();
        }

        public long getStartTime() {
            return startTime;
        }
    }
}
