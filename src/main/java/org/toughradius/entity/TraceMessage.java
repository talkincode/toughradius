package org.toughradius.entity;

public class TraceMessage {
    private String name;
    private String type;
    private String time;
    private String msg;

    public TraceMessage(String name, String time, String msg,String type) {
        this.name = name;
        this.time = time;
        this.msg = msg;
        this.type = type;
    }

    public TraceMessage copy(){
        return new TraceMessage(name,time,msg,type);
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getTime() {
        return time;
    }

    public void setTime(String time) {
        this.time = time;
    }

    public String getMsg() {
        return msg;
    }

    public void setMsg(String msg) {
        this.msg = msg;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }
}
