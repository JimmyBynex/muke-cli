帮用户完成英语课考试提交，按以下步骤执行。不要自行探测文件或系统状态（禁止运行 cat、ls、read 等探测命令），直接按步骤运行 muke 命令，让 CLI 自己处理错误。

遇到 HTTP 404 或 401 错误时，不要让用户去浏览器手动操作，直接运行：
```bash
go run ./cmd/muke refresh
```
然后重试失败的命令。

## 步骤 1：列出英语课考试

如果输出提示"请先运行 muke setup"，则先运行 `go run ./cmd/muke setup`，将课程列表格式化为 markdown 表格，让用户选择英语课程 id，之后再继续。

```bash
go run ./cmd/muke exams
```

将命令输出格式化为 markdown 表格展示给用户，询问选哪场考试的 id。

## 步骤 2：获取题目结构

```bash
go run ./cmd/muke exam <exam_id>
```

记录每道题的 subject_id、题型、每个选项的字母和 option_id。提交时会自动创建 submission，无需手动在浏览器开始考试。

## 步骤 3：读取答案文件

询问用户答案文件路径，读取文件内容，提取每道题的答案。

## 步骤 4：生成 answers.json

根据步骤 2 的 subject_id / option_id 和步骤 3 的答案，生成 answers.json：

- 单选题：{"subject_id": <id>, "option_ids": [<option_id>]}
- 填空题：{"subject_id": <id>, "text_answers": ["答案1", "答案2", ...]}

只包含有答案的题目。subject_id 和 option_id 必须来自步骤 2 的输出，不能猜测。

将内容写入 ./answers.json，展示给用户确认。

## 步骤 5：提交答案

用户确认后运行：

```bash
go run ./cmd/muke submit <exam_id> ./answers.json
```

## 步骤 6：查看成绩

```bash
go run ./cmd/muke result <exam_id>
```

输出成绩给用户。
