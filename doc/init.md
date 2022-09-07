
# 创建文档应用

首先登陆[飞书开放平台](https://open.feishu.cn/app?lang=zh-CN)创建企业自建应用，选择"机器人/网页/小程序"，配置"云文档-多维表格"权限。
然后发布版本，并在[飞书管理后台](https://www.feishu.cn/admin/appCenter/audit) 审批通过应用。
在[云文档](https://www.feishu.cn/drive/home)创建一篇多维表格，从链接中获取`app_token`。类似：bascnQIrLs6MrhIvftGsdYJgRFd, bas 开头。

查看应用[密钥](https://open.feishu.cn/app/cli_a14eda43cb7ad013/baseinfo)
```
cli_a14eda43cb7ad013
l5zyi***********************16Y0
```
拼装一下得到 dsn。
```
bitable://cli_a14eda43cb7ad013:l5zyi***********************16Y0@open.feishu.cn/bascnQIrLs6MrhIvftGsdYJgRFd
```

