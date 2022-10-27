package com.rylink.ckb.util.ckbutil;

import com.rylink.ckb.util.ckbutil.model.SendInfo;
import com.rylink.ckb.util.ckbutil.model.TxInput;
import com.rylink.ckb.util.ckbutil.model.TxOutput;
import com.rylink.ckb.util.ckbutil.model.TxTpl;
import com.rylink.ckb.util.ckbutil.transaction.NumberUtils;
import com.rylink.ckb.util.ckbutil.transaction.ScriptGroupWithPrivateKeys;
import lombok.extern.slf4j.Slf4j;
import org.nervos.ckb.address.AddressUtils;
import org.nervos.ckb.address.CodeHashType;
import org.nervos.ckb.address.Network;
import org.nervos.ckb.system.type.SystemScriptCell;
import org.nervos.ckb.transaction.ScriptGroup;
import org.nervos.ckb.transaction.Secp256k1SighashAllBuilder;
import org.nervos.ckb.type.OutPoint;
import org.nervos.ckb.type.Script;
import org.nervos.ckb.type.Witness;
import org.nervos.ckb.type.cell.CellDep;
import org.nervos.ckb.type.cell.CellInput;
import org.nervos.ckb.type.cell.CellOutput;
import org.nervos.ckb.type.transaction.Transaction;
import org.nervos.ckb.utils.Calculator;
import org.nervos.ckb.utils.Convert;
import org.nervos.ckb.utils.Numeric;
import org.nervos.ckb.utils.Serializer;
import org.nervos.ckb.utils.address.AddressGenerator;

import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.*;

// 多对多，但是如果地址类型是多签地址,需要特殊处理
@Slf4j
public class MultiKeySingleSigTx {

  private static final Network network = Network.MAINNET; // 网络环境
  private static final int CKB_DECIMAL = 8; // 网络环境
  private static AddressUtils addressUtils; // 地址工具类
  private static BigDecimal UnitCKB; // 基本单元 1 = 100000000
  private static BigDecimal txFee; // 默认手续费
  //  private static SystemScriptCell systemScriptCell;
  //  private static SystemScriptCell systemMultiSigCell;

  static {
    addressUtils = new AddressUtils(network, CodeHashType.BLAKE160);
    UnitCKB = new BigDecimal("100000000");
    txFee = new BigDecimal("0.001");

    // script来自：get_block_by_number 高度0 参数为：0x0
    // 获取 transactions 数组的下标0 中 outputs的下标1的type进行Blake2b算法生成txHash,    index 固定为0x0

    // 多签script来自来自 get_block_by_number 高度1 参数为：0x1
    // 获取 transactions 数组的下标0 中 outputs的下标4的type进行Blake2b算法生成txHash,    index 固定为0x1
    //    Script script =
    //            new Script(
    //                    "0x00000000000000000000000000000000000000000000000000545950455f4944",
    //                    "0x8536c9d5d908bd89fc70099e4284870708b6632356aad98734fcf43f6f71c304",
    //                    "type");
    //    String cellHash = script.computeHash();
    //    systemScriptCell =
    //        new SystemScriptCell(
    //            "0x9bd7e06f3ecf4be0f2fcd2188b23f1b9fcc88e5d4b65a8637b17723bbda3cce8",
    //            new OutPoint(
    //                "0xbd864a269201d7052d4eb3f753f49f7c68b8edc386afc8bb6ef3e15a05facca2",
    //                Numeric.toHexStringWithPrefix(BigInteger.ZERO)));

    //    Script muiltScript =
    //        new Script(cell_deps
    //            "0x00000000000000000000000000000000000000000000000000545950455f4944",
    //            "0xd813c1b15bd79c8321ad7f5819e5d9f659a1042b72e64659a2c092be68ea9758",
    //            "type");
    //    String muiltCellHash = muiltScript.computeHash();
    //    System.out.println(muiltCellHash + ":muiltCellHash");
    //    systemMultiSigCell =
    //        new SystemScriptCell(
    //            "0x5c5069eb0857efc65e1bca0c07df34c31663b3622fd3876c876320fc9634e2a8",
    //            new OutPoint(
    //                "0xbd864a269201d7052d4eb3f753f49f7c68b8edc386afc8bb6ef3e15a05facca2",
    //                Numeric.toHexStringWithPrefix(BigInteger.ONE)));
  }

