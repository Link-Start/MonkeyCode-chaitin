# API

OpenAPI 源文件按调用方分为 `admin.yaml` 和 `agent.yaml`。管理后台接口统一维护在 `admin.yaml`，工作 Agent 接口统一维护在 `agent.yaml`，不再按业务模块拆分或生成合并文档。

每个接口必须声明请求体、成功与错误响应、响应 JSON Schema，并为字段补充类型、约束和说明；不能只记录状态码和响应描述。契约字段以实际 HTTP DTO 为准，密钥等不返回的敏感字段需要明确说明。
