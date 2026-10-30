# go-todo

Go 1.22+ 的 Todo REST API，只使用标准库 net/http，不引入第三方依赖。

## 常用命令
- 运行：`make run`
- 测试：`make test`
- 静态检查：`make lint`
- 一键检查（格式化 + 静态检查 + 测试）：`make check`

## 目录结构
- cmd/server：程序入口，只做组装，不写业务逻辑
- internal/todo：业务代码，分 handler（HTTP 层）/ service（业务层）/ store（存储层）

## 代码约定
- 错误必须包装后返回：`fmt.Errorf("做什么失败: %w", err)`，禁止吞掉错误
- handler 不得直接调用 store，必须经过 service
- 所有接口统一返回 JSON：`{"code":0,"msg":"ok","data":...}`
- 单元测试用表驱动写法，测试文件与源文件放同一目录

## 工作方式
- 改动前先给出方案，我确认后再写代码
- 完成后必须运行 `make check`，并把结果贴出来
- 不要修改与当前任务无关的文件

## 禁止事项
- 不要引入新的第三方依赖，除非我明确同意
- 不要删除或跳过已有测试