package org.toughradius.common;

public class RestResult<T> {

    public final static RestResult SUCCESS = new RestResult(0,"success");
    public final static RestResult UNKNOW = new RestResult(1,"unknow");
    private int code;
    private String msg;
    private String msgtype;
    private T data;

    public int getCode() {
        return code;
    }

    public void setCode(int code) {
        this.code = code;
    }

    public String getMsg() {
        return msg;
    }

    public void setMsg(String msg) {
        this.msg = msg;
    }

    public T getData() {
        return data;
    }

    public void setData(T data) {
        this.data = data;
    }

    public String getMsgtype() {
        return msgtype;
    }

    public void setMsgtype(String msgtype) {
        this.msgtype = msgtype;
    }

    public RestResult(int code, String msg) {
        this.code = code;
        this.msg = msg;
        if(code==0){
            setMsgtype("info");
        }else if(code>0){
            setMsgtype("error");
        }
    }

    public RestResult(int code, String msg, T data) {
        this.code = code;
        this.msg = msg;
        this.data = data;
        if(code==0){
            setMsgtype("info");
        }else if(code>0){
            setMsgtype("error");
        }
    }
}
