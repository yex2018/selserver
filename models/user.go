package models

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"

	db "github.com/yex2018/selserver/database"
)

type User struct {
	User_id       int64  `json:"user_id" form:"user_id"`
	Openid        string `json:"openid" form:"openid"`
	Unionid       string `json:"unionid" form:"unionid"`
	Sceneid       int    `json:"sceneid" form:"sceneid"`
	Nick_name     string `json:"nick_name" form:"nick_name"`
	Head_portrait string `json:"head_portrait" form:"head_portrait"`
	Gender        int    `json:"gender" form:"gender"`
	Residence     string `json:"residence" form:"residence"`
	Phone_number  string `json:"phone_number" form:"phone_number"`
	Name          string `json:"name" form:"name"`
	Birth_date    string `json:"birth_date" form:"birth_date"`
}

type Child struct {
	Child_id   int64     `json:"child_id" form:"child_id"`
	Name       string    `json:"name" form:"name"`
	Gender     int       `json:"gender" form:"gender"`
	Birth_date time.Time `json:"birth_date" form:"birth_date"`
}

type Uc_relation struct {
	Uc_relation_id int64 `json:"uc_relation_id" form:"uc_relation_id"`
	User_id        int64 `json:"user_id" form:"user_id"`
	Child_id       int64 `json:"child_id" form:"child_id"`
	Relation       int   `json:"relation" form:"relation"`
}

type Coupon struct {
	Coupon_id int64
	Secen_id  int
	Flag      int
	Ava_count int
	Ava_span  string
	Discount  string
}

type UserCoupon struct {
	User_coupon_id int64     `json:"user_coupon_id" form:"user_coupon_id"`
	User_id        int64     `json:"user_id" form:"user_id"`
	Code           string    `json:"code" form:"code"`
	Ava_count      int       `json:"ava_count" form:"ava_count"`
	Discount       string    `json:"discount" form:"discount"`
	Expiry_date    time.Time `json:"expiry_date" form:"expiry_date"`
}

var g_mapCoupons map[int]Coupon

func init() {
	g_mapCoupons = make(map[int]Coupon)

	rowCoupons, err := db.SqlDB.Query("SELECT coupon_id,secen_id,flag,ava_count,ava_span,discount FROM coupon")
	if err != nil {
		return
	}
	defer rowCoupons.Close()

	for rowCoupons.Next() {
		var coupon Coupon

		err = rowCoupons.Scan(&coupon.Coupon_id, &coupon.Secen_id, &coupon.Flag, &coupon.Ava_count, &coupon.Ava_span, &coupon.Discount)
		if err != nil {
			return
		}

		g_mapCoupons[coupon.Secen_id] = coupon
	}
}

// GetUserByOpenId 通过微信身份标识获取客户信息
func GetUserByOpenId(openid string) (user User, err error) {
	user.Openid = openid

	err = db.SqlDB.QueryRow("SELECT user_id,unionid,nick_name,head_portrait,gender,residence,phone_number,name,birth_date FROM user WHERE openid=?", openid).Scan(&user.User_id, &user.Unionid, &user.Nick_name, &user.Head_portrait, &user.Gender, &user.Residence, &user.Phone_number, &user.Name, &user.Birth_date)
	return
}

// GetUserByUserId 通过ID获取客户信息
func GetUserByUserId(user_id int64) (user User, err error) {
	user.User_id = user_id

	err = db.SqlDB.QueryRow("SELECT openid,unionid,nick_name,head_portrait,gender,residence,phone_number,name,birth_date FROM user WHERE user_id=?", user_id).Scan(&user.Openid, &user.Unionid, &user.Nick_name, &user.Head_portrait, &user.Gender, &user.Residence, &user.Phone_number, &user.Name, &user.Birth_date)
	return
}

// RefreshUser 刷新用户信息
func RefreshUser(openid string, unionid string, sceneid int, nick_name string, head_portrait string, gender int, residence string) (user User, err error) {
	user, err = GetUserByOpenId(openid)
	if err == sql.ErrNoRows {
		user.Openid = openid
		user.Unionid = unionid
		user.Sceneid = sceneid
		user.Nick_name = nick_name
		user.Head_portrait = head_portrait
		user.Gender = gender
		user.Residence = residence

		var rs sql.Result
		rs, err = db.SqlDB.Exec("INSERT INTO user(openid, unionid, sceneid, nick_name, head_portrait, gender, residence, birth_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", openid, unionid, sceneid, nick_name, head_portrait, gender, residence, time.Now())
		if err == nil {
			user.User_id, err = rs.LastInsertId()
		}
	} else if err == nil {
		user.Openid = openid
		user.Unionid = unionid
		user.Sceneid = sceneid
		user.Nick_name = nick_name
		user.Head_portrait = head_portrait
		user.Gender = gender
		user.Residence = residence

		_, err = db.SqlDB.Exec("UPDATE user SET sceneid=?, nick_name=?,head_portrait=?,gender=?,residence=? WHERE openid=?", sceneid, nick_name, head_portrait, gender, residence, openid)
	}

	return
}

// UpdateUser 更新用户信息
func UpdateUser(user_id int64, nick_name string, gender int, name string, birth_date string) (err error) {
	_, err = db.SqlDB.Exec("UPDATE user SET nick_name=?,gender=?,name=?,birth_date=? WHERE user_id=?", nick_name, gender, name, birth_date, user_id)

	return
}

