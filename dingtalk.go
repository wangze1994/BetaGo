package main

import (
	"strings"
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
	"errors"
	"math/rand"
	"log"
	"time"
	"github.com/robfig/cron"
)

const BaseSendURL = "https://oapi.dingtalk.com/robot/send?access_token={ACCESS_TOKEN}"
const JSONType = "application/json"
const RobotToken = "9a004a28c2bba941a71b0062685cac4068a9c73bf8af0f125e385150ef9cfda2"
const WeatherApi = "https://free-api.heweather.com/v5/now?city=beijing&key=eafba1ed10ea4d0cb67e22d81127c703"
const JueJinApi = "https://timeline-merger-ms.juejin.im/v1/get_entry_by_timeline?before=&limit=200&src=ios&tag=5597a05ae4b08a686ce56f6f"
const NewsApi = "http://www.toutiao.com/api/pc/feed/?category=internet&utm_source=toutiao"
const TouTiaoUrl = "http://www.toutiao.com"

type MessageType string
type Orientation string
type AvatarState string

const (
	TypeText            MessageType = "text"
	TypeLink            MessageType = "link"
	TypeMarkdown        MessageType = "markdown"
	TypeActionCard      MessageType = "actionCard"
	TypeFeedCard        MessageType = "feedCard"
	OrientationVertical Orientation = "0"
	OrientationHorizon  Orientation = "1"

	ShowAvatar AvatarState = "0"
	HideAvatar AvatarState = "1"

	WeatherCron string = "0 30 8 * * 1-5"
	JueJinCron  string = "0 30 12,18 * * *"
	NewsCron    string = "0 00 20 * * *"
)

// 定义Message结构体
type DingMessage struct {
	Type       MessageType `json:"msgtype"`
	Text       TextElement `json:"text"`
	Link       LinkElement `json:"link"`
	Markdown   MarkdownElement `json:"markdown"`
	ActionCard ActionCardElement `json:"actionCard"`
	FeedCard   FeedCardElement `json:"feedCard"`
	At         AtElement `json:"at"`
}

// text类型结构体
type TextElement struct {
	Content string `json:"content"`
}

// link类型结构体
type LinkElement struct {
	Title      string `json:"title"`
	Text       string `json:"text"`
	MessageURL string `json:"messageUrl"`
	PictureURL string `json:"picUrl"`
}

// markdown类型结构体
type MarkdownElement struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// ActionCard元素结构体
type ActionCardElement struct {
	Title             string `json:"title"`
	Text              string `json:"text"`
	SingleTitle       string `json:"singleTitle"`
	SingleURL         string `json:"singleURL"`
	ButtonOrientation Orientation `json:"btnOrientation"`
	Avatar            AvatarState `json:"hideAvatar"`
	Buttons           []ButtonElement `json:"btns"`
}

// button元素结构体
type ButtonElement struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

type FeedCardElement struct {
	Links []FeedLinkElement `json:"links"`
}

type FeedLinkElement struct {
	Title      string `json:"title"`
	Text       string `json:"text"`
	MessageURL string `json:"messageURL"`
	PictureURL string `json:"picURL"`
}

type AtElement struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool `json:"isAtAll"`
}

type actionCardBuilder struct {
	actionCard ActionCardElement
}

func NewActionCardBuilder(title string, text string, buttonOrientation Orientation, avatarState AvatarState) *actionCardBuilder {
	return &actionCardBuilder{
		ActionCardElement{
			Title:             title,
			Text:              text,
			ButtonOrientation: buttonOrientation,
			Avatar:            avatarState,
			Buttons:           make([]ButtonElement, 0),
		},
	}
}

func (builder *actionCardBuilder) SingleButton(title string, URL string) *actionCardBuilder {
	builder.actionCard.SingleTitle = title
	builder.actionCard.SingleURL = URL
	return builder
}

func (builder *actionCardBuilder) Button(title string, URL string) *actionCardBuilder {
	builder.actionCard.Buttons = append(builder.actionCard.Buttons, ButtonElement{
		Title:     title,
		ActionURL: URL,
	})
	return builder
}

func (builder *actionCardBuilder) Build() ActionCardElement {
	return builder.actionCard
}

