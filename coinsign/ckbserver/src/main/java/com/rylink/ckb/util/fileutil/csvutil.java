package com.rylink.ckb.util.fileutil;

import lombok.extern.slf4j.Slf4j;

import java.io.*;
import java.util.Arrays;
import java.util.List;

@Slf4j
public class csvutil {


    //读取CSV文件
    public void readCsvFile(String fileName) throws IOException {
        BufferedReader  bufferedReader=null;
        try{
            bufferedReader=new BufferedReader(new FileReader(fileName));
            String line=null;
            while(null!=(line=bufferedReader.readLine())){
                String[] lines = line.split(",");
                log.info("这就是文件的内容"+ Arrays.toString(lines));
            }
        }catch (IOException e){
            throw  new IOException(e);
        }finally {
            if(bufferedReader!=null){
                try {
                    bufferedReader.close();
                } catch (IOException e) {
                    log.error("输入流关流出现异常",e);
                }
            }
        }
    }

    //生产CSV文件
    public static void writerCsvFile(String fileName, List<String[]> list) throws IOException {
        BufferedWriter bufferedWriter=null;
        try{
            bufferedWriter= new BufferedWriter(new OutputStreamWriter(new FileOutputStream(fileName),"UTF8"));
            for (String[] s:list) {
                for (int i=0;i<s.length;i++){
                    bufferedWriter.write(s[i]);
                    bufferedWriter.write(",");
                }
                bufferedWriter.newLine();//记得换行
            }
        }catch (IOException e){
            throw  new IOException(e);
        }finally {
            if(bufferedWriter!=null){
                try {
                    bufferedWriter.close();
                } catch (IOException e2) {
                    log.error("输出流关流出现异常",e2);
                }
            }
        }
    }


}
