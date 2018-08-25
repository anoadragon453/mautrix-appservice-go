package appservice

import (
	"fmt"
	"maunium.net/go/gomatrix"
)

type IntentAPI struct {
	*gomatrix.Client
	bot       *gomatrix.Client
	as        *AppService
	Localpart string
	UserID    string
}

func (as *AppService) NewIntentAPI(localpart string) *IntentAPI {
	userID := fmt.Sprintf("@%s:%s", localpart, as.HomeserverDomain)
	bot := as.BotClient()
	if userID == bot.UserID {
		bot = nil
	}
	return &IntentAPI{
		Client:    as.Client(userID),
		bot:       bot,
		as:        as,
		Localpart: localpart,
		UserID:    userID,
	}
}

func (intent *IntentAPI) Register() error {
	_, _, err := intent.Client.Register(&gomatrix.ReqRegister{
		Username: intent.Localpart,
	})
	if err != nil {
		return err
	}
	return nil
}

func (intent *IntentAPI) EnsureRegistered() error {
	if intent.as.StateStore.IsRegistered(intent.UserID) {
		return nil
	}

	err := intent.Register()
	httpErr, ok := err.(gomatrix.HTTPError)
	if !ok || httpErr.RespError.ErrCode != "M_USER_IN_USE" {
		return err
	}
	intent.as.StateStore.MarkRegistered(intent.UserID)
	return nil
}

func (intent *IntentAPI) EnsureJoined(roomID string) error {
	if intent.as.StateStore.IsInRoom(intent.UserID, roomID) {
		return nil
	}

	if err := intent.EnsureRegistered(); err != nil {
		return err
	}

	resp, err := intent.JoinRoom(roomID, "", nil)
	if err != nil {
		httpErr, ok := err.(gomatrix.HTTPError)
		if !ok || httpErr.RespError.ErrCode != "M_FORBIDDEN" || intent.bot == nil {
			return httpErr
		}
		_, inviteErr := intent.bot.InviteUser(roomID, &gomatrix.ReqInviteUser{
			UserID: intent.UserID,
		})
		if inviteErr != nil {
			return err
		}
		resp, err = intent.JoinRoom(roomID, "", nil)
		if err != nil {
			return err
		}
	}
	intent.as.StateStore.SetMembership(resp.RoomID, intent.UserID, "join")
	return nil
}

func (intent *IntentAPI) SendMessageEvent(roomID string, eventType string, contentJSON interface{}) (*gomatrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendMessageEvent(roomID, eventType, contentJSON)
}

func (intent *IntentAPI) SendMassagedMessageEvent(roomID string, eventType string, contentJSON interface{}, ts int64) (*gomatrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendMassagedMessageEvent(roomID, eventType, contentJSON, ts)
}

func (intent *IntentAPI) SendStateEvent(roomID, eventType, stateKey string, contentJSON interface{}) (*gomatrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendStateEvent(roomID, eventType, stateKey, contentJSON)
}

func (intent *IntentAPI) SendMassagedStateEvent(roomID, eventType, stateKey string, contentJSON interface{}, ts int64) (*gomatrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendMassagedStateEvent(roomID, eventType, stateKey, contentJSON, ts)
}

func (intent *IntentAPI) SendText(roomID, text string) (*gomatrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendText(roomID, text)
}

func (intent *IntentAPI) SendImage(roomID, body, url string) (*gomatrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendImage(roomID, body, url)
}

func (intent *IntentAPI) SendVideo(roomID, body, url string) (*gomatrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendVideo(roomID, body, url)
}

func (intent *IntentAPI) SendNotice(roomID, text string) (*gomatrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.SendNotice(roomID, text)
}

func (intent *IntentAPI) RedactEvent(roomID, eventID string, req *gomatrix.ReqRedact) (*gomatrix.RespSendEvent, error) {
	if err := intent.EnsureJoined(roomID); err != nil {
		return nil, err
	}
	return intent.Client.RedactEvent(roomID, eventID, req)
}

func (intent *IntentAPI) SetRoomName(roomID, roomName string) (*gomatrix.RespSendEvent, error) {
	return intent.SendStateEvent(roomID, "m.room.name", "", map[string]interface{}{
		"name": roomName,
	})
}

func (intent *IntentAPI) SetRoomAvatar(roomID, avatarURL string) (*gomatrix.RespSendEvent, error) {
	return intent.SendStateEvent(roomID, "m.room.avatar", "", map[string]interface{}{
		"url": avatarURL,
	})
}

func (intent *IntentAPI) SetRoomTopic(roomID, topic string) (*gomatrix.RespSendEvent, error) {
	return intent.SendStateEvent(roomID, "m.room.topic", "", map[string]interface{}{
		"topic": topic,
	})
}

func (intent *IntentAPI) SetDisplayName(displayName string) error {
	if err := intent.EnsureRegistered(); err != nil {
		return err
	}
	return intent.Client.SetDisplayName(displayName)
}

func (intent *IntentAPI) SetAvatarURL(avatarURL string) error {
	if err := intent.EnsureRegistered(); err != nil {
		return err
	}
	return intent.Client.SetAvatarURL(avatarURL)
}
