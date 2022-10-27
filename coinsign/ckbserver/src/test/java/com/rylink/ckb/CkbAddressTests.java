package com.rylink.ckb;

import com.google.gson.Gson;
import com.rylink.ckb.util.ckbutil.AddrUtil;
import com.rylink.ckb.util.ckbutil.model.AddrInfo;
import com.rylink.ckb.util.ckbutil.model.MulitSignAddrInfo;
import org.junit.Test;
import org.springframework.boot.test.context.SpringBootTest;

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
public class CkbAddressTests {
  @Test
  public void testAddress() {
    try {
      List<String> publicKeys =
          Arrays.asList(
              "32edb83018b57ddeb9bcc7287c5cc5da57e6e0289d31c9e98cb361e88678d6288",
              "33aeb3fdbfaac72e9e34c55884a401ee87115302c146dd9e314677d826375dc8f",
              "29a685b8206550ea1b600e347f18fd6115bffe582089d3567bec7eba57d04df01");
      MulitSignAddrInfo info = AddrUtil.createMultiSigAddrByPublicKeys(0, 2, publicKeys);
      System.out.println(new Gson().toJson(info));

    } catch (Exception e) {
      e.printStackTrace();
    }
  }

  @Test
  public void testAddressRylink() {
    try {
      List<String> publicKeys =
          Arrays.asList(
              "32edb83018b57ddeb9bcc7287c5cc5da57e6e0289d31c9e98cb361e88678d6288",
              "33aeb3fdbfaac72e9e34c55884a401ee87115302c146dd9e314677d826375dc8f",
              "29a685b8206550ea1b600e347f18fd6115bffe582089d3567bec7eba57d04df01");
      MulitSignAddrInfo info = AddrUtil.createMultiSigAddrByPublicKeys(0, 2, publicKeys);
      System.out.println(new Gson().toJson(info));

    } catch (Exception e) {
      e.printStackTrace();
    }
  }

  //
  // {"addressInfo":[{"mainAddress":"ckb1qyqrvr760ycwyx7ujwttaalpa3ysdap5yn0sklaja2","testAddress":"ckt1qyqrvr760ycwyx7ujwttaalpa3ysdap5yn0st6rd3k","privKey":"750b3bb6b2c7925347e86a192c59043828131d7635853cba5be91cc956fadc5","pubkey":"25dedc6531ce7d7c8c8d7f4a5fc317b71e0d494363fb3f023cd6df06827f714ea","lockArg":"360fda7930e21bdc9396bef7e1ec4906f43424df"},{"mainAddress":"ckb1qyqyy3j5ssu6kxhrjadz0tcazsfxqfmag2asclr9gl","testAddress":"ckt1qyqyy3j5ssu6kxhrjadz0tcazsfxqfmag2as96a6yr","privKey":"be19e74effcef0de6e553c0d39233fedd9cb3041e59d1dd5cec9640e0b2bd6f0","pubkey":"26ab3be4495579c66fe2d74497d911f8a6bfc5b9835c98f41773f6480e586bc22","lockArg":"4246548439ab1ae3975a27af1d141260277d42bb"},{"mainAddress":"ckb1qyqgyv0a02jq0j0ra5nkpanv6aluftea982sevn9ka","testAddress":"ckt1qyqgyv0a02jq0j0ra5nkpanv6aluftea982syfd66p","privKey":"23f9a5ea00d829c962d8b939d462db2eab3b4d5a848e8cf74fc91213fce1d2c","pubkey":"22ad3fa3395f34f4ffb5fa98311433a3baf8696cb34eb865a3dde3881ec2b6f82","lockArg":"8231fd7aa407c9e3ed2760f66cd77fc4af3d29d5"}],"mainAddress":"ckb1qyqc4jax2vuk3xvnte2fpqwpny5dfzr3n3msf9khpx","testAddress":"ckt1qyqc4jax2vuk3xvnte2fpqwpny5dfzr3n3ms5qggd6"}
  //
  // {"addressInfo":[{"pubkey":"25dedc6531ce7d7c8c8d7f4a5fc317b71e0d494363fb3f023cd6df06827f714ea"},{"pubkey":"26ab3be4495579c66fe2d74497d911f8a6bfc5b9835c98f41773f6480e586bc22"},{"pubkey":"22ad3fa3395f34f4ffb5fa98311433a3baf8696cb34eb865a3dde3881ec2b6f82"}],"mainAddress":"ckb1qyqc4jax2vuk3xvnte2fpqwpny5dfzr3n3msf9khpx","testAddress":"ckt1qyqc4jax2vuk3xvnte2fpqwpny5dfzr3n3ms5qggd6","args":"8acba653396899935e549081c19928d488719c77"}

  @Test
  public void testAddressCreate() {
    try {

      MulitSignAddrInfo info = AddrUtil.createMultiSigAddr(0, 2, 3);
      List<String> publicKeys = new ArrayList<>();
      for (AddrInfo addrInfo : info.getAddressInfo()) {
        publicKeys.add(addrInfo.getPubkey());
      }
      System.out.println(new Gson().toJson(info));
      MulitSignAddrInfo info2 = AddrUtil.createMultiSigAddrByPublicKeys(0, 2, publicKeys);
      System.out.println(new Gson().toJson(info2));

    } catch (Exception e) {
      e.printStackTrace();
    }
  }

  @Test
  public void testAddressUtil() {}
}
