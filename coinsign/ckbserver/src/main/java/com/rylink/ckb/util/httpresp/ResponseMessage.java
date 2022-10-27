package com.rylink.ckb.util.httpresp;

import lombok.Data;

//HttpRequest请求返回的最外层对象,用一种统一的格式返回给前端

@Data
public class ResponseMessage<T> {
    //错误码
    private int code;

    //信息描述
    private String msg;

    //具体的信息内容
    private T data;
}
