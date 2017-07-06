package main

import "testing"

var robot *DingRobot

const testRobotToken = "77b9e8567a31c7726d10c4e438277de06de3d80a143136fd82e2fec58701bb4d"

func TestMarkdownMessage(t *testing.T) {
	robot = NewRobot(testRobotToken)
	msg := NewMessageBuilder(TypeMarkdown).Markdown("标题标题", "# **大标题** \n [一条链接](http://qq.com)").Build()

	if err := robot.SendMessage(msg); err != nil {
		t.Error(err)
	}
}

func TestActionCardMessage(t *testing.T) {
	robot = NewRobot(testRobotToken)
	actionCardBuilder := NewActionCardBuilder("这是一个活动", "活动啊活动啊", OrientationHorizon, HideAvatar)
	actionCardBuilder.SingleButton("只有一个按钮", "http://qq.com")
	msg := NewMessageBuilder(TypeActionCard).ActionCard(actionCardBuilder.Build()).Build()
	if err := robot.SendMessage(msg); err != nil {
		t.Error(err)
	}
}

func TestFeedCardBuilder(t *testing.T) {
	robot = NewRobot(testRobotToken)
	feedCardBuilder := NewFeedCardBuilder()
	feedCardBuilder.Link("第一个链接", "http://qq.com", "http://dl.bizhi.sogou.com/images/2012/03/14/124196.jpg")
	feedCardBuilder.Link("第二个链接", "http://qq.com", "http://dl.bizhi.sogou.com/images/2012/03/14/124196.jpg")
	feedCardBuilder.Link("第三个链接", "http://qq.com", "http://dl.bizhi.sogou.com/images/2012/03/14/124196.jpg")
	msg := NewMessageBuilder(TypeFeedCard).FeedCard(feedCardBuilder.Build()).Build()
	if err := robot.SendMessage(msg); err != nil {
		t.Error(err)
	}
}