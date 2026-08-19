# 故障复现

## 项目与标准命令

进入 `env/` 后执行测线摘要频段归属验证：

```bash
go test -count=1 -run TestSummaryKeepsSurveyBandWhenCatalogRangesOverlap ./internal/api
```

## 环境构建与编译

```bash
go build ./...
```

构建应成功。

## 故障触发步骤

1. 创建并启动目标频段为 `coastal-38khz` 的测线。
2. 提交频率为 37500Hz 的读数。
3. 查询该测线摘要中的频段计数。

## 实际错误输出

```text
coastal band count=0, summary=map[sector-02-window-10:1]
```

## 期望行为

测线摘要应优先按该测线选择的目标频段统计这条读数。
