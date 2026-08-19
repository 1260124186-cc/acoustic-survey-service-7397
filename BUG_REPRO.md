# 故障复现

## 项目与标准命令

进入 `env/` 后执行关闭后读数计数验证：

```bash
go test -count=1 -run TestClosedSurveyRetainsRecordedReadingCount ./internal/api
```

## 环境构建与编译

```bash
go build ./...
```

构建应成功。

## 故障触发步骤

1. 创建并启动测线。
2. 提交一条读数并关闭测线。
3. 查询关闭后的测线详情。

## 实际错误输出

```text
closed survey reading_count=0, want 1
```

## 期望行为

关闭仅结束作业状态，已记录读数数量必须继续保留在测线详情中。
