# 修复前故障复现（Docker）

## 项目与标准命令

服务通过 `go run ./cmd/acousticd` 启动，读数写入后可从测线告警接口查询质量结果。

## 环境构建与编译

在当前机器平台执行 `docker build -f benzhi.Dockerfile -t acoustic-survey:bug .`，镜像构建完成后在容器中执行 `go build ./...`。

## 故障触发步骤

在活动的 coastal-38khz 测线上提交频率为 37000 Hz 的回波，再查询告警列表。

## 实际错误输出

`go test -count=1 -run TestCalibrationBoundaryDoesNotCreateAlert ./internal/api`

```text
alerts = 1, want 0
```

## 期望行为

标定区间端点属于允许范围，边界样本不生成范围外告警。
