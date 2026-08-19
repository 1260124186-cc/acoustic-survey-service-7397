# 修复前故障复现（Docker）

## 复现命令

在 Bug 分支的 Go 环境中执行：

```text
go test -count=1 -run TestCloseRejectsSurveyWithBlockingAlert ./internal/api
```

## 预期故障

活动测线写入频率超出 coastal-38khz 标定范围的读数后，会产生未解决的严重告警。修复前结束接口错误返回 200 并将测线标记为 closed；正确行为应返回 400，保留活动状态。低信噪比等普通告警不应阻断结束。
