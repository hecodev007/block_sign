package com.rylink.ckb;

import com.google.gson.FieldNamingPolicy;
import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.rylink.ckb.util.ckbutil.MultiKeySingleSigTx;
import com.rylink.ckb.util.ckbutil.model.TxInput;
import com.rylink.ckb.util.ckbutil.model.TxOutput;
import com.rylink.ckb.util.ckbutil.model.TxTpl;
import org.junit.Test;
import org.nervos.ckb.system.type.SystemScriptCell;
import org.nervos.ckb.type.OutPoint;
import org.nervos.ckb.utils.Numeric;
import org.springframework.boot.test.context.SpringBootTest;

import java.io.IOException;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

// @BeforeClass 在所有测试方法前执行一次，一般在其中写上整体初始化的代码
//
// @AfterClass 在所有测试方法后执行一次，一般在其中写上销毁和释放资源的代码
//
// @Before 在每个测试方法前执行，一般用来初始化方法（比如我们在测试别的方法时，类中与其他测试方法共享的值已经被改变，为了保证测试结果的有效性，我们会在@Before注解的方法中重置数据）
//
// @After 在每个测试方法后执行，在方法执行完成后要做的事情
//
// @Test(timeout = 1000) 测试方法执行超过1000毫秒后算超时，测试将失败
//
// @Test(expected = Exception.class) 测试方法期望得到的异常类，如果方法执行没有抛出指定的异常，则测试失败
//
// @Ignore(“not ready yet”) 执行测试时将忽略掉此方法，如果用于修饰类，则忽略整个类
//
// @Test 编写一般测试用例

// @RunWith(SpringRunner.class)
@SpringBootTest
// @SpringBootTest(classes={Application.class})// 指定启动类
public class CkbSignTests {

  private static SystemScriptCell systemScriptCell;
  private static SystemScriptCell systemMultiSigCell;

  static {
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
    systemScriptCell =
        new SystemScriptCell(
            "0x9bd7e06f3ecf4be0f2fcd2188b23f1b9fcc88e5d4b65a8637b17723bbda3cce8",
            new OutPoint(
                "0x71a7ba8fc96349fea0ed3a5c47992e3b4084b031a42264a018e0072e8172e46c",
                Numeric.toHexStringWithPrefix(BigInteger.ZERO)));

    //    Script muiltScript =
    //        new Script(
    //            "0x00000000000000000000000000000000000000000000000000545950455f4944",
    //            "0xd813c1b15bd79c8321ad7f5819e5d9f659a1042b72e64659a2c092be68ea9758",
    //            "type");
    //    String muiltCellHash = muiltScript.computeHash();
    //    System.out.println(muiltCellHash + ":muiltCellHash");
    systemMultiSigCell =
        new SystemScriptCell(
            "0x5c5069eb0857efc65e1bca0c07df34c31663b3622fd3876c876320fc9634e2a8",
            new OutPoint(
                "0x71a7ba8fc96349fea0ed3a5c47992e3b4084b031a42264a018e0072e8172e46c",
                Numeric.toHexStringWithPrefix(BigInteger.ONE)));
  }

