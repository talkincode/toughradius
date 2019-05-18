package org.toughradius.component;

import org.toughradius.common.DateTimeUtil;
import org.springframework.stereotype.Component;
import org.toughradius.common.SpinLock;

import java.util.Date;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.concurrent.atomic.AtomicInteger;

@Component
public class RadiusStat {

    private final static SpinLock lock = new SpinLock();
    private ConcurrentLinkedDeque<long[]> reqBytesStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> respBytesStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> authReqStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> authRespStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> acctReqStat = new ConcurrentLinkedDeque<long[]>();
    private ConcurrentLinkedDeque<long[]> acctRespStat = new ConcurrentLinkedDeque<long[]>();

    private AtomicInteger online = new AtomicInteger(0);
    private AtomicInteger authReqOld = new AtomicInteger(0);
    private AtomicInteger authRespOld = new AtomicInteger(0);
    private AtomicInteger authReq = new AtomicInteger(0);
    private AtomicInteger authAccept = new AtomicInteger(0);
    private AtomicInteger authReject = new AtomicInteger(0);
    private AtomicInteger authRejectdelay = new AtomicInteger(0);
    private AtomicInteger authDrop = new AtomicInteger(0);
    private AtomicInteger acctStart = new AtomicInteger(0);
    private AtomicInteger acctStop = new AtomicInteger(0);
    private AtomicInteger acctUpdate = new AtomicInteger(0);
    private AtomicInteger acctOn = new AtomicInteger(0);
    private AtomicInteger acctOff = new AtomicInteger(0);
    private AtomicInteger acctReqOld = new AtomicInteger(0);
    private AtomicInteger acctRespOld = new AtomicInteger(0);
    private AtomicInteger acctReq = new AtomicInteger(0);
    private AtomicInteger acctResp = new AtomicInteger(0);
    private AtomicInteger acctRetry = new AtomicInteger(0);
    private AtomicInteger acctDrop = new AtomicInteger(0);
    private AtomicInteger reqBytes = new AtomicInteger(0);
    private AtomicInteger respBytes = new AtomicInteger(0);
    private AtomicInteger reqBytesOld = new AtomicInteger(0);
    private AtomicInteger respBytesOld = new AtomicInteger(0);

    private int lastMaxReq = 0;
    private Date lastMaxReqDate = new Date();
    private int lastMaxResp = 0;
    private Date lastMaxRespDate = new Date();
    private int lastMaxReqBytes = 0;
    private Date lastMaxReqBytesDate = new Date();
    private int lastMaxRespBytes = 0;
    private Date lastMaxRespBytesDate = new Date();

    public Map getData(){
        try {
            lock.lock();
            Map data = new HashMap();
            data.put("reqBytesStat", reqBytesStat.toArray());
            data.put("respBytesStat", respBytesStat.toArray());
            data.put("authReqStat", authReqStat.toArray());
            data.put("authRespStat", authRespStat.toArray());
            data.put("acctReqStat", acctReqStat.toArray());
            data.put("acctRespStat", acctRespStat.toArray());
            data.put("online", online.intValue());
            data.put("authReqOld", authReqOld.intValue());
            data.put("authRespOld", authRespOld.intValue());
            data.put("authReq", authReq.intValue());
            data.put("authAccept", authAccept.intValue());
            data.put("authReject", authReject.intValue());
            data.put("authRejectdelay", authRejectdelay.intValue());
            data.put("authDrop", authDrop.intValue());
            data.put("acctStart", acctStart.intValue());
            data.put("acctStop", acctStop.intValue());
            data.put("acctUpdate", acctUpdate.intValue());
            data.put("acctOn", acctOn.intValue());
            data.put("acctOff", acctOff.intValue());
            data.put("acctReqOld", acctReqOld.intValue());
            data.put("acctRespOld", acctRespOld.intValue());
            data.put("acctReq", acctReq.intValue());
            data.put("acctResp", acctResp.intValue());
            data.put("acctRetry", acctRetry.intValue());
            data.put("acctDrop", acctDrop.intValue());
            data.put("reqBytes", reqBytes.intValue());
            data.put("respBytes", respBytes.intValue());
            data.put("reqBytesOld", reqBytesOld.intValue());
            data.put("respBytesOld", respBytesOld.intValue());

            data.put("lastMaxReq", lastMaxReq);
            data.put("lastMaxReqDate", DateTimeUtil.toDateTimeString(lastMaxReqDate));
            data.put("lastMaxResp", lastMaxResp);
            data.put("lastMaxRespDate", DateTimeUtil.toDateTimeString(lastMaxRespDate));
            data.put("lastMaxReqBytes", lastMaxReqBytes);
            data.put("lastMaxReqBytesDate", DateTimeUtil.toDateTimeString(lastMaxReqBytesDate));
            return data;
        }finally {
            lock.unLock();
        }
    }


