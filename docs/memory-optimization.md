# PicoClaw 记忆机制优化说明

## 概述

本次优化重构了 PicoClaw 的会话记忆系统，从单一扁平的摘要机制升级为**三层记忆架构**，解决了压缩过于粗暴、幻觉较多、历史信息记忆不好的问题。

---

## 优化前后对比

### 架构对比

| 层级 | 优化前 | 优化后 |
|------|--------|--------|
| L1 短期记忆 | 完整 JSONL 历史，无分层 | 保留最近 N 条消息（默认 10 条） |
| L2 中期记忆 | 纯文本摘要，无结构 | **结构化摘要**：事实、决策、偏好、任务状态 |
| L3 长期记忆 | 依赖 memory skill 的 MEMORY.md | 同前，配合新增的 `long_term_retention_days` 配置 |

### 压缩策略对比

| 场景 | 优化前 | 优化后 |
|------|--------|--------|
| 正常摘要触发 | 生成纯文本摘要，保留 4 条消息 | 生成**结构化摘要** + 叙事摘要，保留 4 条消息 |
| Context 超限强制压缩 | 直接丢弃 50% 消息 | 先将被丢弃部分**压缩为结构化摘要**，再丢弃 |

### 摘要质量对比

| 维度 | 优化前 | 优化后 |
|------|--------|--------|
| 格式 | 纯文本段落 | 叙事摘要 + JSON 结构化数据 |
| 决策链 | 无保留，容易丢失上下文 | 显式记录 `{context, choice, reason}` |
| 用户偏好 | 可能隐含在文本中 | 独立 `preferences` 数组 |
| 当前任务状态 | 无 | `current_task` + `pending_items` |
| 增量更新 | 全量替换 | 增量合并（`mergeStructured`） |

---

## 新增文件清单

```
pkg/memory/
├── structured_summary.go      # 结构化摘要类型定义
├── structured_summary_test.go # 单元测试
├── retrieval.go               # 三层检索机制
└── retrieval_test.go          # 检索单元测试

pkg/session/
├── session_store.go           # 新增 Get/SetStructuredSummary 接口
├── jsonl_backend.go           # 实现 L2 结构化摘要存储
└── manager.go                 # 兼容实现（返回 nil）
```

---

## 修改文件清单

### `pkg/memory/store.go`
- 新增接口方法：
  - `GetStructuredSummary(ctx, sessionKey) (*StructuredSummary, error)`
  - `SetStructuredSummary(ctx, sessionKey, *StructuredSummary) error`

### `pkg/memory/jsonl.go`
- 新增 `{sanitized_key}.structured.json` 文件支持
- 实现 `GetStructuredSummary` / `SetStructuredSummary`
- `SetStructuredSummary(nil)` 会删除文件（清空 L2）

### `pkg/agent/loop.go`
- **summarizeSession**：改用 `summarizeBatchWithStructured`，同时生成叙事摘要和结构化摘要
- **forceCompression**：在丢弃前先压缩到 L2
- **BuildMessages 调用处**：使用 `appendStructuredToSummary` 将 L2 注入系统提示
- **/clear 命令**：额外清空 L2 结构化摘要

### `pkg/config/config.go` / `defaults.go`
- 新增配置项：
  ```go
  ShortTermMessages      int  // L1 保留条数，默认 10
  LongTermRetentionDays  int  // L3 保留天数，默认 30
  ```

---

## 配置项说明

在 `config.json` 的 `agents.defaults` 中新增：

```json
{
  "agents": {
    "defaults": {
      "summarize_message_threshold": 20,
      "summarize_token_percent": 75,
      "short_term_messages": 10,
      "long_term_retention_days": 30
    }
  }
}
```

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `summarize_message_threshold` | 20 | 消息数超过此值触发摘要 |
| `summarize_token_percent` | 75 | Token 估算超过 context window 的此百分比触发摘要 |
| `short_term_messages` | 10 | L1 保留的最近消息条数 |
| `long_term_retention_days` | 30 | L3 长期记忆保留天数 |

---

## 结构化摘要格式

存储路径：`{workspace}/sessions/{session_key}.structured.json`

```json
{
  "facts": [
    "用户正在开发一个 Go 项目",
    "项目名称是 picoclaw"
  ],
  "decisions": [
    {
      "context": "选择数据库",
      "choice": "SQLite",
      "reason": "轻量级，无需额外服务"
    }
  ],
  "preferences": [
    "回答要简短",
    "不要使用 markdown 代码块"
  ],
  "current_task": "实现三层记忆架构",
  "pending_items": [
    "添加测试用例",
    "更新文档"
  ],
  "updated_at": "2026-03-13T15:30:00Z"
}
```

---

## 数据流图

```
用户消息 → AddMessage() → .jsonl (append)
                           ↓
               消息数/Token 超阈值？
                           ↓ 是
               summarizeSession()
                           ↓
         ┌─────────────────┴─────────────────┐
         ↓                                    ↓
   叙事摘要 (summary)              结构化摘要 (L2)
   → .meta.json                    → .structured.json
         ↓                                    ↓
         └─────────────────┬─────────────────┘
                           ↓
               TruncateHistory(keepLast=4)
                           ↓
               (可选) Compact() → 物理压缩 .jsonl

Context 超限时:
   forceCompression() → 先压缩丢弃部分到 L2 → 再丢弃消息
```

---

## 预期效果

1. **减少幻觉**：结构化摘要保留关键事实和决策，避免纯文本摘要的信息丢失
2. **更好的连续性**：`current_task` 和 `pending_items` 帮助 Agent 理解当前状态
3. **更智能的压缩**：强制压缩前先提取关键信息到 L2，不再粗暴丢弃
4. **可配置的保留策略**：通过 `short_term_messages` 和 `long_term_retention_days` 调整

---

## 兼容性

- **向后兼容**：旧会话没有 `.structured.json` 文件时，`GetStructuredSummary` 返回 `nil`，不影响现有功能
- **渐进式升级**：新对话会自动创建 L2 结构化摘要
- **清除记忆**：`/clear` 命令会同时清除 L1 历史和 L2 结构化摘要
