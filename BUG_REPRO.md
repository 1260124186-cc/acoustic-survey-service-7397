# 修复前故障复现（Docker）

## 项目与标准命令

服务通过 `go run ./cmd/acousticd` 启动，读数从 `POST /v1/surveys/{id}/readings` 写入，摘要从 `GET /v1/surveys/{id}/summary` 读取。

## 环境构建与编译

在当前机器平台执行 `docker build -f benzhi.Dockerfile -t acoustic-survey:bug .`，镜像构建完成后在容器中执行 `go build ./...`。

## 故障触发步骤

建立并启动一条测线，使用相同的 `X-Capture-ID` 连续提交两次同一份读数，再读取该测线的摘要。

## 实际错误输出

`go test -count=1 -run TestRepeatedCaptureIDDoesNotDuplicateReading ./internal/api`

```text
total readings = 2, want 1
```

## 期望行为

重复采集编号只保留一条读数，摘要总读数为 1。
