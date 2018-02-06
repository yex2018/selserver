package apis

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gin-gonic/gin/binding"

	"github.com/gin-gonic/gin"
	"github.com/yex2018/selserver/conf"
	"github.com/yex2018/selserver/models"
	"github.com/yex2018/selserver/tool"
)

// QryEvaluation 获取测评列表
func QryEvaluation(c *gin.Context) {
	mapCategory := make(map[string][]interface{})

	evaluations := models.GetEvaluations()
	if len(evaluations) > 0 {
		for _, value := range evaluations {
			mapEvaluation := make(map[string]interface{})
			mapEvaluation["evaluation_id"] = value.Evaluation_id
			mapEvaluation["name"] = value.Name
			mapEvaluation["abstract"] = value.Abstract
			mapEvaluation["price"] = value.Price
			mapEvaluation["person_count"] = value.Person_count
			mapEvaluation["picture"] = value.Picture

			mapCategory[value.Category] = append(mapCategory[value.Category], mapEvaluation)
		}
	}

	var data []map[string]interface{}
	for key, value := range mapCategory {
		mapData := make(map[string]interface{})
		mapData["category"] = key
		mapData["data"] = value
		data = append(data, mapData)
	}
	c.JSON(http.StatusOK, models.Result{Data: &data})
}

// QrySingleEvaluation 获取单个测评
func QrySingleEvaluation(c *gin.Context) {
	type param struct {
		EID int `form:"evaluation_id" binding:"required"` //测评ID
	}
	//测评ID
	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	evaluation, err := models.QryEvaluationById(queryStr.EID)
	if err != nil {
		c.Error(err)
		return
	}

	mapData := make(map[string]interface{})
	mapData["evaluation_id"] = evaluation.Evaluation_id
	mapData["name"] = evaluation.Name
	mapData["category"] = evaluation.Category
	mapData["evaluation_id"] = evaluation.Evaluation_id
	mapData["abstract"] = evaluation.Abstract
	mapData["details"] = evaluation.Details
	mapData["price"] = evaluation.Price
	mapData["page_number"] = evaluation.Page_number
	mapData["picture"] = evaluation.Picture
	mapData["person_count"] = evaluation.Person_count
	mapData["sample_report"] = evaluation.Sample_report
	mapData["key_name"] = evaluation.Key_name
	mapData["maxIndex"] = evaluation.MaxIndex

	c.JSON(http.StatusOK, models.Result{Data: &mapData})
	return
}

// AddUserEvaluation 增加用户测评
func AddUserEvaluation(c *gin.Context) {
	type param struct {
		Evaluation_id int `json:"evaluation_id" binding:"required"` //测评ID
		User_id       int `json:"user_id" binding:"required"`       //用户ID
		Child_id      int `json:"child_id" binding:"required"`      //儿童ID
	}

	var postStr param
	if c.ShouldBindWith(&postStr, binding.JSON) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	id, err := models.AddUserEvaluation(postStr.Evaluation_id, postStr.User_id, postStr.Child_id, time.Now(), 0)
	if err != nil {
		c.JSON(http.StatusOK, models.Result{Res: 1, Msg: "增加用户测评失败" + err.Error(), Data: nil})
		return
	}

	mapData := make(map[string]int64)
	mapData["user_evaluation_id"] = id
	c.JSON(http.StatusOK, models.Result{Data: mapData})
}

