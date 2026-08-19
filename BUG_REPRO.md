# 修复前故障复现（Docker）

## 项目与标准命令

项目使用 Go 1.26.2，在当前 linux/arm64 平台验证。镜像通过 `docker build -f benzhi.Dockerfile -t go-acoustic-survey-service__013-bug:20260819 .` 构建；容器内标准检查命令为 `go build ./...` 和 `go test ./...`。

## 环境构建与编译

Bug 环境镜像构建成功，容器内 `go version` 输出 `go version go1.26.2 linux/arm64`，`go build ./...` 退出码为 0。

## 故障触发步骤

在 Bug 环境容器中执行：

```text
go test ./...
```

## 实际错误输出

```text
2026/08/19 13:33:42 POST /v1/surveys completed in 0s
2026/08/19 13:33:42 POST /v1/surveys/quality-close/activate completed in 0s
2026/08/19 13:33:42 POST /v1/surveys/quality-close/readings completed in 0s
2026/08/19 13:33:42 POST /v1/surveys/quality-close/close completed in 0s
--- FAIL: TestCloseRejectsSurveyWithOnlyUnusableReadings (0.00s)
    close_requires_usable_reading_test.go:38: close status = 200, want 400
FAIL
FAIL    example.com/acoustic-survey-service/internal/api    0.002s
FAIL
```

## 期望行为

当测线只包含超出标定范围的回波时，结束请求应返回失败，测线保持可继续补采的状态。
