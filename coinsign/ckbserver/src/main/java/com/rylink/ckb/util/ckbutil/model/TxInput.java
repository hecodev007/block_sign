package com.rylink.ckb.util.ckbutil.model;

import lombok.Data;

import java.math.BigDecimal;

@Data
public class TxInput {
  private String privateKey;
  private String address;
  private String txid;
  private Integer index;
  private BigDecimal amount;

  public TxInput(String privateKey, String address, String txid, Integer index, BigDecimal amount) {
    this.privateKey = privateKey;
    this.address = address;
    this.txid = txid;
    this.index = index;
    this.amount = amount;
  }
}
