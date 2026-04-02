package main

import (
	"encoding/json"
	"fmt"
	"muke-cli/internal/api"
	"muke-cli/internal/client"
	"muke-cli/internal/config"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {
	case "setup":
		cmdSetup()
		return
	case "refresh":
		cmdRefresh()
		return
	}

	cookie, err := client.GetCookies(api.Domain())
	if err != nil {
		fmt.Fprintln(os.Stderr, "获取 Cookie 失败:", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "exams":
		cmdExams(cookie)
	case "exam":
		requireArg("muke exam <exam-id>")
		cmdExam(os.Args[2], cookie)
	case "submit":
		requireArg("muke submit <exam-id> <answers.json>")
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "用法: muke submit <exam-id> <answers.json>")
			os.Exit(1)
		}
		cmdSubmit(os.Args[2], os.Args[3], cookie)
	case "result":
		requireArg("muke result <exam-id>")
		cmdResult(os.Args[2], cookie)
	default:
		usage()
	}
}

// cmdRefresh 清除本地 cookie 缓存，重新从浏览器获取
func cmdRefresh() {
	client.ClearSession()
	fmt.Println("已清除 cookie 缓存，重新获取中...")
	cookie, err := client.GetCookies(api.Domain())
	if err != nil {
		fmt.Fprintln(os.Stderr, "获取 Cookie 失败（请确保浏览器已登录并触发过请求）:", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Cookie 已刷新（长度 %d）\n", len(cookie))
}

// cmdSetup 列出所有课程，让用户选定英语课程并保存
// 用法：muke setup [course-id]
// 若不传 course-id，则打印课程列表后退出，等待用户再次调用并传入 id
func cmdSetup() {
	cookie, err := client.GetCookies(api.Domain())
	if err != nil {
		fmt.Fprintln(os.Stderr, "获取 Cookie 失败:", err)
		os.Exit(1)
	}

	resp, err := api.Courses(cookie)
	fatal(err)

	// 如果直接传了 course-id，直接保存
	if len(os.Args) >= 3 {
		courseID := os.Args[2]
		cfg := &config.Config{EnglishCourseID: courseID}
		if err := config.Save(cfg); err != nil {
			fmt.Fprintln(os.Stderr, "保存配置失败:", err)
			os.Exit(1)
		}
		fmt.Printf("✅ 已保存，英语课程 id = %s\n", courseID)
		return
	}

	// 否则打印课程列表，提示用户传入 id
	fmt.Printf("课程列表（共 %d 门）：\n", len(resp.Courses))
	for i, c := range resp.Courses {
		instructor := ""
		if len(c.Instructors) > 0 {
			instructor = " 教师:" + c.Instructors[0].Name
		}
		fmt.Printf("  %2d. [id=%d] %s  学期:%s%s\n", i+1, c.ID, c.Name, c.Semester.Name, instructor)
	}
	fmt.Println("\n请运行: muke setup <course-id>")
}

// getEnglishCourseID 从配置文件读取课程 ID，没有则提示运行 setup
func getEnglishCourseID() string {
	cfg, err := config.Load()
	if err != nil || cfg.EnglishCourseID == "" {
		fmt.Fprintln(os.Stderr, "未配置英语课程，请先运行: muke setup")
		os.Exit(1)
	}
	return cfg.EnglishCourseID
}

func cmdExams(cookie string) {
	courseID := getEnglishCourseID()
	resp, err := api.Exams(courseID, cookie)
	fatal(err)
	fmt.Printf("英语课考试列表（共 %d 场）：\n", len(resp.Exams))
	for i, e := range resp.Exams {
		status := "未开始"
		if e.IsInProgress {
			status = "进行中"
		} else if e.IsClosed {
			status = "已结束"
		} else if e.IsStarted {
			status = "已开始"
		}
		endTime := e.EndTime
		if len(endTime) > 10 {
			endTime = endTime[:10]
		}
		fmt.Printf("  %2d. [id=%d] %-30s 状态:%-6s 截止:%s\n", i+1, e.ID, e.Title, status, endTime)
	}
}

func cmdExam(examID, cookie string) {
	dist, err := api.GetExam(examID, cookie)
	if err != nil {
		if isHTTP(err, 404) {
			fmt.Fprintln(os.Stderr, "无法获取题目（HTTP 404）：可能原因：")
			fmt.Fprintln(os.Stderr, "  1. 提交次数已用完")
			fmt.Fprintln(os.Stderr, "  2. 考试未在有效期内")
			fmt.Fprintln(os.Stderr, "  3. Cookie 已过期（可运行 muke refresh）")
			os.Exit(1)
		}
		fatal(err)
	}

	qNum := 1
	for _, group := range dist.Subjects {
		for _, sub := range group.SubSubjects {
			parentID := 0
			if sub.ParentID != nil {
				parentID = *sub.ParentID
			}
			fmt.Printf("Q%d subject_id=%d parent_id=%d type=%s\n",
				qNum, sub.ID, parentID, sub.Type)
			fmt.Printf("  %s\n", stripHTML(sub.Description))

			if sub.Type == "single_selection" || sub.Type == "multiple_selection" {
				for _, opt := range sub.Options {
					label := string(rune('A' + opt.Sort))
					fmt.Printf("  %s [option_id=%d] %s\n", label, opt.ID, stripHTML(opt.Content))
				}
			} else if sub.Type == "fill_in_blank" {
				fmt.Printf("  (填空题，共 %d 空)\n", countBlanks(sub))
			}
			fmt.Println()
			qNum++
		}
	}
}

func cmdSubmit(examID, answersFile, cookie string) {
	data, err := os.ReadFile(answersFile)
	fatal(err)

	var inputs []api.AnswerInput
	if err := json.Unmarshal(data, &inputs); err != nil {
		fmt.Fprintln(os.Stderr, "answers.json 格式错误:", err)
		os.Exit(1)
	}

	dist, err := api.GetExam(examID, cookie)
	fatal(err)

	sub, err := api.GetSubmission(examID, cookie)
	fatal(err)
	if sub == nil {
		fmt.Println("自动创建考试 submission...")
		sub, err = api.CreateSubmission(examID, cookie, dist)
		fatal(err)
		fmt.Printf("已创建 submission_id=%d\n", sub.ID)
	}
	if sub.InstanceID == 0 {
		sub.InstanceID = dist.ExamPaperInstanceID
	}

	subjectMap := map[int]api.Subject{}
	for _, group := range dist.Subjects {
		for _, s := range group.SubSubjects {
			subjectMap[s.ID] = s
		}
	}

	var items []api.AnswerItem
	for i, input := range inputs {
		s, ok := subjectMap[input.SubjectID]
		if !ok {
			fmt.Fprintf(os.Stderr, "警告: subject_id=%d 不存在，跳过\n", input.SubjectID)
			continue
		}
		parentID := 0
		if s.ParentID != nil {
			parentID = *s.ParentID
		}
		detail := api.AnswerDetail{
			SubjectID:        input.SubjectID,
			SubjectUpdatedAt: s.LastUpdatedAt,
			ParentID:         parentID,
		}
		if len(input.TextAnswers) > 0 {
			for j, text := range input.TextAnswers {
				detail.Answers = append(detail.Answers, api.FillAnswer{Content: text, Sort: j})
			}
		} else {
			detail.AnswerOptionIDs = input.OptionIDs
		}
		items = append(items, api.AnswerItem{
			Index:     i,
			SubjectID: input.SubjectID,
			Answer:    detail,
		})
	}

	err = api.SaveAnswers(sub.ID, cookie, api.SaveAnswersReq{
		ExamPaperInstanceID: sub.InstanceID,
		SubjectsAnswers:     items,
		Progress:            api.Progress{AnsweredNum: len(items), TotalSubjects: len(items)},
	})
	fatal(err)

	if len(sub.Data.Subjects) == 0 {
		for _, group := range dist.Subjects {
			for _, s := range group.SubSubjects {
				parentID := 0
				if s.ParentID != nil {
					parentID = *s.ParentID
				}
				sub.Data.Subjects = append(sub.Data.Subjects, api.SubjectState{
					SubjectID:        s.ID,
					SubjectUpdatedAt: s.LastUpdatedAt,
					AnswerOptionIDs:  []int{},
					ParentID:         parentID,
				})
			}
		}
	}

	stateMap := map[int]*api.SubjectState{}
	for i := range sub.Data.Subjects {
		stateMap[sub.Data.Subjects[i].SubjectID] = &sub.Data.Subjects[i]
	}
	for _, input := range inputs {
		if state, ok := stateMap[input.SubjectID]; ok {
			if len(input.TextAnswers) > 0 {
				state.Answers = nil
				for j, text := range input.TextAnswers {
					state.Answers = append(state.Answers, api.FillAnswer{Content: text, Sort: j})
				}
			} else {
				state.AnswerOptionIDs = input.OptionIDs
			}
		}
	}

	_, err = api.FinalSubmit(examID, cookie, api.FinalSubmitReq{
		ExamPaperInstanceID: sub.InstanceID,
		ExamSubmissionID:    sub.ID,
		Subjects:            sub.Data.Subjects,
		Progress:            api.Progress{AnsweredNum: len(items), TotalSubjects: len(items)},
		Reason:              "user",
	})
	fatal(err)

	fmt.Println("✅ 交卷成功")
}

func cmdResult(examID, cookie string) {
	result, err := api.GetResults(examID, cookie)
	fatal(err)

	if result.ExamScore != nil {
		fmt.Printf("最终成绩: %.1f (%s)\n", *result.ExamScore, result.ExamScoreRule)
	}
	for _, s := range result.Submissions {
		fmt.Printf("  提交 #%d  得分: %s  时间: %s\n", s.ID, s.Score, s.SubmittedAt)
	}
}

// ── 工具函数 ──────────────────────────────────────────────

func countBlanks(s api.Subject) int {
	count := 0
	desc := s.Description
	marker := "__blank__"
	for i := 0; i <= len(desc)-len(marker); i++ {
		if desc[i:i+len(marker)] == marker {
			count++
		}
	}
	if count == 0 {
		return 1
	}
	return count
}

func requireArg(usage string) {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "用法:", usage)
		os.Exit(1)
	}
}

func fatal(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isHTTP(err error, code int) bool {
	return err != nil && len(err.Error()) > 8 &&
		err.Error()[:8] == fmt.Sprintf("HTTP %d:", code)
}

func stripHTML(s string) string {
	out := make([]byte, 0, len(s))
	inTag := false
	for i := 0; i < len(s); i++ {
		if s[i] == '<' {
			inTag = true
		} else if s[i] == '>' {
			inTag = false
		} else if !inTag {
			out = append(out, s[i])
		}
	}
	return string(out)
}

func usage() {
	fmt.Println(`muke — 英语课考试 CLI

用法:
  muke setup                            选择并保存英语课程
  muke exams                            列出英语课考试
  muke exam <exam-id>                   查看题目
  muke submit <exam-id> <answers.json>  提交答案
  muke result <exam-id>                 查看成绩`)
}
