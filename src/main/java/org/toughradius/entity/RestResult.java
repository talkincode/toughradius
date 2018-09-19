package org.toughradius.entity;


public class RestResult<T> {

    public final static int SUCCESS = 0;
    public final static int UNKNOW_ERROR = 1;
    public final static int LOGIN_ERROR = 101;
    public final static int LOGIN_EXPIRE = 102;


    public final static String MSG_SUCCESS = "操作成功";

    public final static RestResult OptSuccess = new RestResult(SUCCESS,"操作成功");
    public final static RestResult LoginSuccess = new RestResult(SUCCESS,"用户登录成功");
    public final static RestResult LoginExpire = new RestResult(LOGIN_EXPIRE,"会话过期请重新登录");
    public final static RestResult LoginError = new RestResult(LOGIN_ERROR,"用户登录错误");
    public final static RestResult SupervisorError = new RestResult(LOGIN_ERROR,"系统服务调用失败");

    private int code;
    private String msg;
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

    public RestResult(int code, String msg) {
        this.code = code;
        this.msg = msg;
    }
    public RestResult() {

    }


    public RestResult(int code, String msg, T data) {
        this.code = code;
        this.msg = msg;
        this.data = data;
    }

}
