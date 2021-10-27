# SubConfig
subconverter外部配置以及利用actions自动更新订阅转换

重点在利用actions更新订阅转换并发布到自己的服务器，  
[.github/workflows/subconverter.yml](.github/workflows/subconverter.yml)

## Getting Started
fork后点击右上角的 Star 星星按钮即可试用，  
需要使用发布功能的话需要配置几个secrets  
- SUBSCRIBE 订阅链接，一行一个订阅，支持tg格式的http代理socks5代理，
  ```
  tg://http?server=1.2.3.4&port=233&user=user&pass=pass&remarks=Example
  ```
- UPLOAD_SECRET 配置生成打包压缩后加密时用的aes密钥，由如下命令生成的单行的base64字符串，
  ```shell
  head -c 32 /dev/urandom |base64 > upload_secret
  ```
- DEPLOY_URL 发布的上传地址，script目录中有php写的接收端，作用是接收加密的配置解密解压后发布到当前服务器上，
  [script/upload.php](script/upload.php)
  ```
  https://host/upload.php
  ```
- UPLOAD_TOKEN 发布时用于验证的token参数， 
  secret也能起到验证作用，但是安全起见，secret只当密钥不参与网络传输，token明文传输用于接收文件前的验证，
  token由如下命令生成的单行的16进制字符串，
  ```shell
  head -c 32 /dev/urandom |od -A n -v -t x1 | tr -d ' \n' > upload_token
  ```
四个私密数据都设置了才会尝试发布，  
如果SUBSCRIBE订阅没有配置的话就会用默认示例节点打包配置加密再解密上传到actions artifact供下载查看，  
因此**可以fork仓库不配置直接试用**，    

另外为了加速下载外部配置，这里直接缓存了存在大量引用的当前项目和ACL4SSR，并通过地址替换变成本地文件引用避免了大量重复下载配置，

## 下载配置
发布的配置使用script目录下的另一个php提供下载，
[script/sub.php](script/sub.php)
这sub.php不是个真正的subconverter后端，但也可以当后端用，因为sub_secret不匹配时会直接301重定向到某个开放的订阅转换后端服务器上，  
- sub_secret是代替键为url的参数，是个任意url安全的字符串，最好要自己能记得住，以便客户端输入订阅地址可以手动输入，如下方式输入，
  默认使用的配置文件是[subconverter.ini](subconverter.ini)  
  如果url参数中secret后跟着内容，就会拼接到下载地址的subconverter后面，比如url=secret-basic就会使用[subconverter-basic.ini](subconverter-basic.ini)  
  ```
  https://host/sub?url=<secret>
  https://host/sub?url=<secret>-basic
  ```
- 支持target参数和var参数，
  参考[subconverter-README.md](https://github.com/tindy2013/subconverter/blob/master/README-cn.md#%E6%94%AF%E6%8C%81%E7%B1%BB%E5%9E%8B)
  ```
  https://host/sub?target=quan&url=<secret>-basic
  ```
- 参数传递cache=false或者服务器上没有缓存就会使用外部订阅转换后端获取到配置并缓存起来，
  这个过程会注入当前目录下的subscribe中的订阅链接，并忽略其他参数，直接使用写死的[script/sub.php#L11](script/sub.php#L11)
  ```php
  $example_params = "emoji=true&list=false&udp=true&tfo=false&scv=false&fdn=false&sort=false&new_name=true";
  ```

## 服务器上的私密文件要禁止直接请求
只能通过php验证身份后读取，  
参考nginx配置，
```
    location = /subscribe {
        deny all;
    }
    location = /sub_secret {
        deny all;
    }
    location ~ ^/config.* {
        deny all;
    }
    location ~ ^/upload_.* {
        deny all;
    }
```
