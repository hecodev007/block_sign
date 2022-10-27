package com.rylink.ckb.util.httpresp;

//自定义的异常类
public class MyException extends RuntimeException {
    private int code;

    private String errMsg;

    public  MyException(ResultEnum resultEnum){
        this.code = resultEnum.getCode();
        this.errMsg = resultEnum.getMsg();
    }

    public Integer getCode() {
        return code;
    }

    public void setCode(Integer code) {
        this.code = code;
    }

}

