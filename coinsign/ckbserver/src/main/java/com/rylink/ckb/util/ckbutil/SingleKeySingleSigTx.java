package com.rylink.ckb.util.ckbutil;

import com.google.gson.FieldNamingPolicy;
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.rylink.ckb.util.ckbutil.model.TxInput;
import com.rylink.ckb.util.ckbutil.model.TxOutput;
import com.rylink.ckb.util.ckbutil.model.TxTpl;
import org.nervos.ckb.address.AddressUtils;
import org.nervos.ckb.address.CodeHashType;
import org.nervos.ckb.address.Network;
import org.nervos.ckb.system.type.SystemScriptCell;
import org.nervos.ckb.type.OutPoint;
import org.nervos.ckb.type.Script;
import org.nervos.ckb.type.Witness;
import org.nervos.ckb.type.cell.CellDep;
import org.nervos.ckb.type.cell.CellInput;
import org.nervos.ckb.type.cell.CellOutput;
import org.nervos.ckb.type.transaction.Transaction;
import org.nervos.ckb.utils.Numeric;

import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

// 多对多，但是如果地址类型是多签地址,需要特殊处理
public class SingleKeySingleSigTx {

  private static AddressUtils addressUtils; // 地址工具类
  private static final BigDecimal UnitCKB; // 基本单元 1 = 100000000
  private static BigDecimal txFee; // 默认手续费

  static {
    addressUtils = new AddressUtils(Network.TESTNET, CodeHashType.BLAKE160);
    UnitCKB = new BigDecimal("100000000");
    txFee = new BigDecimal("0.01");
  }

  // 这个只能一对多
  public static String singleSigTx(TxTpl tpl) {
    if (tpl == null
        || tpl.getInputs().size() != 1
        || tpl.getOutputs().size() == 0
        || tpl.getTxFee().compareTo(BigDecimal.ZERO) == 0
        || tpl.getTxFee().compareTo(BigDecimal.ZERO) == -1) {
      System.out.println("sigTx参数异常");
      return "";
    }

    BigDecimal changeAmount = BigDecimal.ZERO;
    BigDecimal toAmount = BigDecimal.ZERO;
    BigDecimal fromAmount = BigDecimal.ZERO;

    // script来自：get_block_by_number 高度0 参数为：0x0
    // 获取 transactions 数组的下标0 中 outputs的下标1的type进行Blake2b算法生成txHash,    index 固定为0x0

    // 多签script来自来自 get_block_by_number 高度1 参数为：0x1
    // 获取 transactions 数组的下标0 中 outputs的下标4的type进行Blake2b算法生成txHash,    index 固定为0x1
    Script script =
        new Script(
            "0x00000000000000000000000000000000000000000000000000545950455f4944",
            "0x8536c9d5d908bd89fc70099e4284870708b6632356aad98734fcf43f6f71c304",
            "type");
    SystemScriptCell systemScriptCell =
        new SystemScriptCell(
            script.computeHash(),
            new OutPoint(
                "0xbd864a269201d7052d4eb3f753f49f7c68b8edc386afc8bb6ef3e15a05facca2", "0x0"));

    // 系统cell引用
    //
    List<CellDep> deps = new ArrayList<CellDep>();
    deps.add(
        new CellDep(
            new OutPoint(
                "0xbd864a269201d7052d4eb3f753f49f7c68b8edc386afc8bb6ef3e15a05facca2",
                Numeric.toHexString("0")),
            CellDep.DEP_GROUP));

    // 封装in输入
    List<CellInput> inputs = new ArrayList<CellInput>();
    for (TxInput in : tpl.getInputs()) {
      if (in.getPrivateKey() == ""
          || in.getAmount().compareTo(BigDecimal.ZERO) == 0
          || in.getAmount().compareTo(BigDecimal.ZERO) == -1) {
        System.out.println(
            String.format(
                "txid:%s,address:%s,index:%d,amount:%s,error params",
                in.getTxid(), in.getAddress(), in.getIndex(), in.getAmount().toString()));
        return "";
      }
      inputs.add(
          new CellInput(
              new OutPoint(in.getTxid(), Numeric.toHexString(String.valueOf(in.getIndex()))),
              Numeric.toHexString("0")));
      fromAmount = fromAmount.add(in.getAmount());
    }

    // 封装out输出
    List<CellOutput> outputs = new ArrayList<CellOutput>();
    for (TxOutput outs : tpl.getOutputs()) {
      String blake160 = AddressUtils.parse(outs.getAddress());
      outputs.add(
          new CellOutput(
              Numeric.toHexString(outs.getAmount().multiply(UnitCKB).toBigInteger().toString()),
              new Script(systemScriptCell.cellHash, blake160, Script.TYPE)));

      toAmount = toAmount.add(outs.getAmount());
    }

    if (0 <= tpl.getTxFee().multiply(UnitCKB).toBigInteger().intValue()) {
      // 自动计算手续费
    } else {
      txFee = tpl.getTxFee();
    }

    int compareResult = fromAmount.subtract(toAmount).compareTo(txFee);
    if (compareResult == -1) {
      System.out.println(
          String.format(
              "fromAmount:%s,toAmount:%s,fee:%s,error amount",
              fromAmount.toString(), toAmount.toString(), txFee.toString()));
      return "";
    } else if (compareResult == 1) {
      // 需要找零
      String blake160 = AddressUtils.parse(tpl.getChangeAddr());
      changeAmount = fromAmount.subtract(toAmount).subtract(txFee);
      System.out.println("找零金额：" + changeAmount.toString());
      outputs.add(
          new CellOutput(
              Numeric.toHexString(changeAmount.multiply(UnitCKB).toBigInteger().toString()),
              new Script(systemScriptCell.cellHash, blake160, Script.TYPE)));
    }

    List<Witness> witnesses = Collections.singletonList(new Witness());
    List<String> cellOutputsData = new ArrayList<>();
    for (int i = 0; i < outputs.size(); i++) {
      cellOutputsData.add("0x");
    }

    Transaction tx =
        new Transaction(
            Numeric.toHexString("0"),
            deps,
            Collections.emptyList(),
            inputs,
            outputs,
            cellOutputsData,
            witnesses);

    BigInteger privateKey = Numeric.toBigInt(tpl.getInputs().get(0).getPrivateKey());
    Transaction signedTx = tx.sign(privateKey);
    Gson gson =
        (new GsonBuilder())
            .setFieldNamingPolicy(FieldNamingPolicy.LOWER_CASE_WITH_UNDERSCORES)
            .create();
    System.out.println(gson.toJson(signedTx));

    System.out.println("txid computeHash:" + signedTx.computeHash());

    return gson.toJson(signedTx);
  }
}
