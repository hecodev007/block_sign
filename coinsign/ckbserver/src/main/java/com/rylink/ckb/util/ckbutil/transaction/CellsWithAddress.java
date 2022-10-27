package com.rylink.ckb.util.ckbutil.transaction;

import org.nervos.ckb.type.cell.CellInput;

import java.util.List;

/** Copyright © 2019 Nervos Foundation. All rights reserved. */
public class CellsWithAddress {
  public List<CellInput> inputs;
  public String address;

  public CellsWithAddress(List<CellInput> inputs, String address) {
    this.inputs = inputs;
    this.address = address;
  }
}