    public void runStat(){
        try {
            lock.lock();
            long curtime = System.currentTimeMillis();

            int tmpReqBytes = getReqBytes() - getReqBytesOld();
            if (tmpReqBytes < 0)
                tmpReqBytes = 0;
            setReqBytesOld(getReqBytes());

            int tmpRespBytes = getRespBytes() - getRespBytesOld();
            if (tmpRespBytes < 0)
                tmpRespBytes = 0;
            setRespBytesOld(getRespBytes());
            if (reqBytesStat.size() >= 180) {
                reqBytesStat.pollFirst();
            }
            if (respBytesStat.size() >= 180) {
                respBytesStat.pollFirst();
            }
            reqBytesStat.addLast(new long[]{curtime, tmpReqBytes});
            respBytesStat.addLast(new long[]{curtime, tmpRespBytes});

            int tmpAuthReq = getAuthReq() - getAuthReqOld();
            if (tmpAuthReq < 0)
                tmpAuthReq = 0;
            setAuthReqOld(getAuthReq());

            int tmpAuthResp = getAuthAccept() + getAuthReject() - getAuthRespOld();
            if (tmpAuthResp < 0)
                tmpAuthResp = 0;
            setAuthRespOld(getAuthAccept() + getAuthReject());
            if (authReqStat.size() >= 180) {
                authReqStat.pollFirst();
            }
            if (authRespStat.size() >= 180) {
                authRespStat.pollFirst();
            }
            authReqStat.addLast(new long[]{curtime, tmpAuthReq});
            authRespStat.addLast(new long[]{curtime, tmpAuthResp});

            int tmpAcctReq = getAcctReq() - getAcctReqOld();
            if (tmpAcctReq < 0)
                tmpAcctReq = 0;
            setAcctReqOld(getAcctReq());

            int tmpAcctResp = getAcctResp() - getAcctRespOld();
            if (tmpAcctResp < 0)
                tmpAcctResp = 0;
            setAcctRespOld(getAcctResp());

            if (acctReqStat.size() >= 180) {
                acctReqStat.pollFirst();
            }
            if (acctRespStat.size() >= 180) {
                acctRespStat.pollFirst();
            }
            acctReqStat.addLast(new long[]{curtime, tmpAcctReq});
            acctRespStat.addLast(new long[]{curtime, tmpAcctResp});

            int reqbytesPercount = tmpReqBytes / 5;
            if (getLastMaxReqBytes() < reqbytesPercount) {
                setLastMaxReqBytes(reqbytesPercount);
                setLastMaxReqBytesDate(new Date());
            }

            int respbytesPercount = tmpRespBytes / 5;
            if (getLastMaxRespBytes() < respbytesPercount) {
                setLastMaxRespBytes(respbytesPercount);
                setLastMaxRespBytesDate(new Date());
            }

            int reqPercount = (tmpAuthReq + tmpAcctReq) / 5;
            if (getLastMaxReq() < reqPercount) {
                setLastMaxReq(reqPercount);
                setLastMaxReqDate(new Date());
            }


            int respPercount = (tmpAuthResp + tmpAcctResp) / 5;
            if (getLastMaxResp() < respPercount) {
                setLastMaxResp(respPercount);
                setLastMaxRespDate(new Date());
            }
        }finally {
            lock.unLock();
        }
    }

    public int getOnline() {
        return online.intValue();
    }

    public void cetOnline(int value) {
        this.online.set(value);
    }
    public void incrOnline() {
        this.online.incrementAndGet();
    }

    public int getAuthReqOld() {
        return authReqOld.intValue();
    }

    public void setAuthReqOld(int value) {
        this.authReqOld.set(value);
    }

    public int getAuthRespOld() {
        return authRespOld.intValue();
    }

    public void setAuthRespOld(int value) {
        this.authRespOld.set(value);
    }

    public int getAuthReq() {
        return authReq.intValue();
    }

    public void incrAuthReq() {
        this.authReq.incrementAndGet();
    }

    public int getAuthAccept() {
        return authAccept.intValue();
    }

    public void incrAuthAccept() {
        this.authAccept.incrementAndGet();
    }

    public int getAuthReject() {
        return authReject.intValue();
    }

