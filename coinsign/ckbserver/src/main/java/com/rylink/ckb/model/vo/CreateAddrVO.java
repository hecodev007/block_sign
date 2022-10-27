package com.rylink.ckb.model.vo;

import lombok.Data;

import java.util.List;

@Data
public class CreateAddrVO {

    private int num;
    private String orderId;
    private String mchId;
    private String coinName;
    private List<String> addrs;

}
