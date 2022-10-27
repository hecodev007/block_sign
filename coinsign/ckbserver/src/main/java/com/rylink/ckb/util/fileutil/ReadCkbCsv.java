package com.rylink.ckb.util.fileutil;

import com.rylink.ckb.util.ckbutil.model.AddrInfo;
import com.rylink.ckb.util.fileutil.model.CsvCtx;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.nio.charset.Charset;
import java.util.ArrayList;
import java.util.List;




public class ReadCkbCsv {
    public static List<CsvCtx> ReadByFile(String inpath){
        List<CsvCtx> list = new ArrayList<CsvCtx>(); // 保存读取到的CSV数据
        try {
            File file = new File(inpath); // 判断文件是否存在
            if (!file.exists()) {
                System.out.println("文件不存在！");
            } else {
                System.out.println("文件存在！");
                BufferedReader reader = new BufferedReader(new FileReader(inpath)); // 读取CSV文件
                String line = null;// 循环读取每行
                while ((line = reader.readLine()) != null) {
                    String[] row = line.split(",", -1); // 分隔字符串（这里用到转义），存储到List<AddrInfo>里
                    if(row.length < 2) {
                        return list;
                    }
                    CsvCtx infos = new CsvCtx();
                    infos.setAddress(row[0]);;
                    infos.setContext(row[1]);;
                    list.add(infos);
                }
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
        return list;
    }
}
