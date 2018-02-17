package apis

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yex2018/selserver/models"

	"github.com/yex2018/selserver/tool"

	"github.com/chanxuehong/rand"
	"github.com/chanxuehong/session"
	"github.com/chanxuehong/sid"

	"github.com/chanxuehong/wechat.v2/mp/core"
	"github.com/chanxuehong/wechat.v2/mp/jssdk"
	"github.com/chanxuehong/wechat.v2/mp/media"
	"github.com/chanxuehong/wechat.v2/mp/menu"
	"github.com/chanxuehong/wechat.v2/mp/message/callback/request"
	"github.com/chanxuehong/wechat.v2/mp/message/callback/response"
	"github.com/chanxuehong/wechat.v2/mp/message/custom"
	"github.com/chanxuehong/wechat.v2/mp/message/template"
	mpoauth2 "github.com/chanxuehong/wechat.v2/mp/oauth2"
	"github.com/chanxuehong/wechat.v2/mp/user"

	"github.com/chanxuehong/wechat.v2/oauth2"
	"github.com/gin-gonic/gin"
	"github.com/yex2018/selserver/conf"
)

var (
	wxAppId           = conf.Config.WXAppID
	wxAppSecret       = conf.Config.WXAppSecret
	wxOriId           = conf.Config.WXOriID
	wxToken           = conf.Config.WXToken
	wxEncodedAESKey   = ""
	oauth2RedirectURI = conf.Config.Oauth2RedirectURI
	oauth2Scope       = "snsapi_userinfo"
	// 下面两个变量不一定非要作为全局变量, 根据自己的场景来选择.
	msgHandler core.Handler
	msgServer  *core.Server

	sessionStorage                           = session.New(20*60, 60*60)
	oauth2Endpoint    oauth2.Endpoint        = mpoauth2.NewEndpoint(wxAppId, wxAppSecret)
	accessTokenServer core.AccessTokenServer = core.NewDefaultAccessTokenServer(wxAppId, wxAppSecret, nil)
	wechatClient      *core.Client           = core.NewClient(accessTokenServer, nil)
	ticketserver                             = jssdk.NewDefaultTicketServer(wechatClient)
)

func init() {
	mux := core.NewServeMux()
	mux.DefaultMsgHandleFunc(defaultMsgHandler)
	mux.DefaultEventHandleFunc(defaultEventHandler)
	mux.MsgHandleFunc(request.MsgTypeText, textMsgHandler)
	mux.EventHandleFunc(menu.EventTypeClick, menuClickEventHandler)
	mux.EventHandleFunc(request.EventTypeSubscribe, subscribeEventHandler)
	mux.EventHandleFunc(request.EventTypeScan, scanEventHandler)

	msgHandler = mux
	msgServer = core.NewServer(wxOriId, wxAppId, wxToken, wxEncodedAESKey, msgHandler, nil)
}

func textMsgHandler(ctx *core.Context) {
	log.Printf("收到文本消息:\n%s\n", ctx.MsgPlaintext)

	msg := request.GetText(ctx.MixedMsg)
	resp := response.NewText(msg.FromUserName, msg.ToUserName, msg.CreateTime, msg.Content)
	ctx.RawResponse(resp) // 明文回复
}

