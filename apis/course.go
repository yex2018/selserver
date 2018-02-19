package apis

import (
	"errors"
	"net/http"
	"time"

	"github.com/Fengxq2014/aliyun/vod"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/yex2018/selserver/conf"
	"github.com/yex2018/selserver/models"
)

// QryCourse 获取课程列表
func QryCourse(c *gin.Context) {
	mapCategory := make(map[string][]interface{})

	var categories []string
	courses := models.GetCourses()
	if len(courses) > 0 {
		for _, value := range courses {
			mapCourse := make(map[string]interface{})
			mapCourse["course_id"] = value.Course_id
			mapCourse["name"] = value.Name
			mapCourse["abstract"] = value.Abstract
			mapCourse["price"] = value.Price
			mapCourse["person_count"] = value.Person_count
			mapCourse["picture"] = value.Picture

			if _, ok := mapCategory[value.Category]; !ok {
				categories = append(categories, value.Category)
			}
			mapCategory[value.Category] = append(mapCategory[value.Category], mapCourse)
		}
	}

	var data []map[string]interface{}
	for _, categoriy := range categories {
		mapData := make(map[string]interface{})
		mapData["category"] = categoriy
		mapData["data"] = mapCategory[categoriy]
		data = append(data, mapData)
	}
	c.JSON(http.StatusOK, models.Result{Data: &data})
}

// GetCourseByID 根据id获取课程信息
func GetCourseByID(c *gin.Context) {
	type param struct {
		Course_id int64 `form:"course_id" binding:"required"` //测评ID
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	course, err := models.QryCourseById(queryStr.Course_id)
	if err != nil {
		c.Error(err)
		return
	}

	mapData := make(map[string]interface{})
	mapData["course_id"] = course.Course_id
	mapData["name"] = course.Name
	mapData["category"] = course.Category
	mapData["abstract"] = course.Abstract
	mapData["details"] = course.Details
	mapData["price"] = course.Price
	mapData["picture"] = course.Picture

	c.JSON(http.StatusOK, models.Result{Data: &mapData})
	return
}

// GetResource 获取课程资源
func GetCourseResource(c *gin.Context) {
	type param struct {
		Course_id int64 `form:"course_id" binding:"required"` //测评ID
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	course, err := models.QryCourseById(queryStr.Course_id)
	if err != nil {
		c.Error(err)
		return
	}

	var data []map[string]interface{}
	for _, cresource := range course.CResources {
		mapData := make(map[string]interface{})
		mapData["resource_id"] = cresource.Resource.Resource_id
		mapData["name"] = cresource.Resource.Name
		mapData["type"] = cresource.Resource.Type
		mapData["url"] = cresource.Resource.Url
		mapData["index"] = cresource.Index
		mapData["free"] = cresource.Free
		data = append(data, mapData)
	}
	c.JSON(http.StatusOK, models.Result{Data: &data})
}

// AddUserCourse 增加用户课程
func AddUserCourse(c *gin.Context) {
	type param struct {
		Course_id int64 `form:"course_id" binding:"required"` //关联课程ID
		User_id   int64 `form:"user_id" binding:"required"`   //关联用户ID
	}

	var postStr param
	if c.ShouldBindWith(&postStr, binding.JSON) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	_, err := models.AddUserCourse(postStr.Course_id, postStr.User_id, time.Now())
	if err != nil {
		c.JSON(http.StatusOK, models.Result{Res: 1, Msg: "增加用户课程失败" + err.Error(), Data: nil})
		return
	}

	err = models.UpdatePersonCountForCourse(postStr.Course_id)
	if err != nil {
		c.JSON(http.StatusOK, models.Result{Res: 1, Msg: "增加用户课程失败" + err.Error(), Data: nil})
		return
	}

	c.JSON(http.StatusOK, models.Result{Res: 0, Msg: "", Data: nil})
}

// QryUserCourse 查看用户单个课程
func QryUserCourse(c *gin.Context) {
	type param struct {
		Course_id int64 `form:"course_id" binding:"required"` //课程ID
		User_id   int64 `form:"user_id" binding:"required"`   //用户ID
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	user_course_id, err := models.QryUserCourse(queryStr.Course_id, queryStr.User_id)
	if err != nil {
		c.JSON(http.StatusOK, models.Result{Res: 1, Msg: "获取用户课程失败" + err.Error(), Data: nil})
		return
	}

	c.JSON(http.StatusOK, models.Result{Res: 0, Msg: "", Data: user_course_id})
}

// QryMyCourse 获取本人课程列表
func QryMyCourse(c *gin.Context) {
	type param struct {
		User_id int64 `form:"user_id" binding:"required"` //用户ID
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	usercourses, err := models.QryUserCourseByUserId(queryStr.User_id)
	if err != nil {
		c.Error(err)
		return
	}

	var data []map[string]interface{}
	if len(usercourses) > 0 {
		for _, usercourse := range usercourses {
			course, err := models.QryCourseById(usercourse.Course_id)
			if err == nil {
				mapData := make(map[string]interface{})

				mapData["course_id"] = course.Course_id
				mapData["name"] = course.Name
				mapData["abstract"] = course.Abstract
				mapData["picture"] = course.Picture
				mapData["price"] = course.Price
				mapData["user_course_id"] = usercourse.User_course_id
				mapData["course_time"] = usercourse.Course_time

				data = append(data, mapData)
			}
		}
	}

	c.JSON(http.StatusOK, models.Result{Data: &data})
}

// GetVideo 获取视频播放地址
func GetVideo(c *gin.Context) {
	type param struct {
		Media   string `form:"media" binding:"required"`    //视频ID
		Formats string `form:"formmats" binding:"required"` //视频流格式，多个用逗号分隔，支持格式mp4,m3u8,mp3
	}
	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}
	res := models.Result{}

	playInfo, err := vod.NewAliyunVod(conf.Config.Access_key_id, conf.Config.Access_secret).GetPlayInfo(queryStr.Media, queryStr.Formats, "")
	if err != nil {
		res.Res = 1
		res.Msg = err.Error()
		res.Data = nil
		c.JSON(http.StatusOK, res)
		return
	}
	res.Res = 0
	res.Msg = ""
	res.Data = map[string]string{"playAuth": playInfo.PlayInfoList.PlayInfo[0].PlayURL, "coverurl": playInfo.VideoBase.CoverURL}

	c.JSON(http.StatusOK, res)
}
