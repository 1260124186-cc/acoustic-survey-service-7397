# 修复前故障复现（Docker）

## 项目与标准命令

服务通过 `go run ./cmd/acousticd` 启动，读数由内部读数服务校验并写入。

## 环境构建与编译

在当前机器平台执行 `docker build -f benzhi.Dockerfile -t acoustic-survey:bug .`，镜像构建完成后在容器中执行 `go build ./...`。

## 故障触发步骤

向活动测线提交包含非有限回波值的读数。

## 实际错误输出

`go test -count=1 -run TestRejectsNonFiniteEchoLevel ./internal/reading`

```text
non-finite echo was accepted
```

## 期望行为

无效的非有限测量在持久化前返回错误，不进入摘要和告警流程。
