# MyAgent

MyAgent 是一个使用 Go 构建的 Agent 框架原型，目标是提供一组清晰、可替换的基础组件，用于实现带工具调用和长期记忆的 LLM Agent。

当前仓库已经完成领域模型、LLM 抽象、Agent 工具调用循环、工具执行与校验，以及内存版记忆存储。项目仍处于早期阶段，尚未接入具体的大模型提供商，也没有提供 HTTP 或 CLI 服务。

## 当前能力

- 统一的 LLM 接口，支持普通生成和流式事件的类型定义。
- Agent 运行时，可根据模型的 `tool_calls` 执行工具，并把结果送回模型继续推理。
- 工具注册、JSON Schema 参数校验和串行工具执行。
- 内置 `calculator` 工具，支持加法、减法和乘法。
- 记忆模型与 `Store`、`Extractor` 抽象。
- 线程安全的内存记忆存储，支持按用户隔离、多关键词检索、相关度排序、更新和删除。
- 可注入的测试用记忆提取器，便于后续编写测试。

## 快速开始

### 环境要求

- Go 1.27 或更高版本

### 运行示例

当前入口是记忆存储示例，而不是一个会监听端口的 Web 服务：

```bash
go run ./cmd/server
```

示例会创建两个 `user_001` 的记忆和一个 `user_002` 的记忆，然后以“自然风景”为查询词搜索 `user_001` 的记忆，输出：

```text
用户喜欢自然风景。
```

### 检查项目

```bash
go test ./...
go vet ./...
```

这些命令目前可以通过。`internal/memory` 已包含表驱动单元测试，覆盖多关键词匹配、相关度排序、稳定排序和重复关键词去重。

## 项目结构

```text
.
├── cmd/
│   └── server/              # 当前的可运行示例入口
├── internal/
│   ├── agent/               # Agent 配置、错误和执行循环
│   ├── domain/              # Agent、Task、Run、Message、ToolCall 等领域模型
│   ├── llm/                 # LLM 请求、响应、流式事件和统一接口
│   ├── memory/              # 记忆模型、存储、提取器和内存实现
│   └── tool/
│       ├── builtin/         # 内置工具
│       ├── executor.go      # 工具执行器
│       ├── registry.go      # 线程安全的工具注册表
│       └── validator.go     # JSON Schema 参数校验
├── go.mod
└── go.sum
```

各包的主要职责如下：

| 包 | 职责 |
| --- | --- |
| `internal/domain` | 定义跨组件共享的 Agent、任务、运行记录、消息和工具调用模型 |
| `internal/llm` | 屏蔽不同模型提供商的差异；目前只有接口和数据结构 |
| `internal/tool` | 注册工具、向模型暴露工具定义、校验参数并执行工具 |
| `internal/memory` | 保存和检索用户记忆；提供内存实现及提取器接口 |
| `internal/agent` | 组合 LLM、工具和记忆，驱动多轮 Agent 循环 |
| `cmd/server` | 展示组件用法；暂未提供网络服务 |

## 运行链路

### 当前示例

```text
main
  -> NewInMemoryStore
  -> Save(user memories)
  -> Search(userID, query, limit)
  -> print matched memories
```

### Agent 设计流程

```text
RunRequest
  -> 检索用户相关记忆
  -> 将记忆追加到系统提示词
  -> 获取已注册的工具定义
  -> 调用 LLM
       -> FinishReasonStop：返回最终回答
       -> FinishReasonToolCalls：执行工具
            -> 将工具结果写回消息历史
            -> 再次调用 LLM
       -> FinishReasonLength：返回错误
  -> 达到最大步数后返回错误
```

一次 `Run` 会维护完整的消息历史，包括系统消息、用户输入、Assistant 的工具调用消息，以及每个工具调用对应的 Tool 消息。该设计兼容常见的工具调用协议。

## 核心组件

### LLM

`internal/llm.LLM` 定义了两个方法：

```go
type LLM interface {
    Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
    Stream(ctx context.Context, req *GenerateRequest) (<-chan StreamEvent, error)
}
```

接入具体模型时，需要实现这个接口，并把厂商的请求格式、响应格式和流式事件转换成 `internal/llm` 中的统一结构。当前仓库没有 OpenAI、Anthropic 或其他厂商的实现，所以 Agent 还不能单独发起真实的模型请求。

