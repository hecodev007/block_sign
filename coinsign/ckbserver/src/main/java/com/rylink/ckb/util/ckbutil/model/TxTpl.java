package com.rylink.ckb.util.ckbutil.model;

import lombok.Data;
import org.nervos.ckb.system.type.SystemScriptCell;

import java.math.BigDecimal;
import java.util.List;

@Data
public class TxTpl {
  private List<TxInput> inputs;
  private List<TxOutput> outputs;
  private BigDecimal txFee;
  private String changeAddr;
  private SystemScriptCell systemScriptCell;
  private SystemScriptCell systemMultiSigCell;

  public TxTpl(
      List<TxInput> inputs,
      List<TxOutput> outputs,
      BigDecimal txFee,
      String changeAddr,
      SystemScriptCell systemScriptCell,
      SystemScriptCell systemMultiSigCell) {
    this.inputs = inputs;
    this.outputs = outputs;
    this.txFee = txFee;
    this.changeAddr = changeAddr;
    this.systemScriptCell = systemScriptCell;
    this.systemMultiSigCell = systemMultiSigCell;
  }
}
