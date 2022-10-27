package com.rylink.ckb.controller;

import com.rylink.ckb.util.httpresp.MyException;
import com.rylink.ckb.util.httpresp.ResponseMessage;
import com.rylink.ckb.util.httpresp.ResultEnum;
import com.rylink.ckb.util.httpresp.ResultUtil;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.bind.annotation.ControllerAdvice;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.ResponseBody;



//全局错误拦截返回
@ResponseBody
@Slf4j
@ControllerAdvice
public class MyControllerAdvice {
    @ExceptionHandler(value = Exception.class)
    public ResponseMessage handel(Exception e){
        if(e instanceof MyException){
            MyException myException =(MyException)e;
            return ResultUtil.error( myException.getCode(),myException.getMessage());
        }else {
            log.error("[系统异常] {}",e);
            return ResultUtil.error(ResultEnum.SYSTEM_ERROR);
        }
    }
}