  // 多对多模式
  public static SendInfo multiKeySingleSigTx(TxTpl tpl) throws IOException {
    SendInfo sendInfo = new SendInfo();
    String errorMsg = checkTpl(tpl);
    if (!errorMsg.equals("")) {
      sendInfo.setErrmsg(errorMsg);
      return sendInfo;
    }
    BigDecimal changeAmount = BigDecimal.ZERO;
    BigDecimal toAmount = BigDecimal.ZERO;
    BigDecimal fromAmount = BigDecimal.ZERO;
    // 私钥列表，与地址一一对应
    List<ScriptGroupWithPrivateKeys> scriptGroupWithPrivateKeysList = new ArrayList<>();

    // 排序归类,一个key对应同样的utxo
    HashMap<String, List<CellInput>> mapPrivKey = new HashMap<String, List<CellInput>>();

    // 系统cell引用
    List<CellDep> deps = new ArrayList<CellDep>();
    deps.add(new CellDep(tpl.getSystemScriptCell().outPoint, CellDep.DEP_GROUP));
    //    if (containMultiSig) {
    //      deps.add(new CellDep(systemMultiSigCell.outPoint, CellDep.DEP_GROUP));
    //    }

    // 见证人脚本数量，暂时填入空
    List<Witness> witnesses = new ArrayList<>();
    for (TxInput in : tpl.getInputs()) {
      witnesses.add(new Witness(Witness.EMPTY_LOCK));
      if (in.getPrivateKey() == ""
          || in.getAmount().compareTo(BigDecimal.ZERO) == 0
          || in.getAmount().compareTo(BigDecimal.ZERO) == -1) {
        System.out.println(
            String.format(
                "txid:%s,address:%s,index:%d,amount:%s,error params",
                in.getTxid(), in.getAddress(), in.getIndex(), in.getAmount().toString()));
        sendInfo.setErrmsg(
            String.format(
                "txid:%s,address:%s,index:%d,amount:%s,error params",
                in.getTxid(), in.getAddress(), in.getIndex(), in.getAmount().toString()));
        return sendInfo;
      }
      if (AddressUtils.parseAddressType(in.getAddress()) == CodeHashType.MULTISIG) {
        sendInfo.setErrmsg("暂时不支持多签");
        return sendInfo;
      }
      CellInput cellInput =
          new CellInput(
              new OutPoint(in.getTxid(), Numeric.toHexString(String.valueOf(in.getIndex()))),
              Numeric.toHexString("0"));
      fromAmount = fromAmount.add(in.getAmount());
      if (mapPrivKey.containsKey(in.getPrivateKey())) {
        List<CellInput> inputs = new ArrayList<CellInput>();
        inputs.addAll(mapPrivKey.get(in.getPrivateKey()));
        inputs.add(cellInput);
        mapPrivKey.replace(in.getPrivateKey(), inputs);
      } else {
        List<CellInput> inputs = new ArrayList<CellInput>();
        inputs.add(cellInput);
        mapPrivKey.put(in.getPrivateKey(), inputs);
      }
    }

    // 组合私钥关系
    List<CellInput> cellInputs = new ArrayList<>();
    // startIndex：很奇怪的设计，如果是第一组地址则为0，如果第一组有3，那么第二组下标则为3,其实是代表0-2是A地址的，2之后是B地址
    int startIndex = 0;
    for (Map.Entry<String, List<CellInput>> entry : mapPrivKey.entrySet()) {
      scriptGroupWithPrivateKeysList.add(
          new ScriptGroupWithPrivateKeys(
              new ScriptGroup(NumberUtils.regionToList(startIndex, entry.getValue().size())),
              Collections.singletonList(entry.getKey())));

      cellInputs.addAll(entry.getValue());
      startIndex += entry.getValue().size();
    }

    // 封装out输出
    List<CellOutput> outputs = new ArrayList<CellOutput>();
    for (TxOutput outs : tpl.getOutputs()) {
      String blake160 = AddressUtils.parse(outs.getAddress());
      blake160 = blake160.startsWith("0x") ? blake160 : "0x" + blake160;
      CodeHashType codeHashType = AddressUtils.parseAddressType(outs.getAddress());
      if (codeHashType == CodeHashType.BLAKE160) {
        outputs.add(
            new CellOutput(
                Numeric.toHexString(outs.getAmount().multiply(UnitCKB).toBigInteger().toString()),
                new Script(tpl.getSystemScriptCell().cellHash, blake160, Script.TYPE)));
      } else if (codeHashType == CodeHashType.MULTISIG) {
        outputs.add(
            new CellOutput(
                Numeric.toHexString(outs.getAmount().multiply(UnitCKB).toBigInteger().toString()),
                new Script(tpl.getSystemMultiSigCell().cellHash, blake160, Script.TYPE)));
      } else {
        sendInfo.setErrmsg("addree type error");
        return sendInfo;
      }

      toAmount = toAmount.add(outs.getAmount());
    }

    int resultCompareTo = tpl.getTxFee().compareTo(BigDecimal.ZERO);
    if (-1 == resultCompareTo && resultCompareTo == 0) {
      // 自动计算手续费
    } else {
      txFee = tpl.getTxFee();
    }
    changeAmount = fromAmount.subtract(toAmount).subtract(txFee);
    if (changeAmount.compareTo(BigDecimal.ZERO) == 1) {
      // 找零
      if (changeAmount.compareTo(BigDecimal.valueOf(61)) == -1) {
        sendInfo.setErrmsg(
            String.format(
                "fromAmount:%s,toAmount:%s,changeAmount:%s,fee:%s,error amount,less 61",
                fromAmount.toString(),
                toAmount.toString(),
                changeAmount.toString(),
                txFee.toString()));
        return sendInfo;
      } else {
        List<TxOutput> outs = new ArrayList(tpl.getOutputs());
        // 需要找零
        String blake160 = AddressUtils.parse(tpl.getChangeAddr());
        blake160 = blake160.startsWith("0x") ? blake160 : "0x" + blake160;
        changeAmount = fromAmount.subtract(toAmount).subtract(txFee);
        log.info("找零金额：" + changeAmount.toString());
        log.info("fromAmount：" + fromAmount.toString());
        log.info("toAmount：" + toAmount.toString());
        log.info("txFee：" + txFee.toString());
        CodeHashType changeCodeHashType = AddressUtils.parseAddressType(tpl.getChangeAddr());
        if (changeCodeHashType == CodeHashType.BLAKE160) {
          outputs.add(
              new CellOutput(
                  Numeric.toHexString(changeAmount.multiply(UnitCKB).toBigInteger().toString()),
                  new Script(
                      tpl.getSystemScriptCell().cellHash,
                      Numeric.prependHexPrefix(blake160),
                      Script.TYPE)));
          outs.add(new TxOutput(tpl.getChangeAddr(), changeAmount));
        } else if (changeCodeHashType == CodeHashType.MULTISIG) {
          outputs.add(
              new CellOutput(
                  Numeric.toHexString(changeAmount.multiply(UnitCKB).toBigInteger().toString()),
                  new Script(
                      tpl.getSystemMultiSigCell().cellHash,
                      Numeric.prependHexPrefix(blake160),
                      Script.TYPE)));
          outs.add(new TxOutput(tpl.getChangeAddr(), changeAmount));
        } else {
          sendInfo.setErrmsg("changeAddree type error");
          return sendInfo;
        }

        tpl.setOutputs(outs);
      }
    }

    List<String> cellOutputsData = new ArrayList<>();
    for (int i = 0; i < outputs.size(); i++) {
      cellOutputsData.add("0x");
    }
    for (ScriptGroupWithPrivateKeys scriptGroupWithPrivateKeys : scriptGroupWithPrivateKeysList) {
      List<Witness> witnessesWithPrivateKeys = new ArrayList<>();
      for (int i = 0; i < scriptGroupWithPrivateKeys.privateKeys.size(); i++) {
        witnessesWithPrivateKeys.add(new Witness(Witness.EMPTY_LOCK));
      }
    }

    Transaction tx =
        new Transaction(
            Numeric.toHexString("0"),
            deps,
            Collections.emptyList(),
            cellInputs,
            outputs,
            cellOutputsData,
            witnesses);

    Secp256k1SighashAllBuilder signBuilder = new Secp256k1SighashAllBuilder(tx);

    for (ScriptGroupWithPrivateKeys scriptGroupWithPrivateKeys : scriptGroupWithPrivateKeysList) {
      signBuilder.sign(
          scriptGroupWithPrivateKeys.scriptGroup, scriptGroupWithPrivateKeys.privateKeys.get(0));
    }
    //    tx.witnesses = signedWitnesses;
    tx = signBuilder.buildTx();
    sendInfo.setErrmsg("");
    sendInfo.setTx(Convert.parseTransaction(tx));
    sendInfo.setTxid(tx.computeHash());
    errorMsg = checkSignOut(tpl, sendInfo);
    if (!errorMsg.equals("")) {
      sendInfo = new SendInfo();
      sendInfo.setErrmsg(errorMsg);
      log.error(errorMsg);
      return sendInfo;
    }
    return sendInfo;
  }

