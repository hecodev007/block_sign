package com.rylink.ckb.util.ckbutil;

import com.google.common.primitives.Bytes;
import com.google.gson.Gson;
import com.rylink.ckb.util.ckbutil.model.AddrInfo;
import com.rylink.ckb.util.ckbutil.model.MulitSignAddrInfo;
import org.bouncycastle.jce.ECNamedCurveTable;
import org.bouncycastle.jce.provider.BouncyCastleProvider;
import org.bouncycastle.jce.spec.ECNamedCurveParameterSpec;
import org.nervos.ckb.address.AddressUtils;
import org.nervos.ckb.address.CodeHashType;
import org.nervos.ckb.address.Network;
import org.nervos.ckb.crypto.Hash;
import org.nervos.ckb.crypto.secp256k1.ECKeyPair;
import org.nervos.ckb.utils.Numeric;

import java.io.IOException;
import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.security.SecureRandom;
import java.security.Security;
import java.util.ArrayList;
import java.util.List;

public class AddrUtil {

  private static final Network testNetwork = Network.TESTNET;
  private static final Network mainNetwork = Network.MAINNET;

  public static AddrInfo paseAddr(String privateKey) {
    String publicKey = ECKeyPair.publicKeyFromPrivate(privateKey);
    // 公私钥绑定
    AddrInfo addrInfo = new AddrInfo();
    addrInfo.setPrivKey(privateKey);
    addrInfo.setPubkey(publicKey);

    // 测试网地址
    AddressUtils testUtil = new AddressUtils(testNetwork);
    String testAddress = testUtil.generateFromPublicKey(publicKey);

    // 主网地址
    AddressUtils mainUtil = new AddressUtils(mainNetwork);
    String mainAddress = mainUtil.generateFromPublicKey(publicKey);

    // 地址绑定
    addrInfo.setTestAddress(testAddress);
    addrInfo.setMainAddress(mainAddress);

    // lock-args
    if (!AddressUtils.parse(mainAddress).equals(AddressUtils.parse(testAddress))) {
      return null;
    }
    addrInfo.setLockArg(AddressUtils.parse(mainAddress));
    return addrInfo;
  }

  // 生成地址
  public static AddrInfo createAddr() throws Exception {

    // 公私钥生成
    KeyPair keyPair = newKeyPair();
    ECKeyPair ecKeyPair = ECKeyPair.createWithKeyPair(keyPair);
    String privateKey = ecKeyPair.getPrivateKey().toString(16);
    //        String publicKey = Sign.publicKeyFromPrivate(ecKeyPair.getPrivateKey()).toString(16);
    String publicKey = ECKeyPair.publicKeyFromPrivate(privateKey);

    // 公私钥绑定
    AddrInfo addrInfo = new AddrInfo();
    addrInfo.setPrivKey(privateKey);
    addrInfo.setPubkey(publicKey);

    // 测试网地址
    AddressUtils testUtil = new AddressUtils(testNetwork);
    String testAddress = testUtil.generateFromPublicKey(publicKey);

    // 主网地址
    AddressUtils mainUtil = new AddressUtils(mainNetwork);
    String mainAddress = mainUtil.generateFromPublicKey(publicKey);

    // 地址绑定
    addrInfo.setTestAddress(testAddress);
    addrInfo.setMainAddress(mainAddress);

    // lock-args
    if (!AddressUtils.parse(mainAddress).equals(AddressUtils.parse(testAddress))) {
      return null;
    }
    addrInfo.setLockArg(AddressUtils.parse(mainAddress));
    return addrInfo;
  }

  // fixme 测试阶段，不可用
  public static MulitSignAddrInfo createMultiSigAddr(int requireN, int threshold, int publicKeyN)
      throws Exception {
    if (requireN < 0 || requireN > 255) {
      throw new IOException("requireN should be less than 256");
    }
    if (threshold < 0 || threshold > 255) {
      throw new IOException("threshold should be less than 256");
    }
    if (publicKeyN < 0 || publicKeyN > 255) {
      throw new IOException("Public key number must be less than 256");
    }
    List<AddrInfo> addressInfos = new ArrayList<AddrInfo>();
    for (int i = 0; i < publicKeyN; i++) {
      // 公私钥绑定
      AddrInfo addressInfo = createAddr();
      addressInfos.add(addressInfo);
    }
    StringBuilder multiSigBuffer = new StringBuilder();
    List<Byte> bytes = new ArrayList<>();
    bytes.addAll(Numeric.intToBytes(0));
    bytes.addAll(Numeric.intToBytes(requireN));
    bytes.addAll(Numeric.intToBytes(threshold));
    bytes.addAll(Numeric.intToBytes(addressInfos.size()));
    multiSigBuffer.append(Numeric.toHexStringNoPrefix(Bytes.toArray(bytes)));
    for (AddrInfo addrInfo : addressInfos) {
      multiSigBuffer.append(Hash.blake160(addrInfo.getPubkey()));
    }
    String multisigScriptHash = multiSigBuffer.toString();

    // 测试网地址
    AddressUtils testUtil = new AddressUtils(testNetwork, CodeHashType.MULTISIG);
    // 主网地址
    AddressUtils mainUtil = new AddressUtils(mainNetwork, CodeHashType.MULTISIG);

    MulitSignAddrInfo mulitSignAddrInfo = new MulitSignAddrInfo();
    mulitSignAddrInfo.setAddressInfo(addressInfos);
    mulitSignAddrInfo.setTestAddress(testUtil.generate(Hash.blake160(multisigScriptHash)));
    mulitSignAddrInfo.setMainAddress(mainUtil.generate(Hash.blake160(multisigScriptHash)));
    return mulitSignAddrInfo;
  }

