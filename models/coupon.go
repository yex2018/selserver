package models

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"

	db "github.com/yex2018/selserver/database"
)

type Coupon struct {
	Coupon_id int    `json:"user_coupon_id" form:"user_coupon_id"`
	Flag      int    `json:"flag" form:"flag"`
	Secen_id  int    `json:"code" form:"code"`
	Ava_count int    `json:"ava_count" form:"ava_count"`
	Ava_span  string `json:"ava_span" form:"ava_span"`
	Discount  string `json:"discount" form:"discount"`
}

type UserCoupon struct {
	User_coupon_id int       `json:"user_coupon_id" form:"user_coupon_id"`
	Open_id        string    `json:"open_id" form:"open_id"`
	Code           string    `json:"code" form:"code"`
	Ava_count      int       `json:"ava_count" form:"ava_count"`
	Discount       string    `json:"discount" form:"discount"`
	Expiry_date    time.Time `json:"expiry_date" form:"expiry_date"`
}

// QryCoupon 查询优惠码信息
func QryCoupon(Secen_id int) (coupon Coupon, err error) {
	err = db.SqlDB.QueryRow("SELECT coupon_id,flag,ava_count,ava_span,discount from coupon WHERE secen_id=?", Secen_id).Scan(&coupon.Coupon_id, &coupon.Flag, &coupon.Ava_count, &coupon.Ava_span, &coupon.Discount)
	if err != nil {
		return coupon, err
	}
	coupon.Secen_id = Secen_id

	return coupon, err
}

// QryUserCoupon 查询用户优惠码信息
func QryUserCoupon(User_id int, Code string) (userCoupon UserCoupon, err error) {
	var open_id string
	err = db.SqlDB.QueryRow("SELECT openid from user WHERE user_id=?", User_id).Scan(&open_id)
	if err != nil {
		return userCoupon, err
	}

	err = db.SqlDB.QueryRow("SELECT user_coupon_id,ava_count,discount,expiry_date from user_coupon WHERE openid=? AND code=?", open_id, Code).Scan(&userCoupon.User_coupon_id, &userCoupon.Ava_count, &userCoupon.Discount, &userCoupon.Expiry_date)
	if err != nil {
		return userCoupon, err
	}
	userCoupon.Open_id = open_id
	userCoupon.Code = Code

	return userCoupon, err
}

// UseUserCoupon 使用单个用户优惠码
func UseUserCoupon(User_id int, Code string) (err error) {
	var open_id string
	err = db.SqlDB.QueryRow("SELECT openid from user WHERE user_id=?", User_id).Scan(&open_id)
	if err != nil {
		return err
	}

	_, err = db.SqlDB.Exec("UPDATE user_coupon SET ava_count=ava_count-1 WHERE openid=? AND code=?", open_id, Code)
	return err
}

// AddUserCoupon 添加用户优惠信息
func AddUserCoupon(Open_id string, Secen_id int, Code string, Ava_count int, Discount string, Expiry_date time.Time) (err error) {
	var user_coupon_id int
	err = db.SqlDB.QueryRow("SELECT user_coupon_id from user_coupon WHERE openid=? AND secen_id=?", Open_id, Secen_id).Scan(&user_coupon_id)
	if err == nil {
		return errors.New("用户已存在改场景下的优惠码")
	}

	if err != sql.ErrNoRows {
		return err
	}

	_, err = db.SqlDB.Exec("INSERT INTO user_coupon(openid, secen_id, code, ava_count, discount, expiry_date) VALUES (?, ?, ?, ?, ?, ?)", Open_id, Secen_id, Code, Ava_count, Discount, Expiry_date)
	return err
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
