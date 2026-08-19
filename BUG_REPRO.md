# 修复前故障复现（Docker）

## 项目与标准命令

本服务用于记录近岸声学测线读数。进入 `env/` 后执行关闭后补传验证：

```bash
go test -count=1 -run TestClosedSurveyRejectsLateReading ./internal/api
```

## 环境构建与编译

在当前机器平台构建镜像后，容器内执行：

```bash
go build ./...
```

构建成功。

## 故障触发步骤

1. 创建并启动一条 `coastal-38khz` 测线。
2. 提交一条读数并调用关闭接口。
3. 关闭后再次提交设备补传读数，并查询测线摘要。

## 实际错误输出

```text
closed survey accepted a late reading
```

## 期望行为

测线关闭后应拒绝后续读数，摘要应保留关闭前已经记录的样本数。
