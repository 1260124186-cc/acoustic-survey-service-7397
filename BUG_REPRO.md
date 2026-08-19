# 修复前故障复现（Docker）

## 项目与标准命令

服务通过 `go run ./cmd/acousticd` 启动，读数写入后从测线告警接口读取结果。

## 环境构建与编译

在当前机器平台执行 `docker build -f benzhi.Dockerfile -t acoustic-survey:bug .`，镜像构建完成后在容器中执行 `go build ./...`。

## 故障触发步骤

先写入较晚采集的异常样本，再补传较早采集的异常样本，随后读取告警列表。

## 实际错误输出

`go test -count=1 -run TestAlertsAreOrderedByCaptureTimeAfterLateArrival ./internal/api`

```text
alerts are not chronological
```

## 期望行为

告警列表按样本采集时间从早到晚返回，方便还原现场过程。