  public static MulitSignAddrInfo createMultiSigAddrByPublicKeys(
      int requireN, int threshold, List<String> publicKeys) throws Exception {
    if (requireN < 0 || requireN > 255) {
      throw new IOException("requireN should be less than 256");
    }
    if (threshold < 0 || threshold > 255) {
      throw new IOException("threshold should be less than 256");
    }
    if (publicKeys.size() == 0 || publicKeys.size() > 255) {
      throw new IOException("Public key number must be less than 256");
    }
    List<AddrInfo> addressInfos = new ArrayList<AddrInfo>();
    for (int i = 0; i < publicKeys.size(); i++) {
      // 公私钥绑定
      AddrInfo addressInfo = new AddrInfo();
      addressInfo.setPubkey(publicKeys.get(i));
      addressInfos.add(addressInfo);
    }
    StringBuilder multiSigBuffer = new StringBuilder();
    List<Byte> bytes = new ArrayList<>();
    bytes.addAll(Numeric.intToBytes(0));
    bytes.addAll(Numeric.intToBytes(requireN));
    bytes.addAll(Numeric.intToBytes(threshold));
    bytes.addAll(Numeric.intToBytes(publicKeys.size()));
    multiSigBuffer.append(Numeric.toHexStringNoPrefix(Bytes.toArray(bytes)));
    for (AddrInfo addrInfo : addressInfos) {
      multiSigBuffer.append(Hash.blake160(addrInfo.getPubkey()));
    }
    String multisigScriptHash = multiSigBuffer.toString();

    // 测试网地址
    AddressUtils testUtil = new AddressUtils(testNetwork, CodeHashType.MULTISIG);
    // 主网地址
    AddressUtils mainUtil = new AddressUtils(mainNetwork, CodeHashType.MULTISIG);

    MulitSignAddrInfo mulitSignAddrInfo = new MulitSignAddrInfo();
    mulitSignAddrInfo.setAddressInfo(addressInfos);
    mulitSignAddrInfo.setTestAddress(testUtil.generate(Hash.blake160(multisigScriptHash)));
    mulitSignAddrInfo.setMainAddress(mainUtil.generate(Hash.blake160(multisigScriptHash)));

    if (AddressUtils.parse(mulitSignAddrInfo.getMainAddress())
        .equals(AddressUtils.parse(mulitSignAddrInfo.getTestAddress()))) {
      mulitSignAddrInfo.setArgs(AddressUtils.parse(mulitSignAddrInfo.getMainAddress()));
    } else {
      throw new IOException("args error");
    }
    return mulitSignAddrInfo;
  }

  // 获取lock-args
  public static String getLockArg(String address) {
    return AddressUtils.parse(address);
  }

  /**
   * 创建新的密钥对
   *
   * @return
   * @throws Exception
   */
  private static KeyPair newKeyPair() throws Exception {
    // 注册 BC Provider
    Security.addProvider(new BouncyCastleProvider());
    // 创建椭圆曲线算法的密钥对生成器，算法为 ECDSA
    KeyPairGenerator g = KeyPairGenerator.getInstance("ECDSA", BouncyCastleProvider.PROVIDER_NAME);
    // 椭圆曲线（EC）域参数设定
    //        ECParameterSpec ecSpec = ECNamedCurveTable.getParameterSpec("secp256k1");
    ECNamedCurveParameterSpec ecSpec = ECNamedCurveTable.getParameterSpec("secp256k1");
    g.initialize(ecSpec, new SecureRandom());
    return g.generateKeyPair();
  }

  public static void main(String[] args) throws Exception {
    MulitSignAddrInfo m = createMultiSigAddr(0, 2, 3);
    Gson gson = new Gson();
    System.out.println(gson.toJson(m));
  }
}
