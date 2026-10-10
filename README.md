# AI Resume Optimizer（Go）

面向程序员的 AI 简历优化后端服务：上传/粘贴简历 → 结构化解析 → 多维度评分 → 优化建议 → 生成新版简历 → 导出。

产品需求见 [`docs/PRD-AI简历优化网站.md`](docs/PRD-AI简历优化网站.md)。

## 架构

```
cmd/server          HTTP 入口
internal/
  app/              依赖装配
  config/           环境配置
  domain/           领域模型
  ai/               AI Provider 接口（rule / 可扩展 openai）
  repository/       仓储接口 + memory 实现
  service/
    parser/         格式提取 + 结构化
    analyzer/       五维评分与问题清单
    suggestion/     可执行优化建议
    generator/      模板渲染与版本生成
    exporter/       Markdown / PDF 导出
    resume/         业务流程编排
  handler/          HTTP API
pkg/                通用工具（idgen、response）
```

扩展点：

| 接口 | 用途 |
|------|------|
| `parser.TextExtractor` | 新增 PDF/DOCX 解析器 |
| `parser.StructureBuilder` | 替换为 LLM 结构化 |
| `analyzer.Analyzer` | 替换为 LLM 评分 |
| `suggestion.Suggester` | 替换为 LLM 建议 |
| `generator.TemplateRenderer` | 新增简历模板 |
| `exporter.Exporter` | 新增 DOCX 等导出 |
| `ai.Provider` | 接入 OpenAI 等增强层 |
| `repository.*` | 替换为 PostgreSQL / Redis |

## 快速开始

```bash
# 运行
make run
# 或
go run ./cmd/server

# 测试
make test
```

默认监听 `:8080`。

### 环境变量

| 变量 | 默认 | 说明 |
|------|------|------|
| `SERVER_ADDR` | `:8080` | 监听地址 |
| `MAX_UPLOAD_BYTES` | `10485760` | 上传上限（10MB） |
| `AI_PROVIDER` | `rule` | `rule`（内置规则引擎） |
| `ENABLE_CORS` | `true` | 跨域 |
| `REQUEST_TIMEOUT_SEC` | `60` | 写超时基准 |

## API 一览

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/healthz` | 健康检查 |
| POST | `/api/v1/resumes/analyze` | 上传/粘贴并分析 |
| GET | `/api/v1/resumes/{id}` | 获取简历 |
| PUT | `/api/v1/resumes/{id}/structured` | 修正结构化字段 |
| GET | `/api/v1/resumes/{id}/analysis` | 获取分析报告 |
| GET | `/api/v1/analyses/{id}/suggestions` | 建议列表 |
| PATCH | `/api/v1/suggestions/{id}` | 采纳/忽略/编辑建议 |
| POST | `/api/v1/versions/generate` | 生成优化版 |
| GET | `/api/v1/versions/{id}` | 获取版本 |
| GET | `/api/v1/versions/{id}/export?format=markdown\|pdf` | 导出 |
| GET | `/api/v1/compare?resumeId=&versionId=` | 原版 vs 新版 |
| DELETE | `/api/v1/resumes/{id}` | 删除简历 |

### 示例：粘贴文本分析

```bash
curl -s -X POST http://localhost:8080/api/v1/resumes/analyze \
  -H 'Content-Type: application/json' \
  -d '{
    "text": "张三\nzhang@example.com\n\n专业技能\nGo Redis\n\n工作经历\n某公司\n后端\n2021-2024\n- 负责开发\n",
    "targetRole": "Go 后端工程师",
    "targetJd": "Go Redis Docker 高并发"
  }'
```

### 示例：生成并导出

```bash
# 将上一步返回的 resume.id 填入
curl -s -X POST http://localhost:8080/api/v1/versions/generate \
  -H 'Content-Type: application/json' \
  -d '{"resumeId":"res_xxx","template":"tech"}'

curl -OJ "http://localhost:8080/api/v1/versions/ver_xxx/export?format=markdown"
```

## 设计说明

- **MVP 默认使用规则引擎**：无外部 API Key 即可跑通全流程，评分与建议确定性强、便于单测。
- **禁止捏造**：生成阶段只改写/补充联系方式与已有经历表述，不虚构公司与项目。
- **PDF**：内置最小合法 PDF（Helvetica）；中文完整排版请优先导出 Markdown，或替换 `TextPDFExporter`。
- **存储**：内存仓储，进程重启丢失；实现 `repository` 接口即可持久化。

## License

MIT
