# 声学测线调查服务

声学测线调查服务是一个本地 Go HTTP 服务，用于把近岸声学采样过程中的测线状态、回波读数、质量告警和作业摘要放在同一条可复核流程中。调查技术员建立并启动测线后提交读数；分析员可以在关闭作业后查看聚合摘要；质量复核员可以检查读数告警。

## 运行

```bash
go run ./cmd/acousticd
curl http://127.0.0.1:18087/health
```

服务默认监听 `127.0.0.1:18087`，可用 `ACOUSTIC_ADDR` 覆盖。

## API 入口

- `GET /health`：服务健康状态。
- `GET /v1/bands`：可用声学频段和校准范围。
- `POST /v1/surveys`：建立草稿测线。
- `POST /v1/surveys/{id}/activate`：启动采样。
- `POST /v1/surveys/{id}/readings`：提交声学读数。
- `POST /v1/surveys/{id}/close`：结束测线。
- `GET /v1/surveys/{id}/summary`：读取作业摘要。
- `GET /v1/surveys/{id}/alerts`：读取质量告警。

## 目录结构

- `cmd/acousticd`：HTTP 服务入口和信号处理。
- `internal/api`：路由、请求解码和 JSON 响应。
- `internal/survey`：测线生命周期与内存存储。
- `internal/reading`：读数校验、归一化和质量评分。
- `internal/alert`：质量告警规则和查询。
- `internal/report`：作业摘要和频段统计。
- `internal/catalog`：频段校准资料。
- `internal/model`：跨模块领域数据结构。

## 构建与测试

```bash
go build ./...
go test ./...
```

项目不依赖外部数据库或在线服务；数据只保存在进程内存中，重启服务后会重新初始化。
