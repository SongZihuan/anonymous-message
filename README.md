# 匿名消息接受工具 AnonymousMessage
接受网站上的电子信息，然后转发到企业微信机器人。

## API
`/`，`/message`，`/message/` - 推送消息API，接受JSON，例如：`{"name":"Song","message":"测试","refer":"站点1"}`

`/hello`，`/hello/` - 返回文本`Hello, world!`

## 启动参数
通过`-h`获得`useage`。

例如:
```
--address <HTTP绑定地址端口，默认：:3352>
--webhook <企业微信Webhook>
--redis-address <Redis地址>
--redis-password <Redis密码>
--redis-db <Redis数据库，默认：0>
--origin <空白则允许所有，允许跨域的Origin，若存在多个则以英文逗号分隔，可以使用*匹配全部，但不建议，因为会导致请求头 Access-Control-Allow-Headers 出现问题>
```

### 关于 Access-Control-Allow-Headers
当以`*`全部`Origin`，而请求刚好没`Origin`，则`Access-Control-Allow-Origin`会被设置为`*`。
程序中`Access-Control-Allow-Headers`也设置为`*`表示允许全部请求头。
因此导致问题，根据跨域规则，`Access-Control-Allow-Headers`设置为`*`的时候`Access-Control-Allow-Origin`必须是具体`Origin`，而不能是`*`。

### 关于Origin参数
若设置`Origin`为空白（或不设置），则允许所有跨域，一切请求过来都不做跨域检查，而所有预检都返回允许，并且全部请求都包括允许跨域的请求头。
若设置多个`Origin`则以英文逗号分割，例如：`--origin https://www.song-zh.com,https://song-zh.com`

## 协议
本软件基于[MIT LICENSE](./LICENSE)协议发布。
