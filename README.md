# GO-OPENAI-PROXY

基于 Go + 腾讯云 API 网关 + 云函数（部署到海外节点）实现 OpenAI API 调用代理

### 编译打包：

```bash
./build.sh
```

### 部署测试

然后在腾讯云云函数代码管理界面上传打包好 zip 包即可完成部署：

![](https://image.gstatics.cn/2023/03/06/image-20230306171340547.png)

你可以通过腾讯云云函数提供的测试工具进行测试，也可以本地通过 curl/postman 进行测试，使用的时候只需要将 `api.openai.com` 替换成代理域名 `open.aiproxy.xyz` 即可：
 
![](https://geekr.gstatics.cn/wp-content/uploads/2023/03/image-38.png)

你可以选择自己搭建，也可以直接使用我提供的代理域名 `open.aiproxy.xyz`，反正是免费的。关于代理背后的原理，可以看我在极客书房发布的这篇教程：[国内无法调用 OpenAI 接口的解决办法](https://geekr.dev/posts/chatgpt-website-by-laravel-10#toc-5)。

本地调试走VPN的话可以设置环境变量 `ENV=local`，然后直连 `api.openai.com`：

```go
// 本地测试通过代理请求 OpenAI 接口
if os.Getenv("ENV") == "local" {
    proxyURL, _ := url.Parse("http://127.0.0.1:10809")
    client.Transport = &http.Transport{
        Proxy:           http.ProxyURL(proxyURL),
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
}
```
### 流式响应支持

这个源代码本身是支持 stream 流式响应代理的，但目前腾讯云函数并不支持分块流式传输。所以，如果你需要实现流式响应，可以把编译后的二进制文件 `main` 丢到任意海外云服务器运行，这样就变成支持流式响应的 OpenAI 代理了，如果你不想折腾，可以使用我这边提供的 `open2.aiproxy.xyz` 作为代理进行测试：

<img width="965" alt="image" src="https://user-images.githubusercontent.com/114386672/225609817-ca5c106b-22d4-4ae9-b3df-ca2c46d56843.png">
