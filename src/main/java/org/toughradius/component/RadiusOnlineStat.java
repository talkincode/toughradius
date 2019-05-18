package org.toughradius.component;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.toughradius.common.SpinLock;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ConcurrentLinkedDeque;


@Component
public class RadiusOnlineStat {

    private final static SpinLock lock = new SpinLock();

    private ConcurrentLinkedDeque<long[]> onineStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> onlineDelayStat = new ConcurrentLinkedDeque<long[]>();

    @Autowired
    private OnlineCache onlineCache;

    public Map getData(){
        try{
            lock.lock();
            Map data = new HashMap();
            data.put("onineStat", onineStat.toArray());
            data.put("onlineDelayStat", onlineDelayStat.toArray());
            return data;
        }finally {
            lock.unLock();
        }
    }

    public void runStat() {
        try{
            lock.lock();
            long ctime = System.currentTimeMillis();
            long[] counts = onlineCache.getOnlineStat();
            onineStat.addLast(new long[]{ctime, counts[0]});
            if(onineStat.size()>180){
                onineStat.removeFirst();
            }
            onlineDelayStat.addLast(new long[]{ctime, counts[1]});
            if(onlineDelayStat.size()>180){
                onlineDelayStat.removeFirst();
            }
        }finally {
            lock.unLock();
        }
    }

}
