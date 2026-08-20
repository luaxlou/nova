# Nova AI Coding 协作指南

本文档用于让 AI Agent 在 `nova` 仓库内快速建立上下文，并围绕统一 starter 设计哲学稳定交付。

## 一句话定位

`nova` 是应用侧框架仓库，提供 starter / sdk 能力，目标是让项目持续获得边界清晰、可组合、低心智负担和稳定约定。

## 引入后要拿到的收益

- 用统一初始化范式替代分散样板代码
- 通过 starter 组合减少重复封装
- 让团队在新项目与存量项目里使用一致接入语言
- 把接入沉淀为可演进工程基线

## 修改边界

允许改动：

- `starter/*`
- `examples/*`
- `docs/*`

禁止越界：

- 不实现与 starter / sdk 无关的系统能力
- 不引入与当前任务目标无关的额外职责
- 不做无关的大规模重构

## 标准执行顺序

1. 阅读流程卡：
   - 新项目：`docs/quickstart_new_project.md`
   - 存量项目：`docs/quickstart_existing_project.md`
2. 按固定结构组织输出内容：实施计划、改动文件清单、收益说明、验证结果。
3. 执行统一验证命令并回报结果。

## 统一验证命令

```bash
go test ./...
go vet ./...
```

## 快速索引

- `starter/novaconfig`
- `starter/novagin`
- `starter/novamysql`
- `starter/novagorm`
- `starter/novaredis`
- `starter/novawebsocket`
- `examples/simple-app`
- `docs/starter_conventions.md`
- `docs/starter_composition_matrix.md`
