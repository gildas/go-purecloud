package conversation

import (
	"fmt"
	"encoding/json"

	"github.com/gildas/go-logger"
	"github.com/gildas/go-purecloud"
	"github.com/gorilla/websocket"
)

type createPayload struct {
	OrganizationID string `json:"organizationId"`
	DeploymentID   string `json:"deploymentId"`
	RoutingTarget  Target `json:"routingTarget"`
	Member         Member `json:"memberInfo"`
}

// Create creates a new chat Conversation in PureCloud
func Create(client *purecloud.Client, target Target, member Member) (*Conversation, error) {
	// TODO sanitizing...
	payload, err := json.Marshal(createPayload{
		OrganizationID: client.Organization.ID,
		DeploymentID:   client.DeploymentID,
		RoutingTarget:  target,
		Member:         member,
	})
	if err != nil {
		return nil, err
	}

	conversation := &Conversation{Client: client, Members: make(map[string]*Member)}
	if err = client.Post("webchat/guest/conversations", payload, &conversation); err != nil {
		return nil, err
	}
	conversation.Logger = client.Logger.Record("topic", "conversation").Record("scope", "conversation").Record("conversation", conversation.ID).Child().(*logger.Logger)

	conversation.Socket, _, err = websocket.DefaultDialer.Dial(conversation.EventStream, nil)
	if err != nil {
		conversation.Logger.Errorf("Failed while connecting to %s", conversation.EventStream, err)
		conversation.Close()
		return nil, err
	}
	return conversation, nil
}

// Close terminates a conversation (its websocket as well)
func (conversation *Conversation) Close() error {
	log := conversation.Logger.Record("scope", "close").Child().(*logger.Logger)

	if conversation.Socket != nil {
		log.Debugf("Disconnecting websocket")
		if err := conversation.Socket.Close(); err != nil {
			log.Errorf("Failed while closing websocket", err)
			return err
		}
		log.Infof("Websocket disconnected")
	} else if conversation.Client != nil {
		log.Debugf("Guest Member leaving")
		if err := conversation.Client.Delete(fmt.Sprintf("webchat/guest/conversations/%s/members/%s", conversation.ID, conversation.Member.ID), nil, nil); err != nil {
			log.Errorf("Failed while guest member was leaving chat", err)
			return err
		}
		log.Infof("Guest Member left")
	}
	conversation.Socket = nil
	conversation.Client = nil
	return nil
}