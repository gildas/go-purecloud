package conversation

import (
	"encoding/json"

	"github.com/gildas/go-logger"
	"github.com/gildas/go-purecloud"
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
	return conversation, nil
}