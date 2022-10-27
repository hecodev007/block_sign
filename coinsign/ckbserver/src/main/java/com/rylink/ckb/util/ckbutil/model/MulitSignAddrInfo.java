package com.rylink.ckb.util.ckbutil.model;

import lombok.Data;

import java.util.List;

// 地址生成信息
@Data
public class MulitSignAddrInfo {

  // 参与生成的地址信息
  private List<AddrInfo> addressInfo;

  // 主网地址
  private String mainAddress;

  // 测试网地址
  private String testAddress;

  // args
  private String args;
}
