package org.toughradius.component;

import org.toughradius.common.DateTimeUtil;
import org.toughradius.common.PageResult;
import org.toughradius.common.ValidateUtil;
import org.toughradius.entity.TraceMessage;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class Syslogger {

    public final static String RADIUSD = "radiusd";
    public final static String SYSTEM = "system";
    public final static String BRAS = "bras";
    public final static String API = "api";
    public final static String OTHER = "other";


    private Log logger = LogFactory.getLog(Syslogger.class);

    private Map<String,ArrayDeque<TraceMessage>> traceMap = null;

    public Syslogger() {
        traceMap = new ConcurrentHashMap<>();
        traceMap.put(RADIUSD,new ArrayDeque<TraceMessage>());
        traceMap.put(SYSTEM,new ArrayDeque<TraceMessage>());
        traceMap.put(API,new ArrayDeque<TraceMessage>());
        traceMap.put(BRAS,new ArrayDeque<TraceMessage>());
    }

    public Map<String, ArrayDeque<TraceMessage>> getTraceMap() {
        return traceMap;
    }


    public void trace(String username, String message, String type){
        ArrayDeque<TraceMessage> traceQueue = traceMap.get(type);
        if(traceQueue != null){
            synchronized (traceQueue){
                traceQueue.addFirst(new TraceMessage(username,DateTimeUtil.getDateTimeString(),message, type));
                if(traceQueue.size()>10000){
                    traceQueue.pollLast();
                }
            }
        }


    }

    public void trace(String message,String type){
        ArrayDeque<TraceMessage> traceQueue = traceMap.get(type);
        if(traceQueue != null)
        {
            synchronized (traceQueue){
                traceQueue.addFirst(new TraceMessage("all",DateTimeUtil.getDateTimeString(),message, type));
                if(traceQueue.size()>10000){
                    traceQueue.pollLast();
                }
            }
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
    }

    public void error(String message,Throwable e,String type){
        logger.error(message);
        StringBuilder buf = new StringBuilder(message);
        buf.append("\n");
        for(StackTraceElement err:  e.getStackTrace()){
            buf.append(err.toString()).append("\n");
        }
        trace(buf.toString(),type);
    }

    public void error(String username,String message,String type){
        logger.error(String.format("%s:%s", username,message));
        trace(username,message,type);
    }

    public void error(String username,String message,Throwable e,String type){
        logger.error(String.format("%s:%s", username,message),e);
        StringBuilder buf = new StringBuilder(message);
        buf.append("\n");
        for(StackTraceElement err:  e.getStackTrace()){
            buf.append(err.toString()).append("\n");
        }
        trace(username,buf.toString(),type);
    }


    public PageResult<TraceMessage> queryMessage(int pos, int count, String startTime,String endTime, String type, String username, String keyword){
        ArrayDeque<TraceMessage> traceQueue = traceMap.get(type);
        if(traceQueue == null)
            return new PageResult<TraceMessage>(pos, 0, null);;

        synchronized ( traceQueue ){
            int total = 0;
            int start = pos+1;
            int end = pos +  count ;
            List<TraceMessage> messages = new ArrayList<TraceMessage>();
            for (TraceMessage message : traceQueue) {
                if(ValidateUtil.isNotEmpty(type)&&!message.getType().equalsIgnoreCase(type)) {
                    continue;
                }
                if(ValidateUtil.isNotEmpty(username)&&!message.getName().contains(username)) {
                    continue;
                }
                if(ValidateUtil.isNotEmpty(keyword)&&!message.getMsg().contains(keyword)) {
                    continue;
                }
                if(ValidateUtil.isNotEmpty(startTime)&&DateTimeUtil.compareSecond(message.getTime(),startTime)<0){
                    continue;
                }
                if(ValidateUtil.isNotEmpty(endTime)&&DateTimeUtil.compareSecond(message.getTime(),endTime)>0){
                    continue;
                }
                total++;
                if (total >= start && total <= end) {
                    messages.add(message.copy());
                }
            }
            return new PageResult<TraceMessage>(pos, total, messages);
        }

    }


}