type feedCardBuilder struct {
	feedCard FeedCardElement
}

func NewFeedCardBuilder() *feedCardBuilder {
	return &feedCardBuilder{
		FeedCardElement{
			Links: make([]FeedLinkElement, 0),
		},
	}
}

func (builder *feedCardBuilder) Link(title string, messageURL string, pictureURL string) *feedCardBuilder {
	builder.feedCard.Links = append(builder.feedCard.Links, FeedLinkElement{
		Title:      title,
		MessageURL: messageURL,
		PictureURL: pictureURL,
	})
	return builder
}

func (builder *feedCardBuilder) Build() FeedCardElement {
	return builder.feedCard
}

type ret struct {
	ErrorCode    int `json:"errcode"`
	ErrorMessage string `json:"errmsg"`
}

type DingRobot struct {
	AccessToken string
	SendURL     string
}

func NewRobot(accessToken string) *DingRobot {
	return &DingRobot{
		AccessToken: accessToken,
		SendURL:     strings.Replace(BaseSendURL, "{ACCESS_TOKEN}", accessToken, 1),
	}
}

func (dr DingRobot) SendMessage(msg DingMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	resp, err := http.Post(dr.SendURL, JSONType, bytes.NewReader(body))
	if err != nil {
		return err
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	ret := new(ret)
	err = json.Unmarshal(body, ret)
	if err != nil {
		return err
	}
	if ret.ErrorCode != 0 {
		return errors.New(ret.ErrorMessage)
	}
	return nil
}

type MessageBuilder struct {
	message DingMessage
}

func NewMessageBuilder(msgType MessageType) *MessageBuilder {
	return &MessageBuilder{
		DingMessage{
			Type: msgType,
		},
	}
}

func (builder *MessageBuilder) Text(text string) *MessageBuilder {
	builder.message.Text = TextElement{Content: text}
	return builder
}

func (builder *MessageBuilder) Link(title string, text string, messageURL string, pictureURL string) *MessageBuilder {
	builder.message.Link = LinkElement{
		Title:      title,
		Text:       text,
		MessageURL: messageURL,
		PictureURL: pictureURL,
	}
	return builder
}

func (builder *MessageBuilder) Markdown(title string, text string) *MessageBuilder {
	builder.message.Markdown = MarkdownElement{
		Title: title,
		Text:  text,
	}
	return builder
}

func (builder *MessageBuilder) ActionCard(element ActionCardElement) *MessageBuilder {
	builder.message.ActionCard = element
	return builder
}

func (builder *MessageBuilder) FeedCard(element FeedCardElement) *MessageBuilder {
	builder.message.FeedCard = element
	return builder
}

func (builder *MessageBuilder) At(mobiles []string, isAtAll bool) *MessageBuilder {
	builder.message.At = AtElement{
		AtMobiles: mobiles,
		IsAtAll:   isAtAll,
	}
	return builder
}

func (builder *MessageBuilder) Build() DingMessage {
	return builder.message
}

// 天气结构体
type weatherRet struct {
	Result []WeatherResultElement `json:"HeWeather5"`
}

type WeatherResultElement struct {
	Now NowElement `json:"now"`
}

type NowElement struct {
	Cond CondElement `json:"cond"`
	Fl   string `json:"fl"`
	Hum  string `json:"hum"`
	Tem  string `json:"tmp"`
	Wind WindElement `json:"wind"`
}

type CondElement struct {
	Txt string `json:"txt"`
}

type WindElement struct {
	Deg string `json:"deg"`
	Dir string `json:"dir"`
	Sc  string `json:"sc"`
	Spd string `json:"spd"`
}

// 掘金结构体
type juejinRet struct {
	S int `json:"s"`
	M string `json:"m"`
	D DElement `json:"d"`
}

type DElement struct {
	Entrylist []EntrylistElement `json:"entrylist"`
}

type EntrylistElement struct {
	Title       string `json:"title"`
	Screenshot  string `json:"screenshot"`
	OriginalUrl string `json:"originalUrl"`
}

// 新闻结构体
type newsRet struct {
	Data []DataElement `json:"data"`
}

type DataElement struct {
	ImageList []ImageListElement `json:"image_list"`
	SourceUrl string `json:"source_url"`
	Title     string `json:"title"`
	Abstract  string `json:"abstract"`
}

type ImageListElement struct {
	Url string `json:"url"`
}

// 钉天气
func (dr DingRobot) DingWeather() error {
	var err error
	resp, err := http.Get(WeatherApi)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	weatherRet := new(weatherRet)
	err = json.Unmarshal(body, weatherRet)
	text := "#### 互联网金融中心天气\n" +
		"> 天气状况" + weatherRet.Result[0].Now.Cond.Txt + "，温度" +
		weatherRet.Result[0].Now.Tem + "度，体感温度" +
		weatherRet.Result[0].Now.Fl + "度，相对湿度" +
		weatherRet.Result[0].Now.Hum + "%，" +
		weatherRet.Result[0].Now.Wind.Dir + weatherRet.Result[0].Now.Wind.Sc + "级，" + "\n\n" +
		"> ###### 08点20分发布 数据来自[和风天气](https://www.heweather.com/) \n"
	msg := NewMessageBuilder(TypeMarkdown).Markdown("早上好~", text).Build()
	if err := dr.SendMessage(msg); err != nil {
		return err
	}
	return err
}

// 钉掘金
func (dr DingRobot) DingJueJin() error {
	var err error
	resp, err := http.Get(JueJinApi)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	juejinRet := new(juejinRet)
	err = json.Unmarshal(body, juejinRet)
	feedCardBuilder := NewFeedCardBuilder()
	num := RandInt64(1, 196)
	log.Println("num:", num)
	for _, s := range juejinRet.D.Entrylist[num: num+3] {
		feedCardBuilder.Link(s.Title, s.OriginalUrl, s.Screenshot)
	}
	msg := NewMessageBuilder(TypeFeedCard).FeedCard(feedCardBuilder.Build()).Build()
	if err := dr.SendMessage(msg); err != nil {
		return err
	}
	return err
}

// 钉新闻
func (dr DingRobot) DingNews() error {
	var err error
	resp, err := http.Get(NewsApi)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	newsRet := new(newsRet)
	err = json.Unmarshal(body, newsRet)
	num := RandInt64(1, 5)
	tilte := newsRet.Data[0].Title
	text := "![screenshot](" + newsRet.Data[num].ImageList[0].Url + ") \n" +
		"### " + newsRet.Data[num].Title + "\n " +
		newsRet.Data[num].Abstract
	actionURL := TouTiaoUrl + newsRet.Data[num].SourceUrl
	actionCardBuilder := NewActionCardBuilder(tilte, text, OrientationHorizon, HideAvatar)
	actionCardBuilder.SingleButton("查看详情", actionURL)
	msg := NewMessageBuilder(TypeActionCard).ActionCard(actionCardBuilder.Build()).Build()
	if err := dr.SendMessage(msg); err != nil {
		return err
	}
	return err
}

func (dr DingRobot) DingBasketBall() {
	mobiles := []string{"13552798619"}
	msg := NewMessageBuilder(TypeText).At(mobiles, false).Text("今天你俩谁硬？").Build()
	if err := dr.SendMessage(msg); err != nil {

	}
}

func (dr DingRobot) ExecError(msgType string) {
	msg := NewMessageBuilder(TypeText).Text("抱歉~狗狗今儿没拿到最新" + msgType + "数据。").Build()
	if err := dr.SendMessage(msg); err != nil {

	}
}

// 生成随机数
func RandInt64(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min) + min
}

func main() {
	robot := NewRobot(RobotToken)
	c := cron.New()
	c.AddFunc(WeatherCron, func() {
		log.Println("start weather")
		if err := robot.DingWeather(); err != nil {
			robot.ExecError("天气")
		}
	})
	c.AddFunc(JueJinCron, func() {
		log.Println("start juejin")
		if err := robot.DingJueJin(); err != nil {
			robot.ExecError("掘金技术文章")
		}
	})
	c.AddFunc(NewsCron, func() {
		log.Println("start news")
		if err := robot.DingNews(); err != nil {
			robot.ExecError("头条科技新闻")
		}
	})
	//c.AddFunc(NewsCron, func() {
	//	log.Println("start basketball")
	//	robot.DingBasketBall()
	//})
	c.Start()
	select {} //阻塞主线程不退出
}
