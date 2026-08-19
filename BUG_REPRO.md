# 故障复现

## 项目与标准命令

进入 `env/` 后执行补传读数告警验证：

```bash
go test -count=1 -run TestLateArrivalDoesNotCreateAbruptChangeAlert ./internal/api
```

## 环境构建与编译

```bash
go build ./...
```

构建应成功。

## 故障触发步骤

1. 创建并启动 38kHz 测线。
2. 先提交 10:00 和 10:02 的读数，再补传 10:01 的读数。
3. 查询该测线的质量告警。

## 实际错误输出

```text
arrival order created a false abrupt alert: abrupt_normalized_change
```

## 期望行为

按实际采样时间变化平滑的序列不应仅因补传到达顺序而产生突变告警。
