## 接口文档


- 默认头部请求参数：

1) X-Ca-Nonce: 60分钟内不能重复的随机字符串

2) X-Ca-Time: 精确到秒的时间戳

3) X-Ca-SignStr: 签名字符串，加密公式 AES((Token + X-Ca-Nonce + X-Ca-Time),nonce) 

- Token问题：

1. http头部请求加 Authorization
2. Authorization 内容格式 Bearer token

```shell
Bearer CI6MSwibmFtZSI6ImFkbWluIiwiYWNjb3VudCI6IjQ3MzAyMjQ1N0BxcS5jb20iLCJhZG1pbiI6dHJ1ZSwiZXhwIjoxNjM2NTYxOTMxfQ.e8mdSig7QEHq8X5b2fJmuYxsimroqQU3BA04beUNWA4
```