func defaultMsgHandler(ctx *core.Context) {
	log.Printf("收到消息:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()
}

func menuClickEventHandler(ctx *core.Context) {
	log.Printf("收到菜单 click 事件:\n%s\n", ctx.MsgPlaintext)

	event := menu.GetClickEvent(ctx.MixedMsg)
	resp := response.NewText(event.FromUserName, event.ToUserName, event.CreateTime, "收到 click 类型的事件")
	ctx.RawResponse(resp) // 明文回复
}

func defaultEventHandler(ctx *core.Context) {
	log.Printf("收到事件:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()
}

func subscribeEventHandler(ctx *core.Context) {
	log.Printf("收到subscribe事件:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()

	event := request.GetSubscribeEvent(ctx.MixedMsg)
	if event.FromUserName == "" {
		return
	}

	srv := core.NewDefaultAccessTokenServer(wxAppId, wxAppSecret, nil)
	clt := core.NewClient(srv, nil)
	userInfo, err := user.Get(clt, event.FromUserName, "zh_CN")
	if err != nil {
		return
	}

	sendCustomMsg(clt, event.FromUserName, getSubscribeMsg(userInfo.Nickname))

	sceneid := 0
	scene, err := event.Scene()
	if scene != "" {
		sceneid, err = strconv.Atoi(scene)
		if err != nil {
			sceneid = 0
		}
	}

	user, err := models.RefreshUser(userInfo.OpenId, userInfo.UnionId, sceneid, userInfo.Nickname, userInfo.HeadImageURL, userInfo.Sex, userInfo.Country+","+userInfo.Province+","+userInfo.City)
	if err != nil {
		log.Printf("RefreshUser:%s\n", err.Error())
		return
	}

	coupon, err := models.QryCoupon(sceneid)
	if err != nil {
		log.Printf("QryCoupon:%s\n", err.Error())
		return
	}

	span, err := time.ParseDuration(coupon.Ava_span)
	if err != nil {
		log.Printf("ParseDuration:%s\n", err.Error())
		return
	}

	userCoupon, err := models.AddUserCoupon(user.User_id, sceneid, models.GenCouponCode(), coupon.Ava_count, coupon.Discount, time.Now().Add(span))
	if err == nil {
		sendCustomMsg(clt, event.FromUserName, GetCouponMsg(userInfo.Nickname, &userCoupon))
	} else {
		log.Printf("AddUserCoupon:%s\n", err.Error())
	}
}

func scanEventHandler(ctx *core.Context) {
	log.Printf("收到scan事件:\n%s\n", ctx.MsgPlaintext)
	ctx.NoneResponse()

	event := request.GetScanEvent(ctx.MixedMsg)
	if event.EventKey == "" || event.FromUserName == "" {
		return
	}

	srv := core.NewDefaultAccessTokenServer(wxAppId, wxAppSecret, nil)
	clt := core.NewClient(srv, nil)
	userInfo, err := user.Get(clt, event.FromUserName, "zh_CN")
	if err != nil {
		return
	}

	sceneid := 0
	sceneid, err = strconv.Atoi(event.EventKey)
	if err != nil {
		sceneid = 0
	}

	user, err := models.RefreshUser(userInfo.OpenId, userInfo.UnionId, sceneid, userInfo.Nickname, userInfo.HeadImageURL, userInfo.Sex, userInfo.Country+","+userInfo.Province+","+userInfo.City)
	if err != nil {
		log.Printf("RefreshUser:%s\n", err.Error())
		return
	}

	coupon, err := models.QryCoupon(sceneid)
	if err != nil {
		log.Printf("QryCoupon:%s\n", err.Error())
		return
	}

	if coupon.Flag != 0 {
		return
	}

	span, err := time.ParseDuration(coupon.Ava_span)
	if err != nil {
		log.Printf("ParseDuration:%s\n", err.Error())
		return
	}

	userCoupon, err := models.AddUserCoupon(user.User_id, sceneid, models.GenCouponCode(), coupon.Ava_count, coupon.Discount, time.Now().Add(span))
	if err == nil {
		sendCustomMsg(clt, event.FromUserName, GetCouponMsg(userInfo.Nickname, &userCoupon))
	} else {
		log.Printf("AddUserCoupon:%s\n", err.Error())
	}
}

func WeixinHandler(c *gin.Context) {
	msgServer.ServeHTTP(c.Writer, c.Request, nil)
}

// 建立必要的 session, 然后跳转到授权页面
func Page1Handler(c *gin.Context) {
	sid := sid.New()
	state := string(rand.NewHex())

	if err := sessionStorage.Add(sid, state); err != nil {
		io.WriteString(c.Writer, err.Error())
		log.Println(err)
		return
	}

	cookie := http.Cookie{
		Name:     "sid",
		Value:    sid,
		HttpOnly: true,
		MaxAge:   int(time.Minute / time.Second),
	}
	http.SetCookie(c.Writer, &cookie)

	AuthCodeURL := mpoauth2.AuthCodeURL(wxAppId, oauth2RedirectURI+"?menuType="+c.Query("menuType"), oauth2Scope, state)

	http.Redirect(c.Writer, c.Request, AuthCodeURL, http.StatusFound)
}

// Page2Handler 授权后回调页面
func Page2Handler(c *gin.Context) {
	log.Println(c.Request.RequestURI)

	cookie, err := c.Cookie("sid")
	if err != nil {
		io.WriteString(c.Writer, err.Error())
		log.Println(err)
		return
	}

	session, err := sessionStorage.Get(cookie)
	if err != nil {
		io.WriteString(c.Writer, err.Error())
		log.Println(err)
		return
	}

	savedState := session.(string) // 一般是要序列化的, 这里保存在内存所以可以这么做

	code := c.Query("code")
	if code == "" {
		log.Println("用户禁止授权")
		return
	}

	queryState := c.Query("state")
	if queryState == "" {
		log.Println("state 参数为空")
		return
	}
	if savedState != queryState {
		str := fmt.Sprintf("state 不匹配, session 中的为 %q, url 传递过来的是 %q", savedState, queryState)
		io.WriteString(c.Writer, str)
		log.Println(str)
		return
	}

	oauth2Client := oauth2.Client{
		Endpoint: oauth2Endpoint,
	}
	token, err := oauth2Client.ExchangeToken(code)
	if err != nil {
		io.WriteString(c.Writer, err.Error())
		tool.Error(err)
		return
	}

	userInfo, err := mpoauth2.GetUserInfo(token.AccessToken, token.OpenId, "", nil)
	if err != nil {
		io.WriteString(c.Writer, err.Error())
		tool.Error(err)
		return
	}

	_, err = models.RefreshUser(userInfo.OpenId, userInfo.UnionId, 0, userInfo.Nickname, userInfo.HeadImageURL, userInfo.Sex, userInfo.Country+","+userInfo.Province+","+userInfo.City)
	if err != nil {
		tool.Error(err)
		return
	}

	usercookie1 := http.Cookie{
		Name:  "openid",
		Value: userInfo.OpenId,
	}
	usercookie2 := http.Cookie{
		Name:  "nickname",
		Value: url.QueryEscape(userInfo.Nickname),
	}
	usercookie3 := http.Cookie{
		Name:  "headimgurl",
		Value: userInfo.HeadImageURL,
	}
	accesstoken, err := ticketserver.Ticket()
	if err != nil {
		io.WriteString(c.Writer, err.Error())
		tool.Error(err)
		return
	}
	noncestr := string(rand.NewHex())
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	sign := jssdk.WXConfigSign(accesstoken, noncestr, timestamp, conf.Config.Host+"/front/dist/?")
	ss := []string{conf.Config.WXAppID, timestamp, noncestr, sign}
	wxconfigCookie := http.Cookie{
		Name:  "wxconfig",
		Value: strings.Join(ss, "|"),
	}
	http.SetCookie(c.Writer, &usercookie1)
	http.SetCookie(c.Writer, &usercookie2)
	http.SetCookie(c.Writer, &usercookie3)
	http.SetCookie(c.Writer, &wxconfigCookie)
	AuthCodeURL := ""
	switch menuType := c.Query("menuType"); menuType {
	case "1":
		AuthCodeURL = "/front/dist/?#/appbase/assessment"
	case "2":
		AuthCodeURL = "/front/dist/?#/appbase/course"
	case "3":
		AuthCodeURL = "/front/dist/?#/appbase/mine"
	default:
		AuthCodeURL = "/front/appbase/mine"
	}
	http.Redirect(c.Writer, c.Request, AuthCodeURL, http.StatusFound)
	return
}

// DownloadMedia 通过mediaid下载媒体文件
func DownloadMedia(c *gin.Context) {
	mediaID := c.Query("mediaid")
	if mediaID == "" {
		c.Error(errors.New("参数为空"))
		return
	}
	fileName := getFileName(mediaID)
	if checkFileIsExist(fileName) {
		c.JSON(http.StatusOK, models.Result{Data: "/front/childimg/" + mediaID + ".jpg"})
		return
	}
	myfile, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		c.Error(err)
		return
	}
	_, err = media.DownloadToWriter(wechatClient, mediaID, myfile)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, models.Result{Data: "/front/childimg/" + mediaID + ".jpg"})
}

func getFileName(mediaID string) string {
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, "front", "childimg", mediaID+".jpg")
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// TemplateMessage 发送模板消息
func TemplateMessage(openid, url, evaluationName, evaluationTime, nick_name, childName string) (err error) {
	type TemplateMessage struct {
		ToUser     string          `json:"touser"`        // 必须, 接受者OpenID
		TemplateId string          `json:"template_id"`   // 必须, 模版ID
		URL        string          `json:"url,omitempty"` // 可选, 用户点击后跳转的URL, 该URL必须处于开发者在公众平台网站中设置的域中
		Data       json.RawMessage `json:"data"`          // 必须, 模板数据, JSON 格式的 []byte, 满足特定的模板需求
	}
	json := `{
		"first": {
			"value":"您好，` + nick_name + `，您有一份完整测评报告已生成。\r\n",
			"color":""
		},
		"keyword1":{
			"value":"` + evaluationName + `",
			"color":""
		},
		"keyword2": {
			"value":"` + childName + `",
			"color":""
		},
		"keyword3": {
			"value":"` + evaluationTime + `\r\n",
			"color":""
		},
		"remark":{
			"value":"点击查看完整测评报告。\r\n您可以转发本消息，与家人分享报告。",
			"color":"#89bd41"
		}}`

	tool.Info("json:" + json)
	var jsonBlob = []byte(json)

	msg := TemplateMessage{ToUser: openid, TemplateId: conf.Config.Template_id, URL: url, Data: jsonBlob}
	msgid, err := template.Send(wechatClient, msg)
	id := strconv.FormatInt(msgid, 10)
	if err != nil && id != "" {
		return err
	}
	return err
}

func getSubscribeMsg(nickName string) (msg string) {
	msg = "Hi，" + nickName + "。\r\n" +
		"这里是由一群北大教育学博士发起的薄荷叶教育。\r\n" +
		"我们关注儿童的内在成长，希望每个孩子都有一颗丰盈的内心，生动而独特。\r\n" +
		"我们希望用我们的专业精神，为大家提供一份对儿童真正有意义的教育资源。\r\n\r\n" +
		"想知道SEL是什么？\r\n" +
		"<a href=\"http://sel.bheonline.com/front/dist/?#/courseFree?course_id=1\">点击这里</a>\r\n\r\n" +
		"想了解孩子SEL的发育水平？\r\n" +
		"<a href=\"http://sel.bheonline.com/front/dist/?#/appbase/assessment\">点击这里</a>\r\n\r\n" +
		"有任何问题，请给我们留言。\r\n\r\n" +
		"薄荷叶教育，与孩子一起探索更好的自己！"

	return
}

func GetCouponMsg(nickName string, userCoupon *models.UserCoupon) (msg string) {
	msg = "Hi，" + nickName +
		"，恭喜您获得一个薄荷叶教育产品优惠码。\r\n\r\n优惠码：" + userCoupon.Code +
		"\r\n优惠幅度：免费\r\n有效期限：" + userCoupon.Expiry_date.Format("2006-01-02 15:04:05") +
		"\r\n有效次数：" + strconv.Itoa(userCoupon.Ava_count) +
		"次\r\n\r\n您可以选择一款薄荷叶教育产品，在付款时输入优惠码即可使用。"

	return
}

func sendCustomMsg(clt *core.Client, toUser string, msg string) {
	type TextContent struct {
		Content string `json:"content"`
	}

	type CustomMsg struct {
		Touser  string      `json:"touser"`
		Msgtype string      `json:"msgtype"`
		Text    TextContent `json:"text"`
	}

	textContent := TextContent{Content: msg}
	customMsg := CustomMsg{Touser: toUser, Msgtype: "text", Text: textContent}
	custom.Send(clt, customMsg)
}
