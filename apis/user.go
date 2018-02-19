package apis

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/yex2018/selserver/models"
)

func IndexApi(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

// QryUserByOpenId 查询用户信息
func QryUserByOpenId(c *gin.Context) {
	type param struct {
		Openid string `form:"openid" binding:"required"` //测评ID
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	user, err := models.GetUserByOpenId(queryStr.Openid)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, models.Result{Data: &user})
}

// QryUserByUserId 查询用户信息
func QryUserByUserId(c *gin.Context) {
	type param struct {
		User_id int64 `form:"user_id" binding:"required"`
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	user, err := models.GetUserByUserId(queryStr.User_id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, models.Result{Data: &user})
}

// UpdateUser 更新个人中心信息
func UpdateUser(c *gin.Context) {
	type param struct {
		Name       string `form:"name" binding:"required"`
		Gender     int    `form:"gender" binding:"required"`
		Birth_date string `form:"birth_date" binding:"required"`
		User_id    int64  `form:"user_id" binding:"required"`
		Nick_name  string `form:"nick_name" binding:"required"`
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	err := models.UpdateUser(queryStr.User_id, queryStr.Nick_name, queryStr.Gender, queryStr.Name, queryStr.Birth_date)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, models.Result{Res: 0, Msg: "", Data: nil})
}

// QryChild 查询儿童信息
func QryUserChild(c *gin.Context) {
	type param struct {
		User_id int64 `form:"user_id" binding:"required"` //用户ID
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	childs, err := models.GetChildByUserId(queryStr.User_id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, models.Result{Data: childs})
}

// UpdateUserChild 更新用户儿童信息
func UpdateUserChild(c *gin.Context) {
	type param struct {
		User_id       int64  `form:"user_id" binding:"required"`
		Child_id      int64  `form:"child_id"`
		Gender        int    `form:"gender" binding:"required"`
		Name          string `form:"name" binding:"required"`
		Birth_date    string `form:"birth_date" binding:"required"`
		Head_portrait string `form:"head_portrait"`
		Relation      int    `form:"relation" binding:"required"`
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	if queryStr.Child_id != 0 {
		err := models.UpdateChild(queryStr.Child_id, queryStr.Name, queryStr.Gender, queryStr.Birth_date, queryStr.Head_portrait)
		if err != nil {
			c.Error(err)
			return
		}

		err = models.UpdateUcRelation(queryStr.User_id, queryStr.Child_id, queryStr.Relation)
		if err != nil {
			c.Error(err)
			return
		}
	} else {
		queryStr.Child_id = time.Now().Unix()

		err := models.AddChild(queryStr.Child_id, queryStr.Name, queryStr.Gender, queryStr.Birth_date, queryStr.Head_portrait)
		if err != nil {
			c.Error(err)
			return
		}

		err = models.AddUcRelation(queryStr.User_id, queryStr.Child_id, queryStr.Relation)
		if err != nil {
			c.Error(err)
			return
		}
	}

	c.JSON(http.StatusOK, models.Result{Res: 0, Msg: "", Data: nil})
}

// QrySingleChild 查询单个儿童信息
func QrySingleChild(c *gin.Context) {
	type param struct {
		User_id  int64 `form:"user_id" binding:"required"`
		Child_id int64 `form:"child_id" binding:"required"`
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	child, err := models.GetChildById(queryStr.Child_id)
	if err != nil {
		c.Error(err)
		return
	}

	ucRelation, err := models.GetUcRelation(queryStr.User_id, queryStr.Child_id)
	if err != nil {
		c.Error(err)
		return
	}

	mapData := make(map[string]interface{})
	mapData["child_id"] = child.Child_id
	mapData["name"] = child.Name
	mapData["gender"] = child.Gender
	mapData["birth_date"] = child.Birth_date
	mapData["head_portrait"] = child.Head_portrait
	mapData["relation"] = ucRelation.Relation

	c.JSON(http.StatusOK, models.Result{Data: &mapData})
}

// 获取relation列表
func GetRelation(c *gin.Context) {
	mapData := map[string]string{"10": "其它", "1": "爸爸", "2": "妈妈", "3": "爷爷", "4": "奶奶", "5": "外公", "6": "外婆"}
	c.JSON(http.StatusOK, models.Result{Data: &mapData})
}

// QryUserCoupon 查询用户优惠码信息
func QryUserCoupon(c *gin.Context) {
	type param struct {
		User_id     int64  `form:"user_id" binding:"required"`
		Coupon_code string `form:"coupon_code" binding:"required"`
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	userCoupon, err := models.QryUserCoupon(queryStr.User_id, queryStr.Coupon_code)
	if err != nil || userCoupon.Ava_count <= 0 || time.Now().After(userCoupon.Expiry_date) {
		c.JSON(http.StatusOK, models.Result{Res: 0, Msg: "无效的优惠码！", Data: nil})
		return
	}

	c.JSON(http.StatusOK, models.Result{Res: 0, Msg: "", Data: userCoupon.Discount})
}

// UseUserCoupon 使用用户优惠码
func UseUserCoupon(c *gin.Context) {
	type param struct {
		User_id     int64  `form:"user_id" binding:"required"`
		Coupon_code string `form:"coupon_code" binding:"required"`
	}

	var queryStr param
	if c.ShouldBindWith(&queryStr, binding.Query) != nil {
		c.Error(errors.New("参数为空"))
		return
	}

	err := models.UseUserCoupon(queryStr.User_id, queryStr.Coupon_code)
	if err != nil {
		c.JSON(http.StatusOK, models.Result{Res: 0, Msg: "优惠码使用失败", Data: nil})
		return
	}

	c.JSON(http.StatusOK, models.Result{Res: 0, Msg: "优惠码使用成功", Data: nil})
}
