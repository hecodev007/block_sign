package com.rylink.ckb.util.ckbutil.model;

import com.fasterxml.jackson.annotation.JsonInclude;
import lombok.Data;
import org.codehaus.jackson.annotate.JsonProperty;
import org.nervos.ckb.type.transaction.Transaction;

@Data
public class SendInfo {
  @JsonInclude(JsonInclude.Include.NON_NULL)
  private String txid;

  @JsonInclude(JsonInclude.Include.NON_NULL)
  private Transaction tx;

  @JsonInclude(JsonInclude.Include.NON_NULL)
  private String errmsg;

  @JsonInclude(JsonInclude.Include.NON_NULL)
  @JsonProperty(value = "mchId")
  private String mchId;

  @JsonInclude(JsonInclude.Include.NON_NULL)
  @JsonProperty(value = "orderId")
  private String orderId;

  @JsonInclude(JsonInclude.Include.NON_NULL)
  @JsonProperty(value = "coinName")
  private String coinName;

  public SendInfo() {}

  public SendInfo(
      String txid, Transaction tx, String errmsg, String mchId, String orderId, String coinName) {
    this.txid = txid;
    this.tx = tx;
    this.errmsg = errmsg;
    this.mchId = mchId;
    this.orderId = orderId;
    this.coinName = coinName;
  }
}
