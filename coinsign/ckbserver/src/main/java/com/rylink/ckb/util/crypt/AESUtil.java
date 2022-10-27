package com.rylink.ckb.util.crypt;

import javax.crypto.*;
import javax.crypto.spec.IvParameterSpec;
import javax.crypto.spec.SecretKeySpec;
import java.security.*;
import java.util.Arrays;
import java.util.Base64;
import java.util.UUID;

public class AESUtil {

    private static final String AES_MODEL = "AES/CFB/NoPadding";



    public static String decrypt(String dencoded_string,String key)  {
        String plain_text = "";
        try{
            byte[] encrypted_decoded_bytes = Base64.getDecoder().decode(dencoded_string);
            IvParameterSpec iv = new IvParameterSpec(Arrays.copyOf(key.getBytes(), 16));
            SecretKeySpec skeySpec = new SecretKeySpec(key.getBytes("UTF-8"), "AES");
            Cipher cipher = Cipher.getInstance(AES_MODEL);
            cipher.init(Cipher.DECRYPT_MODE, skeySpec, iv);
            plain_text = new String(cipher.doFinal(encrypted_decoded_bytes));//Returns garbage characters
            return plain_text;
        }  catch (Exception e) {
            System.err.println("Caught Exception: " + e.getMessage());
        }

        return plain_text;
    }


    public static String ecrypt(String encoded_string,String key) {

        try{
            IvParameterSpec iv = new IvParameterSpec(Arrays.copyOf(key.getBytes(), 16));
            SecretKeySpec skeySpec = new SecretKeySpec(key.getBytes("UTF-8"), "AES");
            Cipher cipher = Cipher.getInstance(AES_MODEL);
            cipher.init(Cipher.ENCRYPT_MODE, skeySpec,iv);
            byte[] resultByte = cipher.doFinal(encoded_string.getBytes());
            String resultText = Base64.getEncoder().encodeToString(resultByte);
            return resultText;
        }  catch (Exception e) {
            System.err.println("Caught Exception: " + e.getMessage());
        }

        return "";
    }


    public static String randKey(){
        String uid = UUID.randomUUID().toString();
        String key = uid.replaceAll("-","");
        return key;
    }



  public static void main(String[] args) throws Exception {
    String key = randKey();
    System.out.println(key);
    System.out.println(ecrypt("KyAevMVhccueJq9c4FiUUsXjQJudXWHHCLDeXaCMbx4JAn",key));
    System.out.println(decrypt(ecrypt("KyAevMVhccueJq9c4FiUUsXjQJudXWHHCLDeXaCMbx4JAn",key),key));

//    System.out.println(decrypt("Q+Vmn7o2U7BIsqjawFqDec+jNYisSDpGLQPCNKyIMKwExgGZECDSguR5NSfhTQ==","9YGiaxgBjLVZXbCLnl3UKOHgCBQDZoes"));
//    System.out.println(ecrypt("KyAevMVhccueJq9c4FiUUsXjQJudXWHHCLDeXaCMbx4JAn","9YGiaxgBjLVZXbCLnl3UKOHgCBQDZoes"));
  }
}