  // 校验入口参数,同时附加私钥
  // return 返回错误内容
  public static String checkTpl(TxTpl tpl) {
    if (tpl == null
        || tpl.getInputs().size() < 1
        || tpl.getOutputs().size() < 1
        || tpl.getTxFee().compareTo(BigDecimal.ZERO) <= 0
        || tpl.getTxFee().compareTo(BigDecimal.valueOf(61)) > 0
        || tpl.getSystemScriptCell() == null
        || tpl.getSystemMultiSigCell() == null) {
      System.out.println("sigTx参数异常");
      return ("sigTx参数异常");
    }

    SystemScriptCell systemScriptCell = tpl.getSystemScriptCell();
    SystemScriptCell systemMuiltScriptCell = tpl.getSystemMultiSigCell();

    if (systemScriptCell.cellHash.equals("")
        || !systemScriptCell.outPoint.index.equals(Numeric.toHexString("0"))
        || systemScriptCell.outPoint.txHash.equals("")) {
      return ("systemScriptCell参数异常");
    }
    if (systemMuiltScriptCell.cellHash.equals("")
        || !systemMuiltScriptCell.outPoint.index.equals(Numeric.toHexString("1"))
        || systemMuiltScriptCell.outPoint.txHash.equals("")) {
      return ("systemMuiltScriptCell参数异常");
    }

    // 找零地址检查
    CodeHashType changeCodeHashType = AddressUtils.parseAddressType(tpl.getChangeAddr());
    if (changeCodeHashType != CodeHashType.BLAKE160
        && changeCodeHashType != CodeHashType.MULTISIG) {
      return ("error changeAddress type");
    }

    // out检查
    for (TxOutput output : tpl.getOutputs()) {
      CodeHashType codeHashType = AddressUtils.parseAddressType(output.getAddress());
      if (codeHashType != CodeHashType.BLAKE160 && codeHashType != CodeHashType.MULTISIG) {
        return ("error out address type");
      }

      if (output.getAmount().compareTo(BigDecimal.valueOf(61)) == 0
          || output.getAmount().compareTo(BigDecimal.valueOf(61)) == -1) {
        return (String.format(
            "out:address:%s,amount:%s,error params,less amount 61",
            output.getAddress(), output.getAmount().toString()));
      }
    }

    // input检查
    for (TxInput input : tpl.getInputs()) {
      CodeHashType codeHashType = AddressUtils.parseAddressType(input.getAddress());
      if (codeHashType != CodeHashType.BLAKE160) {
        if (codeHashType == CodeHashType.MULTISIG) {
          return ("error input address type，暂时不支持多签地址");
        } else {
          return ("error input address type,未识别");
        }
      }
      if (input.getIndex() < 0) {
        return ("error input address index");
      }
      if (input.getTxid().equals("")) {
        return ("error input address txid");
      }
      if (input.getAddress().equals("")) {
        return ("error input address address");
      }
      if (input.getAmount().compareTo(BigDecimal.ZERO) == 0
          || input.getAmount().compareTo(BigDecimal.ZERO) == -1) {
        return (String.format(
            "input:address:%s,index:%d,amount:%s,error params",
            input.getAddress(), input.getIndex(), input.getAmount().toString()));
      }
      if (input.getPrivateKey() == null || input.getPrivateKey().equals("")) {
        return (String.format(
            "input:address:%s,index:%d,miss privkey", input.getAddress(), input.getIndex()));
      }
    }
    return "";
  }

