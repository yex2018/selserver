package models

import (
	"database/sql"
	"errors"
	"time"

	db "github.com/yex2018/selserver/database"
)

type Evaluation struct {
	Evaluation_id int64      `json:"evaluation_id"`
	Name          string     `json:"name"`
	Category      string     `json:"category"`
	Abstract      string     `json:"abstract"`
	Details       string     `json:"details"`
	Price         string     `json:"price"`
	Page_number   int        `json:"page_number"`
	Person_count  int        `json:"person_count"`
	Picture       string     `json:"picture"`
	Sample_report string     `json:"sample_report"`
	Key_name      string     `json:"key_name"`
	MaxIndex      int        `json:"maxIndex"`
	Questions     []Question `json:"questions"`
}

type Question struct {
	Question_id    int64  `json:"question_id"`
	Question_index int    `json:"question_index"`
	Content        string `json:"content"`
}

type UserEvaluation struct {
	User_evaluation_id  int64     `json:"user_evaluation_id"`
	Evaluation_id       int64     `json:"evaluation_id"`
	User_id             int64     `json:"user_id"`
	Child_id            int64     `json:"child_id"`
	Evaluation_time     time.Time `json:"evaluation_time"`
	Current_question_id int       `json:"current_question_id"`
	Data_result         string    `json:"data_result"`
	Report_result       string    `json:"report_result"`
}

type UserQuestion struct {
	User_question_id   int64  `json:"user_question_id"`
	User_evaluation_id int64  `json:"user_evaluation_id"`
	Question_id        int64  `json:"question_id"`
	Answer             string `json:"answer"`
}

var g_Evaluations []Evaluation

func init() {
	rowEvaluations, err := db.SqlDB.Query("SELECT evaluation_id,name,category,abstract,details,price,page_number,person_count,picture,sample_report,key_name FROM evaluation")
	if err != nil {
		return
	}
	defer rowEvaluations.Close()

	for rowEvaluations.Next() {
		var eva Evaluation

		err = rowEvaluations.Scan(&eva.Evaluation_id, &eva.Name, &eva.Category, &eva.Abstract, &eva.Details, &eva.Price, &eva.Page_number, &eva.Person_count, &eva.Picture, &eva.Sample_report, &eva.Key_name)
		if err != nil {
			return
		}

		rowQuestions, err := db.SqlDB.Query("SELECT question_id,question_index,content from question WHERE evaluation_id=? ORDER BY question_index", eva.Evaluation_id)
		if err != nil {
			return
		}
		defer rowQuestions.Close()

		eva.MaxIndex = 0
		for rowQuestions.Next() {
			var question Question

			err = rowQuestions.Scan(&question.Question_id, &question.Question_index, &question.Content)
			if err != nil {
				return
			}

			eva.Questions = append(eva.Questions, question)
			eva.MaxIndex++
		}

		g_Evaluations = append(g_Evaluations, eva)
	}
}

// GetEvaluations 获取测评列表
func GetEvaluations() (evaluations []Evaluation) {
	return g_Evaluations
}

// QryEvaluationById 获取单个测评
func QryEvaluationById(evaluation_id int64) (result *Evaluation, err error) {
	for i, _ := range g_Evaluations {
		if g_Evaluations[i].Evaluation_id == evaluation_id {
			result = &g_Evaluations[i]
			err = nil
			return
		}
	}

	return result, errors.New("无效的参数")
}

// UpdatePersonCountForEvaluation 更新测评已测人数
func UpdatePersonCountForEvaluation(evaluation_id int64) (err error) {
	value, err := QryEvaluationById(evaluation_id)
	if err == nil {
		personCount := value.Person_count + 1

		_, err = db.SqlDB.Exec("UPDATE evaluation SET person_count=? WHERE evaluation_id=?", personCount, evaluation_id)
		if err != nil {
			return err
		}

		value.Person_count = personCount
		return nil
	}

	return err
}

// AddUserEvaluation 增加用户测评
func AddUserEvaluation(evaluation_id int64, user_id int64, child_id int64, evaluation_time time.Time, current_question_id int) (id int64, err error) {
	rs, err := db.SqlDB.Exec("INSERT INTO user_evaluation(evaluation_id,user_id,child_id,evaluation_time,current_question_id) VALUES (?, ?, ?, ?, ?)", evaluation_id, user_id, child_id, evaluation_time, current_question_id)
	if err != nil {
		return 0, err
	}
	id, err = rs.LastInsertId()
	return
}

