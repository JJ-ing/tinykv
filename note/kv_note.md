TinyKv note
============================
made by junjie

## 环境准备

- 官网安装go,设置好PATH和GOPATH

        export GOROOT=/usr/local/go # 引号内设置为你自己的go安装目录
        export GOBIN=$GOROOT/bin
        export GOPATH=/home/luo/Workspace/golang # 引号内设置为自己的go项目的工作区间
        export PATH=$PATH:$GOROOT/bin    # 原路径后用冒号连接新路径

    普通用户和root用户都设置；~/.profile和~/.bashrc都设置。

- 官网安装goland

- 将Piangcap的项目fork到自己的git仓库中，本地建立TinyKv仓库并将项目clone到此处。

- 用Goland打开项目，发现import的外部模块出错。

    go开发是基于模块（包）的，传统的方法，项目必须位于GOPATH的src目录下，外部包也要位于该目录下。所以需要根据import的层次建立本地目录，很麻烦。

    Go module 功能需要您的 Go 版本在 Go 1.11 及以上。如果使用go mod，那么可以在任何地方建立go项目，外部包会自动下载到GOPATH 的pkg目录下，我们可以很轻松的import外部包和本项目的其他包。

    为了开启 Go module 功能只需配置一个环境变量。 

        export GO111MODULE="on"

    同时，一般go的官网被墙，需要设置代理，即配置 Goproxy 环境变量

        go env -w GOPROXY="https://goproxy.io,direct"

    将其设置到 ~/.profile文件中即可永久生效

        # 设置你的 bash 环境变量
        echo "export GOPROXY=https://goproxy.io" >> ~/.profile && source ~/.profile

    goproxy.io详情见[官网](https://goproxy.io/)。

- 再次用Goland打开项目，检查file->settings->Go的设置，特别注意开启 Go Modules-> Enable Go Modules integration。

    此时执行make命令就可以利用go mod下载并管理包，编译运行项目了。如果部分外部包无法检测，请更换系统的下载源。


## project1


