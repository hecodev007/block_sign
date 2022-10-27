package com.rylink.ckb.util.springutil;

import com.rylink.ckb.util.crypt.AESUtil;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import javax.annotation.PostConstruct;
import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

@Component
@Slf4j
public class DicMap {
  private static HashMap<String, String> addressMap = new HashMap<>();
  private static HashMap<String, String> aMap = new HashMap<>();
  private static HashMap<String, String> bMap = new HashMap<>();

  @PostConstruct
  public void doConstruct() throws Exception {
    queryDic();
  }
  // 把字典信息放到map中
  public static void queryDic() throws IOException {
    String relativelyPath = System.getProperty("user.dir");
    log.info("relativelyPath:" + relativelyPath);
    File file = new File(relativelyPath + "/csv");
    getAFileByDirectory(file);
    getBFileByDirectory(file);
    if (aMap.size() != bMap.size()) {
      log.error("地址数量不一致");
      return;
    }
    // 私钥解密
    readAddressKey();
    log.info("地址数量：" + addressMap.size());
  }
  // 获取字典信息的内容
  public static String getDicVal(String key) {
    if (addressMap.containsKey(key)) {
      return addressMap.get(key);
    } else {
      return "";
    }
  }

  // 递归遍历获取A文件信息
  private static void getAFileByDirectory(File file) throws IOException {

    File flist[] = file.listFiles();
    if (flist == null || flist.length == 0) {
      return;
    }
    for (File f : flist) {
      if (f.isDirectory()) {
        // 这里将列出所有的文件夹
        //        System.out.println("Dir==>" + f.getAbsolutePath());
        getAFileByDirectory(f);
      } else {
        // 这里将列出所有的文件
        //        System.out.println("fileA==>" + f.getAbsolutePath());
        if (f.getAbsolutePath().contains("ckb_a_usb_")) {
          BufferedReader reader =
              new BufferedReader(new FileReader(f.getAbsolutePath())); // 读取CSV文件
          String line = null; // 循环读取每行
          while ((line = reader.readLine()) != null) {
            String[] row = line.split(",", -1); // 分隔字符串（这里用到转义），存储到List里
            if (row.length < 2) {
              log.error("文件异常!A文件：" + f.getAbsolutePath());
              return;
            }
            aMap.put(row[0], row[1]);
          }
        }
      }
    }
  }

  // 递归遍历获取文件信息
  private static void getBFileByDirectory(File file) throws IOException {

    File flist[] = file.listFiles();
    if (flist == null || flist.length == 0) {
      return;
    }
    for (File f : flist) {
      if (f.isDirectory()) {
        // 这里将列出所有的文件夹
        //        System.out.println("Dir==>" + f.getAbsolutePath());
        getBFileByDirectory(f);
      } else {
        // 这里将列出所有的文件
        //        System.out.println("fileA==>" + f.getAbsolutePath());
        if (f.getAbsolutePath().contains("ckb_b_usb_")) {
          BufferedReader reader =
              new BufferedReader(new FileReader(f.getAbsolutePath())); // 读取CSV文件
          String line = null; // 循环读取每行
          while ((line = reader.readLine()) != null) {
            String[] row = line.split(",", -1); // 分隔字符串（这里用到转义），存储到List里
            if (row.length < 2) {
              log.error("文件异常!B文件：" + f.getAbsolutePath());
              return;
            }
            bMap.put(row[0], row[1]);
          }
        }
      }
    }
  }

  private static void readAddressKey() throws IOException {
    for (Map.Entry<String, String> entry : aMap.entrySet()) {
      //            System.out.println("Key = " + entry.getKey() + ", Value = " + entry.getValue());
      if (!bMap.containsKey(entry.getKey())) {
        log.error(entry.getKey() + "，丢失私钥");
      } else {
        // 解密
        String privKey = AESUtil.decrypt(entry.getValue(), bMap.get(entry.getKey()));
        addressMap.put(entry.getKey(), privKey);
        //        System.out.println(entry.getKey() + ":" + privKey);
        //        log.error(entry.getKey());
      }
    }
  }
}
