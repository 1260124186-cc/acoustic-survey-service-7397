# 故障复现

## 项目与标准命令

进入 `env/` 后执行低信噪比读数摘要验证：

```bash
go test -count=1 -run TestSummaryExcludesLowSignalToNoiseReadingFromValidCount ./internal/api
```

## 环境构建与编译

```bash
go build ./...
```

构建应成功。

## 故障触发步骤

1. 创建并启动 38kHz 测线。
2. 写入一条低于该频段最小信噪比阈值的读数。
3. 查询作业摘要与告警数。

## 实际错误输出

```text
valid readings=1, want 0 for low SNR
```

## 期望行为

低信噪比读数应继续保留告警，但不能计入有效读数或有效均值。
