package com.rylink.ckb.util.fileutil;

import java.io.File;
import java.io.IOException;

public class files {
    // 判断文件是否存在
    public static boolean fileExists(File file) {

        if (file.exists()) {
            System.out.println("file exists");
            return true;

        } else {
//            System.out.println("file not exists, create it ...");
//            try {
//                file.createNewFile();
//            } catch (IOException e) {
//                // TODO Auto-generated catch block
//                e.printStackTrace();
//            }
            return false;
        }

    }

    // 判断文件夹是否存在
    public static boolean dirExists(File file) {

//        if (file.exists()) {
//            if (file.isDirectory()) {
//                System.out.println("dir exists");
//            } else {
//                System.out.println("the same name file exists, can not create dir");
//            }
//        } else {
//            System.out.println("dir not exists, create it ...");
//            file.mkdir();
//        }
        if (file.exists()) {
            if (file.isDirectory()) {
                return true;
            }
        }
        return false;

    }

}