### 工具

一个工具需要实现 `tool.Tool`：

```go
type Tool interface {
    Definition() Definition
    Execute(ctx context.Context, arguments json.RawMessage) (string, error)
}
```

执行流程如下：

1. `Registry` 根据工具名称查找实现。
2. `Executor` 使用工具定义中的 JSON Schema 校验模型生成的参数。
3. 参数合法时调用 `Tool.Execute`。
4. 工具错误和参数错误会作为错误结果返回，并由 Agent 追加到消息历史。

新增工具时，实现 `Tool` 接口并注册到 `Registry` 即可：

```go
registry := tool.NewRegistry()
if err := registry.Register(&builtin.CalculatorTool{}); err != nil {
    return err
}

executor := tool.NewExecutor(registry, &tool.JSONSchemaValidator{})
```

当前 `ExecuteAll` 按顺序执行同轮产生的多个工具调用。

### 记忆

`memory.Store` 定义了三个操作：

```go
type Store interface {
    Save(ctx context.Context, mem *Memory) error
    Search(ctx context.Context, userID string, query string, limit int) ([]Memory, error)
    Delete(ctx context.Context, userID string, id string) error
}
```

`InMemoryStore` 目前具备以下行为：

- 按 `UserID` 隔离不同用户的记忆。
- 使用 `strings.Fields` 按空白拆分查询，支持一次查询多个关键词。
- 采用 OR 匹配：记忆包含任意一个关键词即可进入结果集。
- 查询关键词会先去重，重复关键词不会重复计分。
- 关键词匹配不区分大小写，单个关键词使用子串匹配，不是中文分词或向量检索。
- 优先返回匹配关键词数量更多的记忆；匹配数相同时按 `UpdatedAt` 从新到旧排序，时间仍相同时按 `ID` 排序。
- 完成排序后再应用 `limit`，避免遗漏相关度更高的结果。
- 更新已有记忆时保留原始 `CreatedAt`。
- 使用读写锁保护并发访问。

例如查询：

```text
自然风景 周末旅行
```

同时包含“自然风景”和“周末旅行”的记忆会排在只包含其中一个关键词的记忆之前。由于当前按空白拆词，`周末旅行` 是一个完整关键词，不能匹配内容中的“周末准备出去旅行”；如需这种能力，后续可接入中文分词或向量检索。

`MockExtractor` 用于测试记忆提取流程，它会返回预设记忆，不会从对话中自动总结新记忆。

## 当前限制

这个仓库是可运行的基础框架，不是完整的生产级 Agent：

- `internal/llm` 尚未实现任何真实模型提供商。
- `LLM.Stream` 只有类型定义，Agent 运行时尚未使用流式响应。
- `cmd/server` 目前是一次性示例，没有 HTTP 服务或路由。
- `RunTime` 已保留记忆存储和提取器字段，但 `NewRunTime` 尚未接收并注入它们；记忆检索当前实际返回空结果。
- 记忆提取器尚未接入 Agent 运行结束流程。
- `Task`、`Run` 和 `Agent` 领域模型尚未持久化，也未接入运行时。
- 当前只有记忆搜索的单元测试，Agent、LLM 和工具执行链路仍缺少测试。
- Agent 循环使用 `step < MaxSteps`，默认最多进行 9 次模型调用；后续需要统一 `MaxSteps` 的语义。

## 建议的后续工作

1. 实现一个真实 LLM 提供商，并用模拟模型稳定测试工具调用流程。
2. 通过构造函数或 Option 模式把 `memory.Store` 注入 `RunTime`。
3. 在 Agent 完成后调用 `Extractor`，将新记忆写入 `Store`。
4. 为工具注册、参数校验、工具循环、记忆保存删除和取消操作补充测试。
5. 实现持久化存储和 HTTP 接口，再逐步加入流式输出。

## 设计约定

- 依赖通过接口表达，具体实现可替换。
- 领域模型与具体模型提供商保持分离。
- 工具参数在进入业务逻辑前统一校验。
- 外部依赖或模型错误尽量使用 `%w` 包装，保留错误链。
- 截止到当前版本，对外能力以 `internal` 包中的实现为准，不将“类型已定义”描述成“功能已完成”。
