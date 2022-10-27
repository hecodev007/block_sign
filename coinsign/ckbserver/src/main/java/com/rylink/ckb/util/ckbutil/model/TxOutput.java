package com.rylink.ckb.util.ckbutil.model;

import lombok.Data;

import java.math.BigDecimal;

@Data
public class TxOutput {
  private String address;
  private BigDecimal amount;

  public TxOutput(String address, BigDecimal amount) {
    this.address = address;
    this.amount = amount;
  }
}
