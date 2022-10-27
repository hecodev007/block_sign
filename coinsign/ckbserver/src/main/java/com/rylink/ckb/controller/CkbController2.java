// package com.rylink.ckb.controller;
//
// import com.google.gson.Gson;
// import com.rylink.ckb.model.bo.CreateAddrBO;
// import com.rylink.ckb.model.bo.TxTplBO;
// import com.rylink.ckb.model.vo.AddressVO;
// import com.rylink.ckb.service.ICkbService;
// import com.rylink.ckb.util.ckbutil.model.AddrInfo;
// import com.rylink.ckb.util.ckbutil.model.SendInfo;
// import com.rylink.ckb.util.httpresp.ResponseMessage;
// import com.rylink.ckb.util.httpresp.ResultEnum;
// import com.rylink.ckb.util.httpresp.ResultUtil;
// import io.swagger.annotations.Api;
// import io.swagger.annotations.ApiOperation;
// import io.swagger.annotations.ApiParam;
// import lombok.extern.slf4j.Slf4j;
// import org.springframework.beans.factory.annotation.Autowired;
// import org.springframework.validation.BindingResult;
// import org.springframework.validation.ObjectError;
// import org.springframework.web.bind.annotation.*;
//
// import javax.validation.Valid;
// import java.util.List;
//
// @Slf4j
// @RestController
// @Api(value = "CKB基本接口")
// public class CkbController2 {
//  @Autowired private ICkbService ckbService;
//
//  @PostMapping("/createaddr")
//  @ApiOperation(value = "创建地址", notes = "创建地址,生成ab文件在指定目录")
//  public ResponseMessage CreateAddr(
//      @ApiParam(name = "地址参数", value = "传入json格式", required = true) @RequestBody @Valid
//          CreateAddrBO bo,
//      BindingResult bindingResult)
//      throws Exception {
//    if (bindingResult.hasErrors()) {
//      for (ObjectError error : bindingResult.getAllErrors()) {
//        log.info(error.getDefaultMessage());
//      }
//      // 取第一个错误
//      return ResultUtil.error(
//          ResultEnum.UNKNOWN_ERROR.getCode(),
//          bindingResult.getAllErrors().get(0).getDefaultMessage());
//    }
//    if (!bo.getCoinName().equals("ckb")) {
//      return ResultUtil.error(ResultEnum.PARAMS_ERROR.getCode(), "error coinName");
//    }
//
//    List<AddrInfo> addrs = ckbService.CreateAddrToCsv(bo);
//    if (addrs == null || addrs.size() == 0) {
//      return ResultUtil.error(ResultEnum.UNKNOWN_ERROR.getCode(), "生成地址错误");
//    }
//    return ResultUtil.success(addrs);
//  }
//
//  @PostMapping("/sign")
//  @ApiOperation(value = "ckb签名", notes = "ckb签名")
//  public String SignTx(
//      @ApiParam(name = "签名结构", value = "传入json格式", required = true) @RequestBody @Valid TxTplBO
// bo,
//      BindingResult bindingResult)
//      throws Exception {
//    Gson gson = new Gson();
//    SendInfo sendInfo = ckbService.GetSignHex(bo);
//    if (sendInfo == null) {
//      return gson.toJson(
//          ResultUtil.error(ResultEnum.SIGN_ERROR.getCode(), (ResultEnum.SIGN_ERROR.getMsg())));
//    }
//    sendInfo.setCoinName("ckb");
//    sendInfo.setMchId(bo.getMchId());
//    sendInfo.setOrderId(bo.getOrderId());
//    if (!sendInfo.getErrmsg().equals("")) {
//      return gson.toJson(ResultUtil.error(ResultEnum.SIGN_ERROR.getCode(), sendInfo.getErrmsg()));
//    }
//    // 存在两种方式的json结构，不好控制，返回字符串处理
//    return gson.toJson(ResultUtil.success(sendInfo));
//  }
//
//  @GetMapping("/vaildaddress")
//  @ApiOperation(value = "地址验证", notes = "地址验证,只验证主链地址")
//  public ResponseMessage VaildAddress(@RequestParam(name = "address") String address)
//      throws Exception {
//
//    if (address == null || address.equals("")) {
//      return ResultUtil.error(
//          ResultEnum.PARAMS_ERROR.getCode(), (ResultEnum.PARAMS_ERROR.getMsg()));
//    }
//    AddressVO vo = ckbService.PaseAddress(address);
//    if (vo == null) {
//      return ResultUtil.error(
//          ResultEnum.UNKNOWN_ERROR.getCode(), (ResultEnum.UNKNOWN_ERROR.getMsg()));
//    }
//    return ResultUtil.success(vo);
//  }
// }
