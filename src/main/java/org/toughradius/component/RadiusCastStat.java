package org.toughradius.component;

import org.springframework.stereotype.Component;
import org.toughradius.common.SpinLock;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.concurrent.atomic.AtomicInteger;

@Component
public class RadiusCastStat {

    private final static SpinLock lock = new SpinLock();

    private final ConcurrentLinkedDeque<long[]> authCastStat = new ConcurrentLinkedDeque<long[]>();
    private final ConcurrentLinkedDeque<long[]> acctUpdateCastStat = new ConcurrentLinkedDeque<long[]>();
    private final ConcurrentLinkedDeque<long[]> acctStartCastStat = new ConcurrentLinkedDeque<long[]>();
    private final ConcurrentLinkedDeque<long[]> acctStopCastStat = new ConcurrentLinkedDeque<long[]>();



    public Map getData(){
        try{
            lock.lock();
            Map data = new HashMap();
            data.put("authCastStat", authCastStat.toArray());
            data.put("acctUpdateCastStat", acctUpdateCastStat.toArray());
            data.put("acctStartCastStat", acctStartCastStat.toArray());
            data.put("acctStopCastStat", acctStopCastStat.toArray());
            return data;
        }finally {
            lock.unLock();
        }
    }

    private long lastAuthUpdate = System.currentTimeMillis();
    private AtomicInteger authNum = new AtomicInteger(0);
    private AtomicInteger authCastAvg = new AtomicInteger(0);
    private AtomicInteger authCastTotal  = new AtomicInteger(0);
    private final static SpinLock authCastStatLock = new SpinLock();

    public void updateAuth(int cast){
        long ctime =  System.currentTimeMillis();
        if((ctime - lastAuthUpdate)>=5000){
            this.lastAuthUpdate = ctime;
            authNum.set(0);
            authCastTotal.set(0);
            try{
                authCastStatLock.lock();
                authCastStat.addLast(new long[]{ctime,authCastAvg.getAndSet(0)});
                if(authCastStat.size() >= 180){
                    authCastStat.removeFirst();
                }
            }finally {
                authCastStatLock.unLock();
            }
        }else{
           int num =  authNum.incrementAndGet();
           int total = authCastTotal.addAndGet(cast);
           authCastAvg.set(total/num);
        }
    }


    private long lastAcctStartUpdate = System.currentTimeMillis();
    private AtomicInteger acctStartNum = new AtomicInteger(0);
    private AtomicInteger acctStartCastAvg = new AtomicInteger(0);
    private AtomicInteger acctStartCastTotal  = new AtomicInteger(0);
    private final static SpinLock acctStartCastStatLock = new SpinLock();

    public void updateAcctStart(int cast){
        long ctime =  System.currentTimeMillis();
        if((ctime - lastAcctStartUpdate)>=5000){
            this.lastAcctStartUpdate = ctime;
            acctStartNum.set(0);
            acctStartCastTotal.set(0);
            try{
                acctStartCastStatLock.lock();
                acctStartCastStat.addLast(new long[]{ctime,acctStartCastAvg.getAndSet(0)});
                if(acctStartCastStat.size() >= 180){
                    acctStartCastStat.removeFirst();
                }
            }finally {
                acctStartCastStatLock.unLock();
            }
        }else{
            int num =  acctStartNum.incrementAndGet();
            int total = acctStartCastTotal.addAndGet(cast);
            acctStartCastAvg.set(total/num);
        }
    }


    private long lastAcctUpdateUpdate = System.currentTimeMillis();
    private AtomicInteger acctUpdateNum = new AtomicInteger(0);
    private AtomicInteger acctUpdateCastAvg = new AtomicInteger(0);
    private AtomicInteger acctUpdateCastTotal  = new AtomicInteger(0);
    private final static SpinLock acctUpdateCastStatLock = new SpinLock();

    public void updateAcctUpdate(int cast){
        long ctime =  System.currentTimeMillis();
        if((ctime - lastAcctUpdateUpdate)>=5000){
            this.lastAcctUpdateUpdate = ctime;
            acctUpdateNum.set(0);
            acctUpdateCastTotal.set(0);
            try{
                acctUpdateCastStatLock.lock();
                acctUpdateCastStat.addLast(new long[]{ctime,acctUpdateCastAvg.getAndSet(0)});
                if(acctUpdateCastStat.size() >= 180){
                    acctUpdateCastStat.removeFirst();
                }
            }finally {
                acctUpdateCastStatLock.unLock();
            }
        }else{
            int num =  acctUpdateNum.incrementAndGet();
            int total = acctUpdateCastTotal.addAndGet(cast);
            acctUpdateCastAvg.set(total/num);
        }
    }

    private long lastAcctStopStop = System.currentTimeMillis();
    private AtomicInteger acctStopNum = new AtomicInteger(0);
    private AtomicInteger acctStopCastAvg = new AtomicInteger(0);
    private AtomicInteger acctStopCastTotal  = new AtomicInteger(0);
    private final static SpinLock acctStopCastStatLock = new SpinLock();

    public void updateAcctStop(int cast){
        long ctime =  System.currentTimeMillis();
        if((ctime - lastAcctStopStop)>=5000){
            this.lastAcctStopStop = ctime;
            acctStopNum.set(0);
            acctStopCastTotal.set(0);
            try{
                acctStopCastStatLock.lock();
                acctStopCastStat.addLast(new long[]{ctime,acctStopCastAvg.getAndSet(0)});
                if(acctStopCastStat.size() >= 180){
                    acctStopCastStat.removeFirst();
                }
            }finally {
                acctStopCastStatLock.unLock();
            }
        }else{
            int num =  acctStopNum.incrementAndGet();
            int total = acctStopCastTotal.addAndGet(cast);
            acctStopCastAvg.set(total/num);
        }
    }














}
