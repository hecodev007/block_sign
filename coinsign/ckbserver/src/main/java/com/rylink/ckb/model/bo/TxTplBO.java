package com.rylink.ckb.model.bo;

import com.rylink.ckb.util.ckbutil.model.TxInput;
import com.rylink.ckb.util.ckbutil.model.TxOutput;
import lombok.Data;

import java.math.BigDecimal;
import java.util.List;

@Data
public class TxTplBO {

  private List<TxInput> inputs;
  private List<TxOutput> outputs;
  private BigDecimal txFee;
  private String changeAddr;
  private String mchId;
  private String orderId;
  private String coinName;

  public static String CheckBo(TxTplBO bo) {
    if (bo.getChangeAddr().equals("")) {
      return "error changeAddr";
    }
    if (bo.getTxFee().compareTo(BigDecimal.ZERO) == -1) {
      return "error fee";
    }
    if (bo.getInputs().size() == 0) {
      return "error inputs";
    }

    if (bo.getOutputs().size() == 0) {
      return "error outputs";
    }
    BigDecimal.valueOf(61);
    for (TxOutput out : bo.getOutputs()) {
      if (out.getAmount().compareTo(BigDecimal.valueOf(61)) == -1) {
        return "error outputs amount,less 61";
      }
    }

    return "";
  }
}