// GetChildById 通过ID获取儿童信息
func GetChildById(child_id int64) (child Child, err error) {
	child.Child_id = child_id

	err = db.SqlDB.QueryRow("SELECT name,gender,birth_date FROM child WHERE child_id=?", child_id).Scan(&child.Name, &child.Gender, &child.Birth_date)
	return
}

// GetChildByUserId 通过用户ID获取儿童信息
func GetChildByUserId(user_id int64) (childs []Child, err error) {
	rows, err := db.SqlDB.Query("SELECT child.child_id,child.name,child.gender,child.birth_date FROM child, uc_relation WHERE uc_relation.user_id=? AND uc_relation.child_id = child.child_id", user_id)
	if err == sql.ErrNoRows {
		return childs, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var child Child
		err = rows.Scan(&child.Child_id, &child.Name, &child.Gender, &child.Birth_date)
		if err != nil {
			return nil, err
		}
		childs = append(childs, child)
	}

	return childs, err
}

// AddChild 增加儿童信息
func AddChild(child_id int64, name string, gender int, birth_date string) (err error) {
	_, err = db.SqlDB.Exec("INSERT INTO child(child_id, name, gender, birth_date) VALUES (?, ?, ?, ?)", child_id, name, gender, birth_date)

	return
}

// UpdateChild 更新儿童信息
func UpdateChild(child_id int64, name string, gender int, birth_date string) (err error) {
	_, err = db.SqlDB.Exec("UPDATE child SET name=?,gender=?,birth_date=? WHERE child_id=?", name, gender, birth_date, child_id)

	return
}

// GetUcRelation 获取用户儿童关联关系
func GetUcRelation(user_id int64, child_id int64) (ucRelation Uc_relation, err error) {
	ucRelation.User_id = user_id
	ucRelation.Child_id = child_id

	err = db.SqlDB.QueryRow("SELECT uc_relation_id,relation FROM uc_relation WHERE user_id=? AND child_id=?", user_id, child_id).Scan(&ucRelation.Uc_relation_id, &ucRelation.Relation)
	return
}

// AddUcRelation 增加用户儿童关联关系
func AddUcRelation(user_id int64, child_id int64, relation int) (err error) {
	_, err = db.SqlDB.Exec("INSERT INTO uc_relation(user_id, child_id, relation) VALUES (?, ?, ?)", user_id, child_id, relation)

	return
}

// UpdateUcRelation 更新用户儿童关联关系
func UpdateUcRelation(user_id int64, child_id int64, relation int) (err error) {
	_, err = db.SqlDB.Exec("UPDATE uc_relation SET relation=? WHERE user_id=? AND child_id=?", relation, user_id, child_id)

	return
}

// QryCoupon 查询优惠码信息
func QryCoupon(Secen_id int) (coupon *Coupon, err error) {
	value, ok := g_mapCoupons[Secen_id]
	if ok == true {
		coupon = &value
		err = nil
		return
	}

	return coupon, errors.New("无效的参数")
}

// QryUserCoupon 查询用户优惠码信息
func QryUserCoupon(user_id int64, code string) (userCoupon UserCoupon, err error) {
	err = db.SqlDB.QueryRow("SELECT user_coupon_id,ava_count,discount,expiry_date from user_coupon WHERE user_id=? AND code=?", user_id, code).Scan(&userCoupon.User_coupon_id, &userCoupon.Ava_count, &userCoupon.Discount, &userCoupon.Expiry_date)
	if err != nil {
		return userCoupon, err
	}
	userCoupon.User_id = user_id
	userCoupon.Code = code

	return userCoupon, err
}

// UseUserCoupon 使用单个用户优惠码
func UseUserCoupon(user_id int64, code string) (err error) {
	_, err = db.SqlDB.Exec("UPDATE user_coupon SET ava_count=ava_count-1 WHERE user_id=? AND code=?", user_id, code)
	return err
}

// AddUserCoupon 添加用户优惠信息
func AddUserCoupon(user_id int64, secen_id int, code string, ava_count int, discount string, expiry_date time.Time) (userCoupon UserCoupon, err error) {
	var user_coupon_id int64
	err = db.SqlDB.QueryRow("SELECT user_coupon_id from user_coupon WHERE user_id=? AND secen_id=?", user_id, secen_id).Scan(&user_coupon_id)
	if err == nil {
		return userCoupon, errors.New("用户已存在改场景下的优惠码")
	}

	if err != sql.ErrNoRows {
		return userCoupon, err
	}

	rs, err := db.SqlDB.Exec("INSERT INTO user_coupon(user_id, secen_id, code, ava_count, discount, expiry_date) VALUES (?, ?, ?, ?, ?, ?)", user_id, secen_id, code, ava_count, discount, expiry_date)
	if err == nil {
		userCoupon.User_coupon_id, err = rs.LastInsertId()
		userCoupon.User_id = user_id
		userCoupon.Code = code
		userCoupon.Ava_count = ava_count
		userCoupon.Discount = discount
		userCoupon.Expiry_date = expiry_date
	}

	return userCoupon, err
}

func GenCouponCode() (code string) {
	str := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 8; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
