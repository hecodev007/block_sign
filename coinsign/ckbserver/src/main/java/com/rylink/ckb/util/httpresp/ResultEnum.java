package com.rylink.ckb.util.httpresp;

public enum ResultEnum {
  UNKNOWN_ERROR(20000, "未知错误"),
  PARAMS_ERROR(20001, "参数错误"),
  SIGN_ERROR(20002, "签名错误"),
  SUCCESS(0, "success"),
  SYSTEM_ERROR(10000, "系统错误");

  private Integer code;

  private String msg;

  ResultEnum(Integer code, String msg) {
    this.code = code;
    this.msg = msg;
  }

  public Integer getCode() {
    return code;
  }

  public String getMsg() {
    return msg;
  }
}
