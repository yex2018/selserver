package models

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"

	db "github.com/yex2018/selserver/database"
)

type Coupon struct {
	Coupon_id int
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
func AddUserCoupon(user_id int64, secen_id int, code string, ava_count int, discount string, expiry_date time.Time) (userCoupon UserCoupon, err error) {
	var user_coupon_id int
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
