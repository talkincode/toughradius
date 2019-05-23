package org.toughradius.component;

import org.springframework.stereotype.Component;
import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.PageResult;
import org.toughradius.common.SpinLock;
import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.TraceMessage;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;
import java.util.stream.Stream;

/**
 * 内存日志工具
 */
@Component
public class Memarylogger {

    public final static String RADIUSD = "radiusd";
    public final static String RADIUSD_COA = "radiusd_coa";
    public final static String SYSTEM = "system";
    public final static String PORTAL = "portal";
    public final static String API = "api";
    public final static String ERROR = "error";

    private Log logger = LogFactory.getLog(Memarylogger.class);

    private Map<String,LoggerDeque> traceMap = null;

    public Memarylogger() {
        traceMap = new ConcurrentHashMap<>();
        traceMap.put(RADIUSD,new LoggerDeque(10000));
        traceMap.put(RADIUSD_COA,new LoggerDeque(10000));
        traceMap.put(SYSTEM,new LoggerDeque(10000));
        traceMap.put(API,new LoggerDeque(10000));
        traceMap.put(PORTAL,new LoggerDeque(10000));
        traceMap.put(ERROR,new LoggerDeque(10000));
    }

    public Map<String, LoggerDeque> getTraceMap() {
        return traceMap;
    }


    public void trace(String username, String message, String type){
        LoggerDeque traceQueue = traceMap.get(type);
        if(traceQueue != null) {
            traceQueue.add(new TraceMessage(username, DateTimeUtil.getDateTimeString(), message, type));
        }
    }

    public void trace(String message,String type){
        LoggerDeque traceQueue = traceMap.get(type);
        if(traceQueue != null) {
            traceQueue.add(new TraceMessage(null, DateTimeUtil.getDateTimeString(), message, type));
        }
    }

    public void print(String message){
        logger.info(message);
    }

    public void errprint(String message){
        logger.error(message);
    }

    public void info(String message, String type){
        logger.info(message);
        trace(message,type);
    }

    public void info(String username,String message,String type){
        logger.info(String.format("%s:%s", username,message,type));
        trace(username,message,type);
    }

    public void error(String message,String type){
        logger.error(message);
        trace(message,type);
        trace(message,ERROR);
    }

    public void error(String message,Throwable e,String type){
        logger.error(message,e);
        StringBuilder buf = new StringBuilder(message);
        buf.append("\n");
        for(StackTraceElement err:  e.getStackTrace()){
            buf.append(err.toString()).append("\n");
        }
        String errmessage = buf.toString();
        trace(errmessage,type);
        trace(errmessage,ERROR);
    }

    public void error(String username,String message,String type){
        logger.error(String.format("%s:%s", username,message));
        trace(username,message,type);
        trace(username,message,ERROR);
    }

    public void error(String username,String message,Throwable e,String type){
        logger.error(String.format("%s:%s", username,message),e);
        StringBuilder buf = new StringBuilder(message);
        buf.append("\n");
        for(StackTraceElement err:  e.getStackTrace()){
            buf.append(err.toString()).append("\n");
        }
        String errmessage = buf.toString();
        trace(username,errmessage,type);
        trace(username,errmessage,ERROR);
    }

    private boolean filterData(TraceMessage message, String type,String username, String keyword,String startTime, String endTime){
        if(ValidateUtil.isNotEmpty(startTime) && startTime.length() == 16){
            startTime += ":00";
        }

        if(ValidateUtil.isNotEmpty(endTime) && endTime.length() == 16){
            endTime += ":59";
        }
        if(ValidateUtil.isNotEmpty(type)&&!message.getType().equalsIgnoreCase(type)) {
            return false;
        }
        if(ValidateUtil.isNotEmpty(username)&&!message.getName().contains(username)) {
            return false;
        }
        if(ValidateUtil.isNotEmpty(keyword)&&!message.getMsg().contains(keyword)) {
            return false;
        }
        if(ValidateUtil.isNotEmpty(startTime)&&DateTimeUtil.compareSecond(message.getTime(),startTime)<0){
            return false;
        }
        if(ValidateUtil.isNotEmpty(endTime)&&DateTimeUtil.compareSecond(message.getTime(),endTime)>0){
            return false;
        }
        return true;
    }

    public PageResult<TraceMessage> queryMessage(int pos, int count, String startTime,String endTime, String type, String username, String keyword){
        LoggerDeque traceQueue = traceMap.get(type);
        if(traceQueue == null)
            return new PageResult<>(0, 0, null);
        List<TraceMessage> copyList = new ArrayList<>(traceQueue.getQueue());
        long total  = copyList.stream().filter(x -> filterData(x, type, username, keyword, startTime,endTime)).count();
        Stream<TraceMessage> stream = copyList.stream().filter(x ->  filterData(x, type, username, keyword, startTime,endTime));
        List<TraceMessage> data = stream.skip(pos).limit(count).collect(Collectors.toList());
        return new PageResult<>(pos, total, data);
    }

    static class LoggerDeque {
        private final static SpinLock lock = new SpinLock();
        private final ArrayDeque<TraceMessage> queue = new ArrayDeque<>();
        private int max = 10000;
        public LoggerDeque(int max) {
            this.max = max;
        }

        public ArrayDeque<TraceMessage> getQueue() {
            return queue;
        }

        public void add(TraceMessage message){
            try{
                lock.lock();
                queue.addFirst(message);
                if(queue.size()>this.max){
                    queue.pollLast();
                }
            }finally {
                lock.unLock();
            }
        }
    }


}