  private static String checkSignOut(TxTpl tpl, SendInfo sendInfo) {

    for (TxOutput out : tpl.getOutputs()) {
      boolean has = false;
      String addr = "";
      BigDecimal amount = BigDecimal.ZERO;
      for (CellOutput cellout : sendInfo.getTx().outputs) {
        // 解码地址
        addr = AddressGenerator.generate(network, cellout.lock);
        amount = new BigDecimal(Numeric.toBigInt(cellout.capacity)).movePointLeft(CKB_DECIMAL);
        if (out.getAddress().equals(addr) && out.getAmount().compareTo(amount) == 0) {
          has = true;
          break;
        }
      }
      if (!has) {
        // 地址金额不对应
        return String.format(
            "tpl addreess: %s,tpl amount:%s,sign out address:%s,sign out amount:%s",
            out.getAddress(), out.getAmount().toString(), addr, amount.toString());
      }
    }
    return "";
  }

  public static String getTxFee(int inputNum, int outNum) {
    List<CellInput> cellInputs = new ArrayList<>();
    List<String> cellOutputsData = new ArrayList<>();
    List witnesses = new ArrayList<>();
    List<CellOutput> outputs = new ArrayList<CellOutput>();
    List<CellDep> deps = new ArrayList<CellDep>();
    deps.add(new CellDep(new OutPoint("0x0", "0x0"), CellDep.DEP_GROUP));
    deps.add(new CellDep(new OutPoint("0x0", "0x1"), CellDep.DEP_GROUP));
    //    new OutPoint("0x5fd5155ba542968a43fa8ff94555a04a676d6f364ba76d820a78985a87dccc0b",
    // Numeric.toHexString(String.valueOf(i));
    for (int i = 0; i < inputNum; i++) {
      CellInput cellInput =
          new CellInput(
              new OutPoint("0x0", Numeric.toHexString(String.valueOf(i))),
              Numeric.toHexString("0"));
      cellInputs.add(cellInput);
    }
    for (int i = 0; i < outNum; i++) {
      CellOutput cellOutput = new CellOutput(Numeric.toHexString("100"), new Script("0x0", "0x0"));
      outputs.add(cellOutput);
      witnesses.add(new Witness(Witness.EMPTY_LOCK));
      cellOutputsData.add("0x");
    }

    Transaction tx =
        new Transaction(
            Numeric.toHexString("0"),
            deps,
            Collections.emptyList(),
            cellInputs,
            outputs,
            cellOutputsData,
            witnesses);

    BigInteger transactionFee = calculateTxFee(tx, BigInteger.valueOf(1024));
    return transactionFee.toString();
  }

  private static BigInteger calculateTxFee(Transaction transaction, BigInteger feeRate) {
    int txSize = Serializer.serializeTransaction(transaction).toBytes().length;
    return Calculator.calculateTransactionFee(BigInteger.valueOf(txSize), feeRate);
  }
}
