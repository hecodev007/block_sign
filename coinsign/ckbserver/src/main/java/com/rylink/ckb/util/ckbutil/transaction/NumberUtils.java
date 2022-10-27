package com.rylink.ckb.util.ckbutil.transaction;

import java.util.ArrayList;
import java.util.List;

/** Copyright © 2019 Nervos Foundation. All rights reserved. */
public class NumberUtils {

  public static List<Integer> regionToList(int start, int length) {
    List<Integer> integers = new ArrayList<>();
    for (int i = start; i < (start + length); i++) {
      integers.add(i);
    }
    return integers;
  }
}
