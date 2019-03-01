package org.toughradius.common;


import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ConcurrentLinkedDeque;

public class SysLogger {

    private final static ConcurrentLinkedDeque<String> logque = new ConcurrentLinkedDeque<String>();
    private Logger logger;

    private SysLogger(Class cls){
        logger =  LoggerFactory.getLogger(cls);
    }

    public static SysLogger getLogger(Class cls){
        return new SysLogger(cls);
    }

    private void addMessage(String msg, String level){
        synchronized (logque){
            if(logque.size()>=2048){
                logque.pollLast();
            }
            logque.addFirst(String.format("<span class=\"%s-line\">%s [%s] :: </span> %s", level, DateTimeUtil.getDateTimeString(), level, msg));
        }
    }

    public void info(String msg){
        logger.info(msg);
        addMessage(msg,"info");
    }

    public void debug(String msg){
        logger.debug(msg);
        addMessage(msg,"debug");
    }

    public void error(String msg){
        logger.error(msg);
        addMessage(msg,"error");
    }

    public void error(String msg, Throwable e){
        logger.error(msg,e);
        StringBuilder buf = new StringBuilder(msg);
        buf.append(e.getMessage());
        buf.append("\n");
        for(StackTraceElement err:  e.getStackTrace()){
            buf.append(err.toString()).append("\n");
        }
        addMessage(buf.toString(),"error");
    }

    public void error( Throwable e){
        StringBuilder buf = new StringBuilder(e.getMessage());
        buf.append("\n");
        for(StackTraceElement err:  e.getStackTrace()){
            buf.append(err.toString()).append("\n");
        }
        addMessage(buf.toString(),"error");
    }

    public static List<String> getMessages(){
        List<String> result = new ArrayList<>();
        synchronized (logque){
            while(logque.size() > 0){
                result.add(logque.pollLast());
            }
        }
        return result;
    }

    public static String getMessage(){
        return  logque.pollLast();
    }


}
