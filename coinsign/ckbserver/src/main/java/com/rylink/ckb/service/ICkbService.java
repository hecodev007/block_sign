package com.rylink.ckb.service;

import com.rylink.ckb.model.bo.CreateAddrBO;
import com.rylink.ckb.model.bo.TxTplBO;
import com.rylink.ckb.model.vo.AddressVO;
import com.rylink.ckb.util.ckbutil.model.AddrInfo;
import com.rylink.ckb.util.ckbutil.model.SendInfo;

import java.io.IOException;
import java.util.List;

public interface ICkbService {

  // 生成地址到csv
  List<AddrInfo> CreateAddrToCsv(CreateAddrBO bo) throws Exception;

  // 签名交易
  SendInfo GetSignHex(TxTplBO bo) throws IOException;

  // 地址验证
  AddressVO PaseAddress(String address) throws IOException;
}