  // 一对多测试
  // 如果是多对多则报错
  @Test
  public void testSingleSign() throws IOException {
    //    System.out.println(Numeric.toHexStringWithPrefix(BigInteger.valueOf(3)));
    //    System.out.println(MultiKeySingleSigTx.getTxFee(3, 2));

    //    Script lockScript =
    //        Utils.generateLockScriptWithAddress(
    //
    // "ckt1q3mujwcxx26mdsl0jgk9kl82yz8mpf7yy7sn658p85l7ltghurzepwuh38u4qrs2uk9tqpuqsc6gqvc2lqgds5avaq5",
    //            "0x77c93b0632b5b6c3ef922c5b7cea208fb0a7c427a13d50e13d3fefad17e0c590",
    //            CodeHashType.BLAKE160);

    //    AddressUtils addressUtils = new AddressUtils(Network.TESTNET);
    //    System.out.println(
    //        addressUtils.getArgsFromAddress("ckt1qyqfrcdxl8d4ltlz2tpqh5e5mekqcd9egv5q73t8rj"));
    //    System.out.println(
    //        addressUtils.getArgsFromAddress("ckt1qyq97eaag5nj0lx43upcwexs6gnnjlye00ws6z7azz"));
    //
    //    System.out.println(
    //        addressUtils.getArgsFromAddress(
    //
    // "ckt1q3mujwcxx26mdsl0jgk9kl82yz8mpf7yy7sn658p85l7ltghurzepy0p5muakha0uffvyz7nxn0xcrp5h9pjszywkpv"));
    //    System.out.println(
    //        addressUtils.getArgsFromAddress(
    //
    // "ckt1q3mujwcxx26mdsl0jgk9kl82yz8mpf7yy7sn658p85l7ltghurzeqhm8h4zjwflu6k8s8pmy6rfzwwtun9aa65svte3"));

    //    List<TxInput> inputs =
    //        Arrays.asList(
    //            new TxInput(
    //                "c6dd6ebc66484cb6a0d4f24db37c8778b913efaaac1e53acf0fc1da09943c8b1",
    //                "ckt1qyqth9ufl9gquzh932cq0qyxxjqrxzhczrvqs55dc2",
    //                "0x5fd5155ba542968a43fa8ff94555a04a676d6f364ba76d820a78985a87dccc0b",
    //                4,
    //                new BigDecimal("1000")));

    //    List<TxInput> inputs =
    //        Arrays.asList(
    //            new TxInput(
    //                "c6dd6ebc66484cb6a0d4f24db37c8778b913efaaac1e53acf0fc1da09943c8b1",
    //                "ckt1qyqth9ufl9gquzh932cq0qyxxjqrxzhczrvqs55dc2",
    //                "0xdac430c68a0c80270a34b4b82b0a2f0545a0a8838f91a943cf65b41832e979b4",
    //                9,
    //                new BigDecimal("499.9999")));

    List<TxInput> inputs =
        Arrays.asList(
            new TxInput(
                "b03a4772d46d9de88a8bbb3930b782436f84a7d12a7de3b653f20518938fe789",
                "ckb1qyqt88aehpf7xc8wnpf3jvmseg4t92k72yjs73rysa",
                "0x2be535a3e4448bbfaf361cfcb75155dfcf2f5dbb4c7b055c4158329cb0a28ccc",
                0,
                new BigDecimal("111")));

    List<TxOutput> outputs =
        Arrays.asList(
            new TxOutput("ckb1qyq2xc832laqq9kmg93an6weh63h80r32s7s2wafvr", new BigDecimal("110")));

    TxTpl tpl =
        new TxTpl(
            inputs,
            outputs,
            new BigDecimal("1"),
            "ckb1qyqt88aehpf7xc8wnpf3jvmseg4t92k72yjs73rysa",
            systemScriptCell,
            systemMultiSigCell);
    //    SingleKeySingleSigTx.singleSigTx(tpl);
    Gson gson =
        (new GsonBuilder())
            .setFieldNamingPolicy(FieldNamingPolicy.LOWER_CASE_WITH_UNDERSCORES)
            .create();
    System.out.println(gson.toJson(MultiKeySingleSigTx.multiKeySingleSigTx(tpl)));

    //    // error
    //    inputs =
    //        Arrays.asList(
    //            new TxInput(
    //                "0x55f3366cd3b622ede145faac6073d8f952a1da01d39a79616efdb99ef75b2bcf",
    //                "ckt1qyqfrcdxl8d4ltlz2tpqh5e5mekqcd9egv5q73t8rj",
    //                "0x1e488e0e8fd1ed0fb909aba15ae49919036b9d912fd0980012bb5e1e1a392696",
    //                0,
    //                new BigDecimal("800")),
    //            new TxInput(
    //                "0x55f3366cd3b622ede145faac6073d8f952a1da01d39a79616efdb99ef75b2bcf",
    //                "ckt1qyqfrcdxl8d4ltlz2tpqh5e5mekqcd9egv5q73t8rj",
    //                "0x1e488e0e8fd1ed0fb909aba15ae49919036b9d912fd0980012bb5e1e1a392696",
    //                0,
    //                new BigDecimal("800")));
    //    tpl.setInputs(inputs);
    //    SingleKeySingleSigTx.singleSigTx(tpl);
  }

  // 如果是多对多则报错
  @Test
  public void testSingleSign2() throws IOException {

    BigDecimal avg = new BigDecimal("62");
    BigDecimal fee = new BigDecimal("0.0001");
    String toaddr = "ckb1qyq2xc832laqq9kmg93an6weh63h80r32s7s2wafvr";
    String changeAddr = "ckb1qyq2xc832laqq9kmg93an6weh63h80r32s7s2wafvr";

    if (avg.compareTo(BigDecimal.valueOf(62)) < 0) {
      System.out.println("error avg amout,less 62");
      return;
    }

    List<TxInput> inputs = new ArrayList<>();
    inputs.add(
        new TxInput(
            "b03a4772d46d9de88a8bbb3930b782436f84a7d12a7de3b653f20518938fe789",
            "ckb1qyqt88aehpf7xc8wnpf3jvmseg4t92k72yjs73rysa",
            "0x2be535a3e4448bbfaf361cfcb75155dfcf2f5dbb4c7b055c4158329cb0a28ccc",
            0,
            new BigDecimal("111.0000")));

    BigDecimal fromAmout = BigDecimal.ZERO;
    BigDecimal tmp = BigDecimal.ZERO;
    for (TxInput in : inputs) {
      fromAmout = fromAmout.add(in.getAmount());
    }
    fromAmout = fromAmout.subtract(fee);

    if (fromAmout.compareTo(BigDecimal.ZERO) < 0) {
      System.out.println("总额少于平均值");
      return;
    }
    List<TxOutput> outputs = new ArrayList<>();
    while (true) {
      tmp = tmp.add(avg);
      if (tmp.compareTo(fromAmout) == -1) {
        outputs.add(new TxOutput(toaddr, avg));

        if (fromAmout.subtract(tmp).compareTo(BigDecimal.valueOf(62)) < 0) {
          // 因此附加在上一笔
          int index = outputs.size() - 1;
          BigDecimal toAm = outputs.get(index).getAmount();
          String toAddr = outputs.get(index).getAddress();
          outputs.get(index).setAmount(toAm.add(fromAmout.subtract(tmp)));
          outputs.get(index).setAddress(toAddr);
          System.out.println(fromAmout.subtract(tmp));
          break;
        }
      } else {
        break;
      }
    }

    if (outputs.size() == 0) {
      System.out.println("error out size");
      return;
    }
    System.out.println("new Gson().toJson(outputs):");
    System.out.println(new Gson().toJson(outputs));
    TxTpl tpl = new TxTpl(inputs, outputs, fee, changeAddr, systemScriptCell, systemMultiSigCell);
    //    SingleKeySingleSigTx.singleSigTx(tpl);
    Gson gson =
        (new GsonBuilder())
            .setFieldNamingPolicy(FieldNamingPolicy.LOWER_CASE_WITH_UNDERSCORES)
            .create();
    System.out.println(gson.toJson(MultiKeySingleSigTx.multiKeySingleSigTx(tpl)));
  }
}
