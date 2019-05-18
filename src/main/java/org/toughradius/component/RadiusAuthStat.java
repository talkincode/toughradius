package org.toughradius.component;

import org.springframework.stereotype.Component;
import org.toughradius.common.SpinLock;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.concurrent.atomic.AtomicInteger;

@Component
public class RadiusAuthStat {

    private final static SpinLock lock = new SpinLock();

    public final static String ACCEPT = "accept";
    public final static String NOT_EXIST = "not_exist";
    public final static String PWD_ERR = "pwd_err";
    public final static String LIMIT_ERR = "limit_err";
    public final static String RATE_ERR = "rate_err";
    public final static String STATUS_ERR = "status_err";
    public final static String BRAS_LIMIT_ERR = "bras_limit_err";
    public final static String BIND_ERR = "bind_err";
    public final static String OTHER_ERR = "other_err";
    public final static String DROP = "drop";


    private ConcurrentLinkedDeque<long[]> authAcceptStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> AuthNotExistErrStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> AuthPwdErrStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> AuthLimitErrStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> AuthRateErrStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> AuthStatusErrStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> AuthBrasLimitErrStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> AuthBindErrStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> AuthOtherErrStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> AuthDropStat = new ConcurrentLinkedDeque<long[]>();

    public Map getData(){
        try{
            lock.lock();
            Map data = new HashMap();
            data.put("authAcceptStat", authAcceptStat.toArray());
            data.put("AuthNotExistErrStat", AuthNotExistErrStat.toArray());
            data.put("AuthPwdErrStat", AuthPwdErrStat.toArray());
            data.put("AuthLimitErrStat", AuthLimitErrStat.toArray());
            data.put("AuthRateErrStat", AuthRateErrStat.toArray());
            data.put("AuthStatusErrStat", AuthStatusErrStat.toArray());
            data.put("AuthBrasLimitErrStat", AuthBrasLimitErrStat.toArray());
            data.put("AuthBindErrStat", AuthOtherErrStat.toArray());
            data.put("AuthOtherErrStat", AuthOtherErrStat.toArray());
            data.put("AuthDropStat", AuthDropStat.toArray());
            return data;
        }finally {
            lock.unLock();
        }
    }

    private AtomicInteger authAccept = new AtomicInteger(0);
    private AtomicInteger notExist = new AtomicInteger(0);
    private AtomicInteger pwdErr = new AtomicInteger(0);
    private AtomicInteger limitErr = new AtomicInteger(0);
    private AtomicInteger rateErr = new AtomicInteger(0);
    private AtomicInteger statusErr = new AtomicInteger(0);
    private AtomicInteger brasLimitErr = new AtomicInteger(0);
    private AtomicInteger bindErr = new AtomicInteger(0);
    private AtomicInteger othertErr = new AtomicInteger(0);
    private AtomicInteger authDrop = new AtomicInteger(0);

    public void runStat(){
        try{
            lock.lock();
            long ctime =  System.currentTimeMillis();
            authAcceptStat.addLast(new long[]{ctime,authAccept.getAndSet(0)});
            if(authAcceptStat.size()>180){
                authAcceptStat.removeFirst();
            }
            AuthNotExistErrStat.addLast(new long[]{ctime,notExist.getAndSet(0)});
            if(AuthNotExistErrStat.size()>180){
                AuthNotExistErrStat.removeFirst();
            }
            AuthPwdErrStat.addLast(new long[]{ctime,pwdErr.getAndSet(0)});
            if(AuthPwdErrStat.size()>180){
                AuthPwdErrStat.removeFirst();
            }
            AuthLimitErrStat.addLast(new long[]{ctime,limitErr.getAndSet(0)});
            if(AuthLimitErrStat.size()>180){
                AuthLimitErrStat.removeFirst();
            }
            AuthRateErrStat.addLast(new long[]{ctime,rateErr.getAndSet(0)});
            if(AuthRateErrStat.size()>180){
                AuthRateErrStat.removeFirst();
            }
            AuthStatusErrStat.addLast(new long[]{ctime,statusErr.getAndSet(0)});
            if(AuthStatusErrStat.size()>180){
                AuthStatusErrStat.removeFirst();
            }
            AuthBrasLimitErrStat.addLast(new long[]{ctime,brasLimitErr.getAndSet(0)});
            if(AuthBrasLimitErrStat.size()>180){
                AuthBrasLimitErrStat.removeFirst();
            }
            AuthOtherErrStat.addLast(new long[]{ctime,othertErr.getAndSet(0)});
            if(AuthOtherErrStat.size()>180){
                AuthOtherErrStat.removeFirst();
            }
            AuthBindErrStat.addLast(new long[]{ctime,bindErr.getAndSet(0)});
            if(AuthBindErrStat.size()>180){
                AuthBindErrStat.removeFirst();
            }
            AuthDropStat.addLast(new long[]{ctime,authDrop.getAndSet(0)});
            if(AuthDropStat.size()>180){
                AuthDropStat.removeFirst();
            }
        }finally {
            lock.unLock();
        }

    }

    public void update(String type){
        switch (type) {
            case ACCEPT:
                authAccept.incrementAndGet();
                break;
            case NOT_EXIST:
                notExist.incrementAndGet();
                break;
            case PWD_ERR:
                pwdErr.incrementAndGet();
                break;
            case LIMIT_ERR:
                limitErr.incrementAndGet();
                break;
            case RATE_ERR:
                rateErr.incrementAndGet();
                break;
            case STATUS_ERR:
                statusErr.incrementAndGet();
                break;
            case BRAS_LIMIT_ERR:
                brasLimitErr.incrementAndGet();
                break;
            case BIND_ERR:
                bindErr.incrementAndGet();
                break;
            case OTHER_ERR:
                othertErr.incrementAndGet();
                break;
            case DROP:
                authDrop.incrementAndGet();
                break;
        }
    }
}
