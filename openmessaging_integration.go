package gcloudcx

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/google/uuid"
)

// OpenMessagingIntegration  describes an GCloud OpenMessaging Integration
//
// See https://developer.genesys.cloud/api/digital/openmessaging
type OpenMessagingIntegration struct {
	ID               uuid.UUID             `json:"id"`
	Name             string                `json:"name"`
	WebhookURL       *url.URL              `json:"-"`
	WebhookToken     string                `json:"outboundNotificationWebhookSignatureSecretToken"`
	Recipient        *DomainEntityRef      `json:"recipient,omitempty"`
	SupportedContent *AddressableEntityRef `json:"supportedContent,omitempty"`
	DateCreated      time.Time             `json:"dateCreated,omitempty"`
	CreatedBy        *DomainEntityRef      `json:"createdBy,omitempty"`
	DateModified     time.Time             `json:"dateModified,omitempty"`
	ModifiedBy       *DomainEntityRef      `json:"modifiedBy,omitempty"`
	CreateStatus     string                `json:"createStatus,omitempty"`
	CreateError      *ErrorBody            `json:"createError,omitempty"`
	SelfURI          URI                   `json:"selfUri,omitempty"`
	Client           *Client               `json:"-"`
	Logger           *logger.Logger        `json:"-"`
}

// Initialize initializes this from the given Client
//
//   if the parameters contain a uuid.UUID, the corresponding integration is fetched
//
//   implements Initializable
func (integration *OpenMessagingIntegration) Initialize(parameters ...interface{}) error {
	client, logger, id, err := parseParameters(integration, parameters...)
	if err != nil {
		return err
	}
	if id != uuid.Nil {
		if err := client.Get(NewURI("/conversations/messaging/integrations/open/%s", id), &integration); err != nil {
			return err
		}
	}
	integration.Client = client
	integration.Logger = logger.Child("openmessagingintegration", "openmessagingintegration", "openmesssagingintegration", integration.ID)
	return nil
}

// FetchOpenMessagingIntegrations Fetches all OpenMessagingIntegration object
func FetchOpenMessagingIntegrations(parameters ...interface{}) ([]*OpenMessagingIntegration, error) {
	client, logger, _, err := parseParameters(nil, parameters...)
	if err != nil {
		return nil, err
	}
	response := struct {
		Integrations []*OpenMessagingIntegration `json:"entities"`
		PageSize     int                         `json:"pageSize"`
		PageNumber   int                         `json:"pageNumber"`
		PageCount    int                         `json:"pageCount"`
		PageTotal    int                         `json:"total"`
		FirstURI     string                      `json:"firstUri"`
		SelfURI      string                      `json:"selfUri"`
		LastURI      string                      `json:"lastUri"`
	}{}
	if err = client.Get("/conversations/messaging/integrations/open", &response); err != nil {
		return nil, err
	}
	logger.Record("response", response).Infof("Got a response")
	for _, integration := range response.Integrations {
		integration.Client = client
		integration.Logger = logger.Child("openmessagingintegration", "openmessagingintegration", "openmesssagingintegration", integration.ID)
	}
	return response.Integrations, nil
}

// FetchOpenMessagingIntegration Fetches an OpenMessagingIntegration object
//
// If a UUID is given, fetches by UUID
//
// If a string is given, fetches by name
func FetchOpenMessagingIntegration(parameters ...interface{}) (*OpenMessagingIntegration, error) {
	client, logger, id, err := parseParameters(nil, parameters...)
	if err != nil {
		return nil, err
	}

	integration := &OpenMessagingIntegration{}
	if id != uuid.Nil {
		if err := client.Get(NewURI("/conversations/messaging/integrations/open/%s", id), &integration); err != nil {
			return nil, err
		}
	} else {
		var name string
		for _, parameter := range parameters {
			switch object := parameter.(type) {
			case string:
				name = object
			}
		}
		if len(name) == 0 {
			return nil, errors.ArgumentMissing.With("name").WithStack()
		}
		response := struct {
			Integrations []*OpenMessagingIntegration `json:"entities"`
			PageSize     int                         `json:"pageSize"`
			PageNumber   int                         `json:"pageNumber"`
			PageCount    int                         `json:"pageCount"`
			PageTotal    int                         `json:"total"`
			FirstURI     string                      `json:"firstUri"`
			SelfURI      string                      `json:"selfUri"`
			LastURI      string                      `json:"lastUri"`
		}{}
		if err = client.Get("/conversations/messaging/integrations/open", &response); err != nil {
			return nil, err
		}
		nameLowercase := strings.ToLower(name)
		for _, item := range response.Integrations {
			if strings.Compare(strings.ToLower(item.Name), nameLowercase) == 0 {
				integration = item
				break
			}
		}
		if integration == nil || integration.ID == uuid.Nil {
			return nil, errors.NotFound.With("name", name).WithStack()
		}
	}
	integration.Client = client
	integration.Logger = logger.Child("openmessagingintegration", "openmessagingintegration", "openmessagingintegration", integration.ID)
	return integration, nil
}

// Create creates a new OpenMessaging Integration
func (integration *OpenMessagingIntegration) Create(name string, webhookURL *url.URL, token string) error {
	response := &OpenMessagingIntegration{}
	err := integration.Client.Post(
		"/conversations/messaging/integrations/open",
		struct {
			Name    string `json:"name"`
			Webhook string `json:"outboundNotificationWebhookUrl"`
			Token   string `json:"outboundNotificationWebhookSignatureSecretToken"`
		}{
			Name:    name,
			Webhook: webhookURL.String(),
			Token:   token,
		},
		&response,
	)
	if err != nil {
		return err
	}
	integration.Logger.Record("response", response).Debugf("Created integration %#v", response)
	integration.ID = response.ID
	return nil
}

