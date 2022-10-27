package com.rylink.ckb.service.impl;

import com.rylink.ckb.model.bo.CreateAddrBO;
import com.rylink.ckb.model.bo.TxTplBO;
import com.rylink.ckb.model.vo.AddressVO;
import com.rylink.ckb.service.ICkbService;
import com.rylink.ckb.util.ckbutil.AddrUtil;
import com.rylink.ckb.util.ckbutil.MultiKeySingleSigTx;
import com.rylink.ckb.util.ckbutil.model.AddrInfo;
import com.rylink.ckb.util.ckbutil.model.SendInfo;
import com.rylink.ckb.util.ckbutil.model.TxInput;
import com.rylink.ckb.util.ckbutil.model.TxTpl;
import com.rylink.ckb.util.crypt.AESUtil;
import com.rylink.ckb.util.fileutil.ReadCkbCsv;
import com.rylink.ckb.util.fileutil.csvutil;
import com.rylink.ckb.util.fileutil.files;
import com.rylink.ckb.util.fileutil.model.CsvCtx;
import com.rylink.ckb.util.springutil.DicMap;
import lombok.extern.slf4j.Slf4j;
import org.nervos.ckb.address.AddressUtils;
import org.nervos.ckb.address.CodeHashType;
import org.nervos.ckb.address.Network;
import org.nervos.ckb.system.type.SystemScriptCell;
import org.nervos.ckb.type.OutPoint;
import org.nervos.ckb.utils.Numeric;
import org.springframework.stereotype.Service;

import java.io.File;
import java.io.IOException;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

@Slf4j
@Service("CkbService")
public class CkbServiceImpl implements ICkbService {
  @Override
  public List<AddrInfo> CreateAddrToCsv(CreateAddrBO bo) throws Exception {

    List<String[]> listA = new ArrayList<String[]>();
    List<String[]> listB = new ArrayList<String[]>();
    List<String[]> listC = new ArrayList<String[]>();
    List<String[]> listD = new ArrayList<String[]>();
    List<AddrInfo> addrs = new ArrayList<AddrInfo>();

    //        filename := createPath + "/" + params.CoinName + "_%s_usb_" + params.OrderId + ".csv"
    String filePath = "./csv/" + bo.getMchId() + "/";
    String fileNameA = bo.getCoinName() + "_a_usb_" + bo.getOrderId() + ".csv";
    String fileNameB = bo.getCoinName() + "_b_usb_" + bo.getOrderId() + ".csv";
    String fileNameC = bo.getCoinName() + "_c_usb_" + bo.getOrderId() + ".csv";
    String fileNameD = bo.getCoinName() + "_d_usb_" + bo.getOrderId() + ".csv";
    if (files.fileExists(new File(filePath + fileNameA))) {
      // 读取内容返回
      List<CsvCtx> ctxs = ReadCkbCsv.ReadByFile(filePath + fileNameA);
      if (ctxs.size() == 0) {
        log.error("read filse error");
        return addrs;
      } else {
        for (int i = 0; i < ctxs.size(); i++) {
          AddrInfo addrInfo = new AddrInfo();
          addrInfo.setTestAddress(ctxs.get(i).getAddress());
          addrs.add(addrInfo);
        }
      }
    } else {
      // 重新生成
      for (int i = 0; i < bo.getNum(); i++) {
        AddrInfo addrinfo = AddrUtil.createAddr();
        if (addrinfo == null) {
          log.error("create address error");
          return addrs;
        }
        //               log.info("addrinfo:{}", addrinfo);
        String aesKey = AESUtil.randKey();
        String aesCtx = AESUtil.ecrypt(addrinfo.getPrivKey(), aesKey);

        listA.add(new String[] {addrinfo.getMainAddress(), aesCtx, addrinfo.getTestAddress()});
        listB.add(new String[] {addrinfo.getMainAddress(), aesKey, addrinfo.getTestAddress()});
        listC.add(
            new String[] {
              addrinfo.getMainAddress(), addrinfo.getPrivKey(), addrinfo.getTestAddress()
            });
        listD.add(new String[] {addrinfo.getMainAddress(), addrinfo.getTestAddress()});
        // 清除私钥返回
        addrinfo.setPrivKey(null);
        addrinfo.setPubkey(null);
        addrs.add(addrinfo);
      }
      File dirFile = new File(filePath);
      if (!files.dirExists(dirFile)) {
        dirFile.mkdir();
      }
      if (listA.size() != listB.size()
          || listA.size() != listC.size()
          || listA.size() != listD.size()) {
        log.error("create address csv error");
        return addrs;
      }
      csvutil.writerCsvFile(filePath + fileNameA, listA);
      csvutil.writerCsvFile(filePath + fileNameB, listB);
      csvutil.writerCsvFile(filePath + fileNameC, listC);
      csvutil.writerCsvFile(filePath + fileNameD, listD);
    }
    return addrs;
  }

