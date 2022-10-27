package com.rylink.ckb.util.ckbutil.model;

import com.fasterxml.jackson.annotation.JsonInclude;
import lombok.Data;


//地址生成信息
@Data
public class AddrInfo {

    //主网地址
    private String mainAddress;

    //测试网地址
    private String testAddress;

    //私钥
    @JsonInclude(JsonInclude.Include.NON_NULL)
    private  String privKey;

    //公钥
    @JsonInclude(JsonInclude.Include.NON_NULL)
    private String pubkey;

    //lock-arg
    @JsonInclude(JsonInclude.Include.NON_NULL)
    private  String lockArg;
}
