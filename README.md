# GO-OPENAI-PROXY

基于 Go + 腾讯云云函数（部署到海外节点）实现 OpenAI API 代理

编译打包：

```bash
./bash.sh
```

然后在腾讯云云函数代码管理界面上传打包好 zip 包即可完成部署：

![](https://image.gstatics.cn/2023/03/06/image-20230306171340547.png)

你可以通过腾讯云云函数提供的测试工具进行测试，也可以本地通过 curl/postman 进行测试：

![](https://image.gstatics.cn/2023/03/06/image-20230306173648325.png)

你可以选择自己搭建，也可以直接使用我提供的代理，域名是 `openai.geekr.cool`，反正是免费的。关于代理背后的原理，可以看我在极客书房发布的这篇教程：[国内无法调用 OpenAI 接口的解决办法](https://geekr.dev/posts/chatgpt-website-by-laravel-10#toc-5)。

