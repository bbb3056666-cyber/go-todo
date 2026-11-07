# go-todo

Go 1.22+ 的 Todo REST API，只使用标准库 net/http，不引入第三方依赖。
数据存在内存中，重启丢失（持久化不在当前范围内）。

## 常用命令
- 运行：`make run`
- 测试：`make test`
- 静态检查：`make lint`
- 一键检查（格式化 + 静态检查 + 测试）：`make check`。其中的格式化会直接改写文件，运行后用 `git status` 确认改动范围。
- 找不到 make 时，分别运行 `gofmt -l .`、`go vet ./...`、`go test ./... -race -cover`，并说明用了替代命令。

## 开发环境
- Windows + PowerShell。发 HTTP 请求用 `curl.exe`（必须带 .exe）或 `Invoke-RestMethod`。
- 启动服务做验证后，必须关闭该进程，不要一直占用 8080 端口。
- 换行符统一为 LF（见 .gitattributes）。

## 目录结构
- cmd/server：程序入口，只做组装，不写业务逻辑
- internal/todo：业务代码，分 handler（HTTP 层）/ service（业务层）/ store（存储层）
- docs/requirements：原始需求
- docs/research：调研结论
- docs/specs：Spec；docs/templates/spec.md 是模板
- docs/plans：Plan
- docs/review-checklist.md：审查清单；docs/backlog.md：审查发现但暂不处理的问题

## 代码约定
- 错误必须包装后返回：`fmt.Errorf("做什么失败: %w", err)`，禁止吞掉错误
- handler 不得直接调用 store，必须经过 service
- 所有业务接口统一返回 JSON：`{"code":0,"msg":"ok","data":...}`（路由层的 404/405 保持标准库默认行为）
- 接收 JSON 的接口：字段名区分大小写，拒绝未知字段；必须限制请求体大小（现为 64KB）
- HTTP 服务器必须设置超时（见 cmd/server）
- 单元测试用表驱动写法，测试文件与源文件放同一目录

## 开发流程
收到开发需求时，按 `.agents/skills/development/SKILL.md` 执行。

- 默认连续完成整个已批准的任务，自行执行测试先行和验证。
- 大需求实现并提交后，由我在新会话调用 `$review`；我决定哪些问题修复、记入 `docs/backlog.md` 或忽略。AI 批量修复我确认的问题并报告结果。
- 我检查结果和 diff；提交、推送、PR 和合并由我处理。未经明确指示不要自行交付。
- 不要修改与当前任务无关的文件。发现 Spec 或 Plan 有问题，停下告诉我，不要自行改需求。

## 禁止事项
- 不要引入新的第三方依赖，除非我明确同意
- 不要删除或跳过已有测试
- 未经我同意，不要修改 AGENTS.md、Makefile、go.mod
