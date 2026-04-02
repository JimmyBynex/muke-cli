package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const baseURL = "https://mooc2.uestc.edu.cn"
const domain = "mooc2.uestc.edu.cn"

// Domain 返回平台域名，供 client.GetCookies 使用
func Domain() string { return domain }

// get 发带 Cookie 的 GET 请求
func get(url, cookie string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	setHeaders(req, cookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// post 发带 Cookie 的 POST/PUT 请求
func post(method, url, cookie string, body any) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	setHeaders(req, cookie)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s\n%s", resp.StatusCode, url, body)
	}
	return io.ReadAll(resp.Body)
}

func setHeaders(req *http.Request, cookie string) {
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
}

// Courses 获取我的课程列表
func Courses(cookie string) (*CoursesResp, error) {
	url := baseURL + "/api/my-courses?conditions=%7B%22classify_type%22%3A%22recently_started%22%7D&page_size=20&page=1"
	data, err := get(url, cookie)
	if err != nil {
		return nil, err
	}
	var result CoursesResp
	return &result, json.Unmarshal(data, &result)
}

// Exams 获取某课程的考试列表
func Exams(courseID, cookie string) (*ExamsResp, error) {
	data, err := get(fmt.Sprintf("%s/api/courses/%s/exams", baseURL, courseID), cookie)
	if err != nil {
		return nil, err
	}
	var result ExamsResp
	return &result, json.Unmarshal(data, &result)
}

// GetExam 获取考试题目
func GetExam(examID, cookie string) (*ExamDistribute, error) {
	data, err := get(fmt.Sprintf("%s/api/exams/%s/distribute", baseURL, examID), cookie)
	if err != nil {
		return nil, err
	}
	var result ExamDistribute
	return &result, json.Unmarshal(data, &result)
}

// GetSubmission 获取当前进行中的 submission（没有则返回 nil, nil）
func GetSubmission(examID, cookie string) (*Submission, error) {
	data, err := get(fmt.Sprintf("%s/api/exams/%s/submissions/storage", baseURL, examID), cookie)
	if err != nil {
		// 404 表示没有进行中的 submission，其他错误才往上抛
		if strings.Contains(err.Error(), "HTTP 404") {
			return nil, nil
		}
		return nil, err
	}
	var result Submission
	return &result, json.Unmarshal(data, &result)
}

// CreateSubmission 用题目数据创建新的 submission
func CreateSubmission(examID, cookie string, dist *ExamDistribute) (*Submission, error) {
	// 构造每道小题的初始状态（空答案）
	var subjects []SubjectState
	for _, group := range dist.Subjects {
		for _, s := range group.SubSubjects {
			parentID := 0
			if s.ParentID != nil {
				parentID = *s.ParentID
			}
			subjects = append(subjects, SubjectState{
				SubjectID:        s.ID,
				SubjectUpdatedAt: s.LastUpdatedAt,
				AnswerOptionIDs:  []int{},
				ParentID:         parentID,
			})
		}
	}

	body := map[string]any{
		"exam_paper_instance_id": dist.ExamPaperInstanceID,
		"exam_submission_id":     nil,
		"subjects":               subjects,
		"progress": Progress{
			AnsweredNum:   0,
			TotalSubjects: len(subjects),
		},
	}

	data, err := post("POST", fmt.Sprintf("%s/api/exams/%s/submissions/storage", baseURL, examID), cookie, body)
	if err != nil {
		return nil, err
	}
	var result Submission
	return &result, json.Unmarshal(data, &result)
}

// SaveAnswers 保存答案（PUT）
func SaveAnswers(submissionID int, cookie string, req SaveAnswersReq) error {
	url := fmt.Sprintf("%s/api/exams/submissions/%d/multiple-subjects", baseURL, submissionID)
	_, err := post("PUT", url, cookie, req)
	return err
}

// FinalSubmit 最终交卷
func FinalSubmit(examID, cookie string, req FinalSubmitReq) ([]byte, error) {
	return post("POST", fmt.Sprintf("%s/api/exams/%s/submissions", baseURL, examID), cookie, req)
}

// GetResults 获取成绩
func GetResults(examID, cookie string) (*ResultResp, error) {
	data, err := get(fmt.Sprintf("%s/api/exams/%s/submissions", baseURL, examID), cookie)
	if err != nil {
		return nil, err
	}
	var result ResultResp
	return &result, json.Unmarshal(data, &result)
}
