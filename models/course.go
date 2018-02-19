package models

import (
	"database/sql"
	"errors"
	"time"

	db "github.com/yex2018/selserver/database"
)

type Resource struct {
	Resource_id int64  `json:"resource_id" form:"resource_id"`
	Name        string `json:"name" form:"name"`
	Type        int    `json:"type" form:"resource_type"`
	Url         string `json:"url" form:"url"`
}

type CResource struct {
	Cresource_id int64    `json:"cresource_id" form:"cresource_id"`
	Resource     Resource `json:"resource" form:"resource"`
	Index        int      `json:"index" form:"index"`
	Free         int      `json:"free" form:"free"`
}

type Course struct {
	Course_id    int64       `json:"course_id"`
	Name         string      `json:"name"`
	Category     string      `json:"category"`
	Abstract     string      `json:"abstract"`
	Details      string      `json:"details"`
	Price        string      `json:"price"`
	Person_count int         `json:"person_count"`
	Picture      string      `json:"picture"`
	CResources   []CResource `json:"cresources"`
}

type UserCourse struct {
	User_course_id int64     `json:"user_course_id" form:"user_course_id"`
	Course_id      int64     `json:"course_id" form:"course_id"`
	User_id        int64     `json:"user_id" form:"user_id"`
	Course_time    time.Time `json:"course_time" form:"course_time"`
}

var g_mapResources map[int64]Resource
var g_Courses []Course

func init() {
	// Read Resources
	rowResources, err := db.SqlDB.Query("SELECT resource_id,name,type,url FROM resource")
	if err != nil {
		return
	}
	defer rowResources.Close()

	g_mapResources = make(map[int64]Resource)
	for rowResources.Next() {
		var resource Resource

		err = rowResources.Scan(&resource.Resource_id, &resource.Name, &resource.Type, &resource.Url)
		if err != nil {
			return
		}

		g_mapResources[resource.Resource_id] = resource
	}

	// Read Courses
	rowCourses, err := db.SqlDB.Query("SELECT course_id,name,category,abstract,details,price,person_count,picture FROM course")
	if err != nil {
		return
	}
	defer rowCourses.Close()

	for rowCourses.Next() {
		var course Course

		err = rowCourses.Scan(&course.Course_id, &course.Name, &course.Category, &course.Abstract, &course.Details, &course.Price, &course.Person_count, &course.Picture)
		if err != nil {
			return
		}

		rowCResources, err := db.SqlDB.Query("SELECT `cresource_id`,`resource_id`,`index`,`free` FROM cresource WHERE `course_id`=? ORDER BY `index`", course.Course_id)
		if err != nil {
			return
		}
		defer rowCResources.Close()

		for rowCResources.Next() {
			var cresource CResource

			var resource_id int64
			err = rowCResources.Scan(&cresource.Cresource_id, &resource_id, &cresource.Index, &cresource.Free)
			if err != nil {
				return
			}
			cresource.Resource = g_mapResources[resource_id]

			course.CResources = append(course.CResources, cresource)
		}

		g_Courses = append(g_Courses, course)
	}
}

// GetCourses 获取课程列表
func GetCourses() (courses []Course) {
	return g_Courses
}

// QryCourseByID 根据id获取课程信息
func QryCourseById(course_id int64) (result *Course, err error) {
	for i, _ := range g_Courses {
		if g_Courses[i].Course_id == course_id {
			result = &g_Courses[i]
			err = nil
			return
		}
	}

	return result, errors.New("无效的参数")
}

// UpdatePersonCountForCourse 更新测评已测人数
func UpdatePersonCountForCourse(course_id int64) (err error) {
	value, err := QryCourseById(course_id)
	if err == nil {
		personCount := value.Person_count + 1

		_, err = db.SqlDB.Exec("UPDATE course SET person_count=? WHERE course_id=?", personCount, course_id)
		if err != nil {
			return err
		}

		value.Person_count = personCount
		return nil
	}

	return err
}

// AddUserEvaluation 增加用户课程
func AddUserCourse(course_id int64, user_id int64, course_time time.Time) (id int64, err error) {
	rs, err := db.SqlDB.Exec("INSERT INTO user_course(course_id,user_id,course_time) VALUES (?, ?, ?)", course_id, user_id, course_time)
	if err != nil {
		return 0, err
	}
	id, err = rs.LastInsertId()
	return
}

// QryUserCourse 查看用户单个课程
func QryUserCourse(course_id, user_id int64) (user_course_id int64, err error) {
	err = db.SqlDB.QueryRow("SELECT user_course_id from user_course WHERE course_id=? AND user_id=?", course_id, user_id).Scan(&user_course_id)
	if err == nil {
		return user_course_id, err
	} else if err == sql.ErrNoRows {
		return 0, nil
	}

	return 0, err
}

// QryUserCourseByUserId 根据用户ID获取用户课程
func QryUserCourseByUserId(user_id int64) (usercourses []UserCourse, err error) {
	rows, err := db.SqlDB.Query("SELECT user_course_id,course_id,course_time FROM user_course WHERE user_id=? ORDER BY course_time DESC", user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var usercourse UserCourse
		usercourse.User_id = user_id
		err = rows.Scan(&usercourse.User_course_id, &usercourse.Course_id, &usercourse.Course_time)
		if err != nil {
			return nil, err
		}
		usercourses = append(usercourses, usercourse)
	}

	return usercourses, err
}
