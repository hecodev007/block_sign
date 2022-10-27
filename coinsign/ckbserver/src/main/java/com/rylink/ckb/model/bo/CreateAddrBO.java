package com.rylink.ckb.model.bo;

import lombok.Data;

import javax.validation.constraints.Max;
import javax.validation.constraints.Min;
import javax.validation.constraints.NotEmpty;

@Data
public class CreateAddrBO {

  // 生成数量
  @Max(value = 1000000, message = "num数量不能大于1000000")
  @Min(value = 1, message = "num数量不能小于1")
  private int num;

  // 订单号
  @NotEmpty(message = "orderId不能为空")
  private String orderId;

  // 商户ID
  @NotEmpty(message = "mchId不能为空")
  private String mchId;

  // 币种名称
  @NotEmpty(message = "coinName不能为空")
  private String coinName;
}