// Delete deletes an OpenMessaging Integration
//
// If the integration was not created, nothing is done
func (integration *OpenMessagingIntegration) Delete() error {
	if integration.ID == uuid.Nil {
		return nil
	}
	return integration.Client.Delete(NewURI("/conversations/messaging/integrations/open/%s", integration.ID), nil)
}

// Update updates an OpenMessaging Integration
//
// If the integration was not created, an error is return without reaching GENESYS Cloud
func (integration *OpenMessagingIntegration) Update(name string, webhookURL *url.URL, token string) error {
	if integration.ID == uuid.Nil {
		return errors.ArgumentMissing.With("ID").WithStack()
	}
	response := &OpenMessagingIntegration{}
	err := integration.Client.Patch(
		NewURI("/conversations/messaging/integrations/open/%s", integration.ID),
		struct {
			Name    string `json:"name"`
			Webhook string `json:"outboundNotificationWebhookUrl"`
			Token   string `json:"outboundNotificationWebhookSignatureSecretToken"`
		}{
			Name:    name,
			Webhook: webhookURL.String(),
			Token:   token,
		},
		&response,
	)
	if err != nil {
		return errors.CreationFailed.Wrap(err)
	}
	integration.Logger.Record("response", response).Debugf("Updated integration %#v", response)
	return nil
}

// SendInboundTextMessage sends a text message from the middleware to GENESYS Cloud
//
// See https://developer.genesys.cloud/api/digital/openmessaging/inboundMessages#send-an-inbound-open-message
func (integration *OpenMessagingIntegration) SendInboundTextMessage(from *OpenMessageFrom, messageID, text string) (*OpenMessageResult, error) {
	if integration.ID == uuid.Nil {
		return nil, errors.ArgumentMissing.With("ID").WithStack()
	}
	result := &OpenMessageResult{}
	err := integration.Client.Post(
		"/conversations/messages/inbound/open",
		&OpenMessage{
			Direction: "Inbound",
			Channel: NewOpenMessageChannel(
				messageID,
				&OpenMessageTo{ ID: integration.ID.String() },
				from,
			),
			Type: "Text",
			Text: text,
		},
		&result,
	)
	return result, err
}

// SendInboundImageMessage sends a text message from the middleware to GENESYS Cloud
//
// See https://developer.genesys.cloud/api/digital/openmessaging/inboundMessages#inbound-message-with-attached-photo
func (integration *OpenMessagingIntegration) SendInboundImageMessage(from *OpenMessageFrom, messageID, text string, imageMimeType string, imageURL *url.URL) (*OpenMessageResult, error) {
	if integration.ID == uuid.Nil {
		return nil, errors.ArgumentMissing.With("ID").WithStack()
	}
	result := &OpenMessageResult{}
	err := integration.Client.Post(
		"/conversations/messages/inbound/open",
		&OpenMessage{
			Direction: "Inbound",
			Channel: NewOpenMessageChannel(
				messageID,
				&OpenMessageTo{ ID: integration.ID.String() },
				from,
			),
			Type: "Text",
			Text: text,
			Content: []*OpenMessageContent{
				{
					Type: "Attachment",
					Attachment: &OpenMessageAttachment{
						Type: "Image",
						Mime: imageMimeType,
						URL:  imageURL,
					},
				},
			},
		},
		&result,
	)
	return result, err
}

// SendOutboundMessage sends a message from GENESYS Cloud to the middleware
//
// The message can be only text as it is sent bia the AgentLess Message API.
//
// This is mainly for debugging purposes
//
// See https://developer.genesys.cloud/api/digital/openmessaging/outboundMessages#send-an-agentless-outbound-text-message
func (integration *OpenMessagingIntegration) SendOutboundMessage(destination, text string) (*AgentlessMessageResult, error) {
	if integration.ID == uuid.Nil {
		return nil, errors.ArgumentMissing.With("ID").WithStack()
	}
	result := &AgentlessMessageResult{}
	err := integration.Client.Post(
		"/conversations/messages/agentless",
		AgentlessMessage{
			From:          integration.ID.String(),
			To:            destination,
			MessengerType: "Open",
			Text:          text,
		},
		&result,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetID gets the identifier of this
//
//   implements Identifiable
func (integration OpenMessagingIntegration) GetID() uuid.UUID {
	return integration.ID
}

// String gets a string version
//
//   implements the fmt.Stringer interface
func (integration OpenMessagingIntegration) String() string {
	if len(integration.Name) > 0 {
		return integration.Name
	}
	return integration.ID.String()
}

// MarshalJSON marshals this into JSON
func (integration OpenMessagingIntegration) MarshalJSON() ([]byte, error) {
	type surrogate OpenMessagingIntegration
	data, err := json.Marshal(struct {
		surrogate
		W *core.URL `json:"outboundNotificationWebhookUrl"`
	}{
		surrogate: surrogate(integration),
		W:         (*core.URL)(integration.WebhookURL),
	})
	return data, errors.JSONMarshalError.Wrap(err)
}

// UnmarshalJSON unmarshals JSON into this
func (integration *OpenMessagingIntegration) UnmarshalJSON(payload []byte) (err error) {
	type surrogate OpenMessagingIntegration
	var inner struct {
		surrogate
		W *core.URL `json:"outboundNotificationWebhookUrl"`
	}

	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*integration = OpenMessagingIntegration(inner.surrogate)
	integration.WebhookURL = (*url.URL)(inner.W)
	return
}

// TODO: There is also a PATCH method... we might want to provide some func
