# acoustic-survey-service__006 Docker 交付说明

## 项目概览
- 声学测线调查服务是一个本地 Go HTTP 服务，用于把近岸声学采样过程中的测线状态、回波读数、质量告警和作业摘要放在同一条可复核流程中。调查技术员建立并启动测线后提交读数；分析员可以在关闭作业后查看聚合摘要；质量复核员可以检查读数告警。
- Go module: `example.com/acoustic-survey-service`

## 标准命令

```bash
go build ./...
go test ./...
```

## 实际启动入口

```bash
go run ./cmd/acousticd
```

## Docker 构建

```bash
./build_benzhi_docker.sh acoustic-survey-service__006-benzhi linux/amd64
docker run --rm -it acoustic-survey-service__006-benzhi bash
```

## 环境

- 基础镜像: `golang:1.26.2`
- 依赖在镜像构建阶段预下载，容器内可直接执行 Go 构建和测试命令。
- 代码目录: `/app`
- 源码中检测到的服务端口: `30`, `18087`, `38000`
