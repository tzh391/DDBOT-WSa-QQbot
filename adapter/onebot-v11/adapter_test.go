package onebot

import (
	"testing"
	"time"

	"github.com/cnxysoft/DDBOT-WSa/adapter"
	"github.com/stretchr/testify/assert"
)

func TestContainsURIWithTypedMessageSegments(t *testing.T) {
	params := adapter.BuildMessageParams([]adapter.MessageSegment{
		{
			Type: "text",
			Data: map[string]interface{}{"text": "news"},
		},
		{
			Type: "image",
			Data: map[string]interface{}{"file": "https://example.com/image.png"},
		},
	})

	assert.True(t, containsURI(params))
}

func TestCalcSendTimeoutForTypedMessageSegments(t *testing.T) {
	defaultTimeout := 10 * time.Second
	a := NewOneBotAdapter(&adapter.AdapterConfig{Timeout: defaultTimeout})

	remoteImage := adapter.BuildMessageParams([]adapter.MessageSegment{{
		Type: "image",
		Data: map[string]interface{}{"url": "https://example.com/image.png"},
	}})
	plainText := adapter.BuildMessageParams([]adapter.MessageSegment{{
		Type: "text",
		Data: map[string]interface{}{"text": "hello"},
	}})

	assert.Equal(t, uriMessageTimeout, a.calcSendTimeout("send_group_msg", remoteImage))
	assert.Less(t, uriMessageTimeout, time.Minute)
	assert.Equal(t, defaultTimeout, a.calcSendTimeout("send_group_msg", plainText))
}

func TestMessageEventsPreserveSenderProfile(t *testing.T) {
	a := NewOneBotAdapter(&adapter.AdapterConfig{})
	msg := map[string]interface{}{
		"user_id": float64(42), "group_id": float64(100),
		"sender": map[string]interface{}{
			"user_id": float64(999), "nickname": "事件昵称", "card": "事件名片", "role": "owner",
		},
	}
	group := a.parseGroupMessageEvent(msg)
	private := a.parsePrivateMessageEvent(msg)
	for _, sender := range []*adapter.SenderInfo{group.Sender, private.Sender} {
		if assert.NotNil(t, sender) {
			assert.Equal(t, int64(42), sender.UserID)
			assert.Equal(t, int64(42), sender.Uin)
			assert.Equal(t, "事件昵称", sender.Nickname)
			assert.Equal(t, "事件名片", sender.Card)
			assert.Equal(t, "owner", sender.Role)
		}
	}
	delete(msg, "sender")
	assert.Nil(t, a.parseGroupMessageEvent(msg).Sender)
	assert.Nil(t, a.parsePrivateMessageEvent(msg).Sender)
}
