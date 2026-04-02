package api

// ── 课程列表 ──────────────────────────────────────────────

type CoursesResp struct {
	Courses []Course `json:"courses"`
}

type Course struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Semester   struct {
		Name string `json:"name"`
	} `json:"semester"`
	Instructors []struct {
		Name string `json:"name"`
	} `json:"instructors"`
}

// ── 考试列表 ──────────────────────────────────────────────

type ExamsResp struct {
	Exams []Exam `json:"exams"`
}

type Exam struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	IsInProgress bool   `json:"is_in_progress"`
	IsClosed     bool   `json:"is_closed"`
	IsStarted    bool   `json:"is_started"`
}

// ── 题目结构 ──────────────────────────────────────────────

type ExamDistribute struct {
	ExamPaperInstanceID int       `json:"exam_paper_instance_id"`
	Subjects            []Subject `json:"subjects"`
}

type Subject struct {
	ID            int       `json:"id"`
	Type          string    `json:"type"` // analysis / single_selection / fill_in_blank
	Description   string    `json:"description"`
	Options       []Option  `json:"options"`
	SubSubjects   []Subject `json:"sub_subjects"`
	ParentID      *int      `json:"parent_id"`
	Sort          int       `json:"sort"`
	LastUpdatedAt string    `json:"last_updated_at"`
}

type Option struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	Sort    int    `json:"sort"`
}

// ── Submission ────────────────────────────────────────────

type Submission struct {
	ID          int            `json:"id"`
	InstanceID  int            `json:"instance_id"`
	LeftTime    float64        `json:"left_time"`
	SubmittedAt *string        `json:"submitted_at"`
	Data        SubmissionData `json:"submission_data"`
}

type SubmissionData struct {
	Subjects []SubjectState `json:"subjects"`
	Progress Progress       `json:"progress"`
}

type FillAnswer struct {
	Content string `json:"content"`
	Sort    int    `json:"sort"`
}

type SubjectState struct {
	SubjectID        int          `json:"subject_id"`
	SubjectUpdatedAt string       `json:"subject_updated_at,omitempty"`
	AnswerOptionIDs  []int        `json:"answer_option_ids,omitempty"`
	Answers          []FillAnswer `json:"answers,omitempty"`
	ParentID         int          `json:"parent_id"`
}

type Progress struct {
	AnsweredNum   int `json:"answered_num"`
	TotalSubjects int `json:"total_subjects"`
}

// ── 保存答案（PUT multiple-subjects）────────────────────────

type SaveAnswersReq struct {
	ExamPaperInstanceID int          `json:"exam_paper_instance_id"`
	SubjectsAnswers     []AnswerItem `json:"subjects_answers"`
	PlayRecord          struct{}     `json:"play_record"`
	Progress            Progress     `json:"progress"`
}

type AnswerItem struct {
	Index     int          `json:"index"`
	SubjectID int          `json:"subject_id"`
	Answer    AnswerDetail `json:"answer"`
}

type AnswerDetail struct {
	SubjectID        int          `json:"subject_id"`
	SubjectUpdatedAt string       `json:"subject_updated_at,omitempty"`
	AnswerOptionIDs  []int        `json:"answer_option_ids,omitempty"`
	Answers          []FillAnswer `json:"answers,omitempty"`
	ParentID         int          `json:"parent_id"`
}

// ── 最终交卷（POST submissions）──────────────────────────────

type FinalSubmitReq struct {
	ExamPaperInstanceID int            `json:"exam_paper_instance_id"`
	ExamSubmissionID    int            `json:"exam_submission_id"`
	Subjects            []SubjectState `json:"subjects"`
	Progress            Progress       `json:"progress"`
	Reason              string         `json:"reason"`
}

// ── Claude 生成的答案文件格式 ─────────────────────────────────

// AnswerInput 是 answers.json 里每条记录的格式
// 单选题：填 option_ids（一个元素）
// 多选题：填 option_ids（多个元素）
// 填空题：填 text_answers（每空一个字符串）
type AnswerInput struct {
	SubjectID   int      `json:"subject_id"`
	OptionIDs   []int    `json:"option_ids,omitempty"`
	TextAnswers []string `json:"text_answers,omitempty"`
}

// ── 成绩 ──────────────────────────────────────────────────

type ResultResp struct {
	ExamScore     *float64     `json:"exam_score"`
	ExamScoreRule string       `json:"exam_score_rule"`
	Submissions   []SubRecord  `json:"submissions"`
}

type SubRecord struct {
	ID          int     `json:"id"`
	Score       string  `json:"score"`
	SubmittedAt string  `json:"submitted_at"`
}
