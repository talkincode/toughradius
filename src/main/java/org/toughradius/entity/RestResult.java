package org.toughradius.entity;


public class RestResult<T> {

    public final static int SUCCESS = 0;
    public final static int FAILURE = 1;

    public final static String MSG_SUCCESS = "Success";
    public final static String MSG_FAILURE = "Failure";
    public final static RestResult OptSuccess = new RestResult(SUCCESS,MSG_SUCCESS);
    public final static RestResult OptFailure = new RestResult(FAILURE,MSG_FAILURE);

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
