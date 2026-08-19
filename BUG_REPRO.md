# 故障复现

## 项目与标准命令

本服务用于记录近岸声学测线读数。进入 `env/` 后执行并发读数验证：

```bash
go test -race -count=1 -run TestConcurrentReadingsReceiveDistinctIDs ./internal/reading
```

## 环境构建与编译

```bash
go build ./...
```

构建应成功。

## 故障触发步骤

1. 创建并启动一条 `coastal-38khz` 测线。
2. 让 32 个采集协程同时提交读数。
3. 检查每次提交返回的读数编号。

## 实际错误输出

```text
duplicate reading id returned: parallel-line-001
```

## 期望行为

每条并发提交的读数都应被保留，并获得唯一、连续且可追溯的编号。