// QryUserQuestion 获取用户测评题目
func QryUserQuestion(c *gin.Context) {
	type param struct {
		User_evaluation_id int `form:"user_evaluation_id" binding:"required"` //用户测评ID
		Evaluation_id      int `form:"evaluation_id" binding:"required"`      //测评ID
		Question_index     int `form:"question_index" binding:"required"`     //题目号
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	evaluation, err := models.QryEvaluationById(queryStr.Evaluation_id)
	if err != nil {
		c.Error(err)
		return
	}

	if queryStr.Question_index <= 0 || queryStr.Question_index > evaluation.MaxIndex {
		c.Error(errors.New("无效的题目序号"))
		return
	}

	question := evaluation.Questions[queryStr.Question_index-1]
	uq, _ := models.QryUserQuestion(queryStr.User_evaluation_id, question.Question_id)

	mapData := make(map[string]interface{})
	mapData["question_id"] = question.Question_id
	mapData["user_question_id"] = uq.User_question_id
	mapData["content"] = question.Content
	mapData["answer"] = uq.Answer

	c.JSON(http.StatusOK, models.Result{Data: &mapData})
	return
}

// UpdateUserQuestion 更新用户测评题目
func UpdateUserQuestion(c *gin.Context) {
	type param struct {
		User_evaluation_id int    `form:"user_evaluation_id" binding:"required"` //用户测评ID
		User_question_id   int    `form:"user_question_id"`                      //用户题目ID
		Evaluation_id      int    `form:"evaluation_id" binding:"required"`      //测评ID
		Question_id        int    `form:"question_id" binding:"required"`        //题目ID
		Question_Index     int    `form:"question_index" binding:"required"`     //题目序号
		Answer             string `form:"answer" binding:"required"`             //答案
	}

	var queryStr param
	err := c.ShouldBindWith(&queryStr, binding.Query)
	if err != nil {
		c.Error(err)
		return
	}

	evaluation, err := models.QryEvaluationById(queryStr.Evaluation_id)
	if err != nil {
		c.Error(err)
		return
	}

	if queryStr.Question_Index <= 0 || queryStr.Question_Index > evaluation.MaxIndex {
		c.Error(errors.New("无效的题目序号"))
		return
	}

	err = models.UpdateUserQuestion(queryStr.User_evaluation_id, queryStr.User_question_id, queryStr.Question_id, queryStr.Question_Index, queryStr.Answer)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, models.Result{Res: 0, Msg: "", Data: nil})
	return
}

// QryReport 生成报告
func QryReport(c *gin.Context) {
	type param struct {
		User_evaluation_id int    `form:"user_evaluation_id"` //用户测评ID
		TypeId             string `form:"typeid"`             //查看报告1；生成报告0
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	userevaluation, err := models.QryUserEvaluationById(queryStr.User_evaluation_id)
	if err != nil {
		c.Error(err)
		return
	}

	evaluation, err := models.QryEvaluationById(userevaluation.Evaluation_id)
	if err != nil {
		c.Error(err)
		return
	}

	if len(userevaluation.Report_result) <= 0 && queryStr.TypeId == "0" {
		ueIdString := strconv.Itoa(userevaluation.User_evaluation_id)
		reportFileName := evaluation.Key_name + "_" + ueIdString + "_" + time.Now().Format("20060102150405") + ".pdf"

		if runPrint("selreport", ueIdString+","+reportFileName) {
			err = models.UpdatePersonCountForEvaluation(userevaluation.Evaluation_id)
			if err != nil {
				c.Error(err)
				return
			}

			err = models.UpdateCurrentQuestionIdForUserEvaluation(userevaluation.User_evaluation_id, -1)
			if err != nil {
				c.Error(err)
				return
			}

			userevaluation, err = models.QryUserEvaluationById(queryStr.User_evaluation_id)
			if err != nil {
				c.Error(err)
				return
			}
		} else {
			c.Error(errors.New("生成报告失败"))
			return
		}
	}

	mapData := make(map[string]interface{})
	mapData["data_result"] = userevaluation.Data_result

	c.JSON(http.StatusOK, models.Result{Data: &mapData})
	return
}

// SendReport 发送报告
func SendReport(c *gin.Context) {
	type param struct {
		User_evaluation_id int    `form:"user_evaluation_id" binding:"required"` //用户测评ID
		OpenId             string `form:"openid" binding:"required"`             //用户openid
	}
	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	userevaluation, err := models.QryUserEvaluationById(queryStr.User_evaluation_id)
	if err != nil {
		c.Error(err)
		return
	}

	evaluation, err := models.QryEvaluationById(userevaluation.Evaluation_id)
	if err != nil {
		c.Error(err)
		return
	}

	user := models.User{Openid: queryStr.OpenId}
	uses, err := user.GetUserByOpenid()
	if err != nil {
		c.Error(err)
		return
	}

	childInfo, err := models.GetChildById(userevaluation.Child_id)
	if err != nil {
		c.Error(err)
		return
	}

	err = TemplateMessage(queryStr.OpenId, conf.Config.Host+userevaluation.Report_result, evaluation.Name, userevaluation.Evaluation_time.Format("2006-01-02 15:04:05"), uses.Nick_name, childInfo.Name)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, models.Result{Res: 0, Msg: "", Data: nil})
	return
}