    public void incrAuthReject() {
        this.authReject.incrementAndGet();
    }
    public int getAuthRejectdelay() {
        return authRejectdelay.intValue();
    }

    public void incrAuthRejectdelay() {
        this.authRejectdelay.incrementAndGet();
    }

    public int getAuthDrop() {
        return authDrop.intValue();
    }

    public void incrAuthDrop() {
        this.authDrop.incrementAndGet();
    }

    public int getAcctStart() {
        return acctStart.intValue();
    }

    public void incrAcctStart() {
        this.acctStart.incrementAndGet();
    }

    public int getAcctStop() {
        return acctStop.intValue();
    }

    public void incrAcctStop() {
        this.acctStop.incrementAndGet();
    }

    public int getAcctUpdate() {
        return acctUpdate.intValue();
    }

    public void incrAcctUpdate() {
        this.acctUpdate.incrementAndGet();
    }

    public int getAcctOn() {
        return acctOn.intValue();
    }

    public void incrAcctOn() {
        this.acctOn.intValue();
    }

    public int getAcctOff() {
        return acctOff.intValue();
    }

    public void incrAcctOff() {
        this.acctOff.incrementAndGet();
    }

    public int getAcctReqOld() {
        return acctReqOld.intValue();
    }

    public void setAcctReqOld(int value) {
        this.acctReqOld.set(value);
    }

    public int getAcctRespOld() {
        return acctRespOld.intValue();
    }

    public void setAcctRespOld(int value) {
        this.acctRespOld.set(value);
    }

    public int getAcctReq() {
        return acctReq.intValue();
    }

    public void incrAcctReq() {
        this.acctReq.incrementAndGet();
    }

    public int getAcctResp() {
        return acctResp.intValue();
    }

    public void incrAcctResp() {
        this.acctResp.incrementAndGet();
    }

    public int getAcctRetry() {
        return acctRetry.intValue();
    }

    public void incrAcctRetry() {
        this.acctRetry.incrementAndGet();
    }

    public int getAcctDrop() {
        return acctDrop.intValue();
    }

    public void incrAcctDrop() {
        this.acctDrop.intValue();
    }

    public int getReqBytes() {
        return reqBytes.intValue();
    }

    public void incrReqBytes(int value) {
        this.reqBytes.addAndGet(value);
    }

    public int getRespBytes() {
        return respBytes.intValue();
    }

    public void incrRespBytes(int value) {
        this.respBytes.addAndGet(value);
    }

    public int getReqBytesOld() {
        return reqBytesOld.intValue();
    }

    public void setReqBytesOld(int value) {
        this.reqBytesOld.set(value);
    }

    public int getRespBytesOld() {
        return respBytesOld.intValue();
    }

    public void setRespBytesOld(int value) {
        this.respBytesOld.set(value);
    }

    public int getLastMaxReq() {
        return lastMaxReq;
    }

    public void setLastMaxReq(int value) {
        this.lastMaxReq = value;
    }

    public Date getLastMaxReqDate() {
        return lastMaxReqDate;
    }

    public void setLastMaxReqDate(Date lastMaxReqDate) {
        this.lastMaxReqDate = lastMaxReqDate;
    }

    public int getLastMaxResp() {
        return lastMaxResp;
    }

    public void setLastMaxResp(int lastMaxResp) {
        this.lastMaxResp = lastMaxResp;
    }

    public Date getLastMaxRespDate() {
        return lastMaxRespDate;
    }

    public void setLastMaxRespDate(Date lastMaxRespDate) {
        this.lastMaxRespDate = lastMaxRespDate;
    }

    public int getLastMaxReqBytes() {
        return lastMaxReqBytes;
    }

    public void setLastMaxReqBytes(int lastMaxReqBytes) {
        this.lastMaxReqBytes = lastMaxReqBytes;
    }

    public Date getLastMaxReqBytesDate() {
        return lastMaxReqBytesDate;
    }

    public void setLastMaxReqBytesDate(Date lastMaxReqBytesDate) {
        this.lastMaxReqBytesDate = lastMaxReqBytesDate;
    }

    public int getLastMaxRespBytes() {
        return lastMaxRespBytes;
    }

    public void setLastMaxRespBytes(int lastMaxRespBytes) {
        this.lastMaxRespBytes = lastMaxRespBytes;
    }

    public Date getLastMaxRespBytesDate() {
        return lastMaxRespBytesDate;
    }

    public void setLastMaxRespBytesDate(Date lastMaxRespBytesDate) {
        this.lastMaxRespBytesDate = lastMaxRespBytesDate;
    }
}
