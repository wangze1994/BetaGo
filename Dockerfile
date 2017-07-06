# 得到最新的 golang docker 镜像
FROM golang:latest
# 配置环境变量
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
ENV GOBIN /go/bin
# 创建工作目录
RUN mkdir -p /go/src/dingTalk
# 定义dingTalk工作目录
WORKDIR /go/src/dingTalk
# 复制dingTalk目录到容器中
COPY . /go/src/dingTalk
# 下载并安装第三方依赖到容器中
RUN go get github.com/robfig/cron
# 将dingtalk编译到GOBIN下
RUN go install -v dingtalk.go
# 告诉 Docker 启动容器运行的命令
ENTRYPOINT [ "dingtalk" ]