// QryMyEvaluation 查询本人测评
func QryMyEvaluation(c *gin.Context) {
	type param struct {
		User_id int `form:"user_id" binding:"required"` //用户ID
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	userevaluations, err := models.QryUserEvaluationByUserId(queryStr.User_id)
	if err != nil {
		c.Error(err)
		return
	}

	var data []map[string]interface{}
	if len(userevaluations) > 0 {
		for _, userevaluation := range userevaluations {
			evaluation, err := models.QryEvaluationById(userevaluation.Evaluation_id)
			if err == nil {
				mapData := make(map[string]interface{})

				mapData["evaluation_id"] = evaluation.Evaluation_id
				mapData["name"] = evaluation.Name
				mapData["abstract"] = evaluation.Abstract
				mapData["picture"] = evaluation.Picture
				mapData["maxIndex"] = evaluation.MaxIndex
				mapData["key_name"] = evaluation.Key_name
				mapData["user_evaluation_id"] = userevaluation.User_evaluation_id
				mapData["current_question_id"] = userevaluation.Current_question_id
				mapData["evaluation_time"] = userevaluation.Evaluation_time

				data = append(data, mapData)
			}
		}
	}

	c.JSON(http.StatusOK, models.Result{Data: &data})
}

// QryEvaluationByChildId 查询所属儿童测评列表
func QryEvaluationByChildId(c *gin.Context) {
	type param struct {
		Child_id int `form:"child_id" binding:"required"`
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	userevaluations, err := models.QryUserEvaluationByChildId(queryStr.Child_id)
	if err != nil {
		c.Error(err)
		return
	}

	var data []map[string]interface{}
	if len(userevaluations) > 0 {
		for _, userevaluation := range userevaluations {
			evaluation, err := models.QryEvaluationById(userevaluation.Evaluation_id)
			if err == nil {
				mapData := make(map[string]interface{})

				mapData["evaluation_id"] = evaluation.Evaluation_id
				mapData["name"] = evaluation.Name
				mapData["abstract"] = evaluation.Abstract
				mapData["picture"] = evaluation.Picture
				mapData["maxIndex"] = evaluation.MaxIndex
				mapData["key_name"] = evaluation.Key_name
				mapData["user_evaluation_id"] = userevaluation.User_evaluation_id
				mapData["current_question_id"] = userevaluation.Current_question_id
				mapData["evaluation_time"] = userevaluation.Evaluation_time

				data = append(data, mapData)
			}
		}
	}

	c.JSON(http.StatusOK, models.Result{Data: &data})
}

func runPrint(cmd string, args ...string) bool {
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", "c:/sel/selreport", os.PathListSeparator, os.Getenv("PATH")))
	tool.Debug("调用report:", cmd)
	tool.Debug(args)
	ecmd := exec.Command(cmd, args...)
	var errorout bytes.Buffer
	var out bytes.Buffer
	ecmd.Stdout = &out
	ecmd.Stderr = &errorout
	err := ecmd.Run()
	if err != nil {
		tool.Error(err)
	}
	if ecmd.ProcessState.Success() {
		return true
	}
	tool.Error(fmt.Sprintf("processstate:%v,out:%v,error:%v", ecmd.ProcessState, out.String(), errorout.String()))
	return false
}