// QryEvaluationById 根据ID获取单个用户测评
func QryUserEvaluationById(user_evaluation_id int64) (result UserEvaluation, err error) {
	result.User_evaluation_id = user_evaluation_id

	err = db.SqlDB.QueryRow("SELECT evaluation_id,user_id,child_id,evaluation_time,current_question_id,IFNULL(data_result,''),IFNULL(report_result,'') FROM user_evaluation WHERE user_evaluation_id=?", user_evaluation_id).Scan(&result.Evaluation_id, &result.User_id, &result.Child_id, &result.Evaluation_time, &result.Current_question_id, &result.Data_result, &result.Report_result)
	return
}

// QryUserEvaluationByUserId 根据用户ID获取用户测评
func QryUserEvaluationByUserId(user_id int64) (userevaluations []UserEvaluation, err error) {
	rows, err := db.SqlDB.Query("SELECT user_evaluation_id,evaluation_id,child_id,evaluation_time,current_question_id,IFNULL(data_result,''),IFNULL(report_result,'') FROM user_evaluation WHERE user_id=? ORDER BY evaluation_time DESC", user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userevaluation UserEvaluation
		userevaluation.User_id = user_id
		err = rows.Scan(&userevaluation.User_evaluation_id, &userevaluation.Evaluation_id, &userevaluation.Child_id, &userevaluation.Evaluation_time, &userevaluation.Current_question_id, &userevaluation.Data_result, &userevaluation.Report_result)
		if err != nil {
			return nil, err
		}
		userevaluations = append(userevaluations, userevaluation)
	}

	return userevaluations, err
}

// QryUserEvaluationByChildId 根据儿童ID获取用户测评
func QryUserEvaluationByChildId(child_id int64) (userevaluations []UserEvaluation, err error) {
	rows, err := db.SqlDB.Query("SELECT user_evaluation_id,evaluation_id,user_id,evaluation_time,current_question_id,IFNULL(data_result,''),IFNULL(report_result,'') FROM user_evaluation WHERE child_id=? AND current_question_id=-1 ORDER BY evaluation_time DESC", child_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userevaluation UserEvaluation
		userevaluation.Child_id = child_id
		err = rows.Scan(&userevaluation.User_evaluation_id, &userevaluation.Evaluation_id, &userevaluation.User_id, &userevaluation.Evaluation_time, &userevaluation.Current_question_id, &userevaluation.Data_result, &userevaluation.Report_result)
		if err != nil {
			return nil, err
		}
		userevaluations = append(userevaluations, userevaluation)
	}

	return userevaluations, err
}

// QryUserQuestion 获取用户测评题目
func QryUserQuestion(user_evaluation_id int64, question_id int64) (result UserQuestion, err error) {
	result.User_question_id = 0
	result.User_evaluation_id = user_evaluation_id
	result.Question_id = question_id

	err = db.SqlDB.QueryRow("SELECT user_question_id,answer FROM user_question WHERE user_evaluation_id=? AND question_id=?", user_evaluation_id, question_id).Scan(&result.User_question_id, &result.Answer)
	return
}

// UpdateUserQuestion 更新用户测评题目
func UpdateUserQuestion(user_evaluation_id int64, user_question_id int64, question_id int64, question_index int, answer string) (err error) {
	var current_question_id int

	err = db.SqlDB.QueryRow("SELECT current_question_id FROM user_evaluation WHERE user_evaluation_id=?", user_evaluation_id).Scan(&current_question_id)
	if err != nil {
		return err
	}

	if current_question_id < question_index {
		_, err = db.SqlDB.Exec("UPDATE user_evaluation SET current_question_id=? WHERE user_evaluation_id=?", question_index, user_evaluation_id)
		if err != nil {
			return err
		}
	}

	if user_question_id <= 0 {
		err = db.SqlDB.QueryRow("SELECT user_question_id FROM user_question WHERE user_evaluation_id=? AND question_id=?", user_evaluation_id, question_id).Scan(&user_question_id)
		if err != nil && err != sql.ErrNoRows {
			return err
		}
	}

	if user_question_id > 0 {
		_, err = db.SqlDB.Exec("UPDATE user_question SET answer=? WHERE user_question_id=?", answer, user_question_id)
		if err != nil {
			return err
		}
	} else {
		_, err = db.SqlDB.Exec("INSERT INTO user_question(user_evaluation_id, question_id, answer) VALUES (?, ?, ?)", user_evaluation_id, question_id, answer)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateCurrentQuestionIdForUserEvaluation 更新用户测评的current_question_id
func UpdateCurrentQuestionIdForUserEvaluation(user_evaluation_id int64, current_question_id int) (err error) {
	_, err = db.SqlDB.Exec("UPDATE user_evaluation SET current_question_id=? WHERE user_evaluation_id=?", current_question_id, user_evaluation_id)
	return
}
