package adapter

import "testing"

type senderMessageDispatcher struct {
	*mockDispatcher
	group   *GroupMessage
	private *PrivateMessage
}

func (d *senderMessageDispatcher) DispatchGroupMessage(msg *GroupMessage) { d.group = msg }
func (d *senderMessageDispatcher) DispatchPrivateMessage(msg *PrivateMessage) {
	d.private = msg
}

func TestGroupMessageSenderWithIncompleteMemberCache(t *testing.T) {
	for _, tc := range []struct {
		name     string
		incoming *SenderInfo
		member   *GroupMemberInfo
		friend   *FriendInfo
		want     string
	}{
		{"event_without_cache", &SenderInfo{Nickname: "事件昵称", Role: "owner"}, nil, nil, "事件昵称"},
		{"event_over_cached_card", &SenderInfo{Nickname: "新昵称", Card: "新名片"}, &GroupMemberInfo{Uin: 42, Nickname: "旧昵称", CardName: "旧名片"}, nil, "新名片"},
		{"legacy_cache_fallback", nil, &GroupMemberInfo{Uin: 42, Nickname: "缓存昵称", CardName: "缓存名片"}, nil, "缓存名片"},
		{"friend_fallback", nil, nil, &FriendInfo{Uin: 42, Nickname: "好友昵称"}, "好友昵称"},
		{"unknown_id_fallback", nil, nil, nil, "42"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatcher := &senderMessageDispatcher{mockDispatcher: newMockDispatcher()}
			m := &Messenger{Adapter: &mockAdapter{}, eventDispatcher: dispatcher}
			if tc.member != nil {
				m.GroupList = []*GroupInfo{{Code: 100, Members: []*GroupMemberInfo{tc.member}}}
			}
			if tc.friend != nil {
				m.FriendList = []*FriendInfo{tc.friend}
			}
			m.handleGroupMessage(&GroupMessageEvent{GroupID: 100, UserID: 42, Sender: tc.incoming})
			if dispatcher.group == nil || dispatcher.group.Sender.DisplayName() != tc.want {
				t.Fatalf("群消息发送者 = %+v, want %q", dispatcher.group, tc.want)
			}
			if dispatcher.group.Sender.UserID != 42 || dispatcher.group.Sender.Uin != 42 {
				t.Fatal("群消息发送者 ID 丢失")
			}
			if tc.incoming != nil && dispatcher.group.Sender == tc.incoming {
				t.Fatal("消息转换应复制事件资料，避免改写共享输入")
			}
		})
	}
}

func TestPrivateMessageSenderProfileAndFriendFallback(t *testing.T) {
	for _, tc := range []struct {
		name     string
		incoming *SenderInfo
		friend   *FriendInfo
		stranger string
		want     string
	}{
		{"event_profile", &SenderInfo{Nickname: "事件昵称"}, &FriendInfo{Uin: 42, Nickname: "旧昵称"}, "", "事件昵称"},
		{"cached_friend", nil, &FriendInfo{Uin: 42, Nickname: "好友昵称"}, "", "好友昵称"},
		{"stranger_fallback", nil, nil, "陌生人昵称", "陌生人昵称"},
		{"unknown_id_fallback", nil, nil, "", "42"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatcher := &senderMessageDispatcher{mockDispatcher: newMockDispatcher()}
			m := &Messenger{
				Adapter:         &mockAdapter{strangerInfo: map[int64]map[string]interface{}{42: {"nickname": tc.stranger}}},
				eventDispatcher: dispatcher,
			}
			if tc.friend != nil {
				m.FriendList = []*FriendInfo{tc.friend}
			}
			m.handlePrivateMessage(&PrivateMessageEvent{UserID: 42, Sender: tc.incoming})
			if dispatcher.private == nil || dispatcher.private.Sender.DisplayName() != tc.want {
				t.Fatalf("私聊发送者 = %+v, want %q", dispatcher.private, tc.want)
			}
		})
	}
}