  @Override
  public SendInfo GetSignHex(TxTplBO bo) throws IOException {

    SendInfo sendInfo = new SendInfo();
    String errorMsg = TxTplBO.CheckBo(bo);
    if (!errorMsg.equals("")) {
      sendInfo.setErrmsg(errorMsg);
      return sendInfo;
    }
    // 普通转账cell
    SystemScriptCell systemScriptCell =
        new SystemScriptCell(
            "0x9bd7e06f3ecf4be0f2fcd2188b23f1b9fcc88e5d4b65a8637b17723bbda3cce8",
            new OutPoint(
                "0x71a7ba8fc96349fea0ed3a5c47992e3b4084b031a42264a018e0072e8172e46c",
                Numeric.toHexStringWithPrefix(BigInteger.ZERO)));

    // 多签cell
    SystemScriptCell systemMultiSigCell =
        new SystemScriptCell(
            "0x5c5069eb0857efc65e1bca0c07df34c31663b3622fd3876c876320fc9634e2a8",
            new OutPoint(
                "0x71a7ba8fc96349fea0ed3a5c47992e3b4084b031a42264a018e0072e8172e46c",
                Numeric.toHexStringWithPrefix(BigInteger.ONE)));

    for (int i = 0; i < bo.getInputs().size(); i++) {
      TxInput input = (TxInput) bo.getInputs().get(i);
      String privkey = DicMap.getDicVal(input.getAddress());
      if (privkey == null || privkey.equals("")) {
        sendInfo.setErrmsg("miss key, address:" + input.getAddress());
        return sendInfo;
      }
      // 私钥赋值
      bo.getInputs().get(i).setPrivateKey(privkey);
    }
    TxTpl tpl =
        new TxTpl(
            bo.getInputs(),
            bo.getOutputs(),
            bo.getTxFee(),
            bo.getChangeAddr(),
            systemScriptCell,
            systemMultiSigCell);

    sendInfo = MultiKeySingleSigTx.multiKeySingleSigTx(tpl);
    return sendInfo;
  }

  @Override
  public AddressVO PaseAddress(String address) {
    AddressVO vo = new AddressVO();
    if (address.equals("") || address.length() > 70) {
      vo.setVaild(false);
      vo.setAddress(address);
      return vo;
    }
    if (!address.startsWith("ckb")) {
      if (!address.startsWith("ckt")) {
        vo.setVaild(false);
        vo.setAddress(address);
        return vo;
      }
    }
    try {
      CodeHashType codeHashType = AddressUtils.parseAddressType(address);
      String payload = AddressUtils.parse(address);
      vo.setAddress(address);
      vo.setCodeHashType(codeHashType);
      vo.setLockArgs(payload);
      String addr = "";
      if (address.startsWith("ckb")) {
        AddressUtils util = new AddressUtils(Network.MAINNET);
        addr = util.generate(payload);
        vo.setNetword(Network.MAINNET.name());
      } else if (address.startsWith("ckt")) {
        AddressUtils util = new AddressUtils(Network.TESTNET);
        vo.setNetword(Network.TESTNET.name());
        addr = util.generate(payload);
      } else {
        vo = new AddressVO();
        vo.setAddress(address);
        vo.setVaild(false);
        return vo;
      }
      if (addr.equals(address)) {
        vo.setVaild(true);
      } else {
        vo.setVaild(false);
      }
      return vo;
    } catch (Exception e) {
      vo.setVaild(false);
      vo.setAddress(address);
      return vo;
    }
  }
}
