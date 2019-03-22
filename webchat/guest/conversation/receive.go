package conversation

import (
	"os"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gildas/go-logger"
)

// MessageHandlers holds the various callbacks used when receiving messages
type MessageHandlers struct {
	OnClosed       func(conversation *Conversation, message *Message, member *Member)
	OnStateChanged func(conversation *Conversation, message *Message, member *Member)
	OnMessage      func(conversation *Conversation, message *Message, member *Member)
	OnTyping       func(conversation *Conversation, message *Message, member *Member)
}
// HandleMessages is the incoming message loop
func (conversation *Conversation) HandleMessages(handlers MessageHandlers) (err error) {
	if conversation.Socket == nil {
		return fmt.Errorf("Conversation Not Connected")
	}

	log := conversation.Logger.Record("scope", "receive").Child().(*logger.Logger)

	for {
		// get a message body and decode it. (ReadJSON is nice, but in case of unknown message, I cannot get the original string)
		var body []byte

		if _, body, err = conversation.Socket.ReadMessage(); err != nil {
			log.Errorf("Failed to read incoming message", err)
			continue // TODO: Should we bail out?!?
		}

		message := &Message{}
		if err = json.Unmarshal(body, &message); err != nil {
			log.Errorf("Malformed JSON message: %s", body, err)
			continue
		}
		message.Logger = log.Record("correlation", message.Metadata.CorrelationID).Record("message", message.EventBody.ID).Child().(*logger.Logger)

		switch strings.ToLower(message.TopicName) {
		case "channel.metadata":
			if message.EventBody.Message == "WebSocket Heartbeat" {
				// Since this adds a lot to the logs, log heartbeat only if the environment demands it
				// TODO: Document this
				if _, ok := os.LookupEnv("PURECLOUD_TRACE_HEARTBEAT"); ok {
					message.Logger.Debugf("<< %s", message.EventBody.Message)
				}
			} else {
				message.Logger.Warnf("Unknown: %s, \n%s,\n%+v", message.TopicName, body, message)
			}

		case "v2.conversations.chats." + conversation.ID + ".members":
			switch strings.ToLower(message.Metadata.Type) {
			case "member-change":
				message.Logger.Record("correlation", message.Metadata.CorrelationID).Debugf("Timestamp %s", message.EventBody.Timestamp)
				member, err := conversation.GetMember(message.EventBody.Member.ID)
				if err != nil {
					message.Logger.Errorf("Failed to get member info for %s", message.EventBody.Member.ID, err)
					member = &Member{
						ID:    message.EventBody.Member.ID,
						State: message.EventBody.Member.State,
					}
				}
				message.Logger.Debugf("State Change for %s Member %s (%s): %s at %s", member.Role, member.ID, member.DisplayName, member.State, message.EventBody.Timestamp)
				// If the chat guest disconnected, the whole chat should close
				if message.EventBody.Member.ID == conversation.Member.ID && message.EventBody.Member.State == "DISCONNECTED" {
					defer conversation.Close()
					if handlers.OnClosed != nil {
						handlers.OnClosed(conversation, message, member)
					}
					return nil // Break the incoming message loop
				}
				if handlers.OnStateChanged != nil {
					handlers.OnStateChanged(conversation, message, member)
				}
			default:
				return fmt.Errorf("Unknown Metadata %s", message.Metadata.Type)
			}

		case "v2.conversations.chats." + conversation.ID + ".messages":
			sender, err := conversation.GetMember(message.EventBody.Sender.ID)
			if err != nil {
				message.Logger.Errorf("Failed to get sender info for %s", message.EventBody.Sender.ID, err)
				sender = &Member{ ID: message.EventBody.Sender.ID }
			}
			switch strings.ToLower(message.Metadata.Type) {
			case "message":
				// TODO: Do NOT send the same message twice!
				message.Logger.Debugf("Message from %s (%s) at %s", sender.ID, sender.DisplayName, message.EventBody.Timestamp)
				if sender.ID != conversation.Member.ID && handlers.OnMessage != nil {
					handlers.OnMessage(conversation, message, sender)
				}
			case "typing-indicator":
				// TODO: Do NOT send the same message twice!
				message.Logger.Debugf("Typing Indicator from %s (%s) at %s", sender.ID, sender.DisplayName, message.EventBody.Timestamp)
				if handlers.OnMessage != nil {
					handlers.OnTyping(conversation, message, sender)
				}
			default:
				message.Logger.Warnf("Unknown: %s, \n%s, \n%+v", message.Metadata.Type, body, message)
			}
		default:
			message.Logger.Warnf("Unknown: %s, \n%s, \n%+v", message.TopicName, body, message)
		}
	}
}