# 修复前故障复现（Docker）

## 项目与标准命令

服务通过 `go run ./cmd/acousticd` 启动，读数写入与告警查询分别使用测线读数和告警接口。

## 环境构建与编译

在当前机器平台执行 `docker build -f benzhi.Dockerfile -t acoustic-survey:bug .`，镜像构建完成后在容器中执行 `go build ./...`。

## 故障触发步骤

先写入较早和较晚的回波，再补传中间时刻的稳定回波，随后查询该测线告警。

## 实际错误输出

`go test -count=1 -run TestLateArrivalUsesChronologicalPredecessorForAlerts ./internal/api`

```text
alerts = 2, want 1
```

## 期望行为

补传样本只和采样时间上真正相邻的前一条读数比较，保留唯一真实的突变告警。
