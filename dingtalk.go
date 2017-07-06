package main

import (
	"strings"
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
	"errors"
	"strconv"
	"math/rand"
	"log"
	"time"
	"github.com/robfig/cron"
)

const BaseSendURL = "https://oapi.dingtalk.com/robot/send?access_token={ACCESS_TOKEN}"
const JSONType = "application/json"
const RobotToken = "77b9e8567a31c7726d10c4e438277de06de3d80a143136fd82e2fec58701bb4d"
const WeatherApi = "https://api.caiyunapp.com/v2/TAkhjf8d1nlSlspN/116.320295,39.985358/realtime.json"
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

	WeatherCron string = "0 37 23 * * 1-5"
	JueJinCron  string = "0 0 10 * * *"
	NewsCron    string = "0 0 14,20 * * *"
)

var skycon = map[string]string{
	"CLEAR_DAY":           "晴天",
	"CLEAR_NIGHT":         "晴夜",
	"PARTLY_CLOUDY_DAY":   "多云",
	"PARTLY_CLOUDY_NIGHT": "多云",
	"CLOUDY":              "阴",
	"RAIN":                "雨",
	"SNOW":                "雪",
	"WIND":                "风",
	"FOG":                 "雾",
	"HAZE":                "霾",
	"SLEET":               "冻雨",
}

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
	Result WeatherResultElement `json:"result"`
}

type WeatherResultElement struct {
	Temperature float64 `json:"temperature"`
	Skycon      string `json:"skycon"`
	Pm25        float64 `json:"pm25"`
	Humidity    float64 `json:"humidity"`
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
func (dr DingRobot) DingWeather() {
	resp, err := http.Get(WeatherApi)
	if err != nil {

	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

	}
	weatherRet := new(weatherRet)
	err = json.Unmarshal(body, weatherRet)
	text := "#### 互联网金融中心天气\n" +
		"> " + skycon[weatherRet.Result.Skycon] + "天气，温度" +
		strconv.FormatFloat(weatherRet.Result.Temperature, 'f', 1, 64) + "度，pm25值为" +
		strconv.FormatFloat(weatherRet.Result.Pm25, 'f', 1, 64) +
		"，相对湿度" + strconv.FormatFloat(weatherRet.Result.Humidity, 'f', 1, 64) + "\n\n" +
		"> ###### 08点20分发布 [天气](http://www.weather.com.cn/) \n"
	msg := NewMessageBuilder(TypeMarkdown).Markdown("早上好~", text).Build()
	if err := dr.SendMessage(msg); err != nil {

	}
}

// 钉掘金
func (dr DingRobot) DingJueJin() {
	resp, err := http.Get(JueJinApi)
	if err != nil {

	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

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

	}
}

// 钉新闻
func (dr DingRobot) DingNews() {
	resp, err := http.Get(NewsApi)
	if err != nil {

	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

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
		robot.DingWeather()
	})
	c.AddFunc(JueJinCron, func() {
		log.Println("start juejin")
		robot.DingJueJin()
	})
	c.AddFunc(NewsCron, func() {
		log.Println("start news")
		robot.DingNews()
	})
	c.Start()
	select {} //阻塞主线程不退出
}
