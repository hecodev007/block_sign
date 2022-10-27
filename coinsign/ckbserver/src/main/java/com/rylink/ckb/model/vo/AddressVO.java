package com.rylink.ckb.model.vo;

import lombok.Data;
import org.nervos.ckb.address.CodeHashType;

@Data
public class AddressVO {
  private String address;
  private CodeHashType codeHashType;
  private String lockArgs;
  private Boolean vaild;
  private String netword;
}
