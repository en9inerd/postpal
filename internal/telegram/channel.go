package telegram

import (
	"encoding/json"
	"fmt"
)

// parseMessageResult parses the Result interface{} into a Message
func parseMessageResult(result interface{}) (*Message, error) {
	if result == nil {
		return nil, fmt.Errorf("result is nil")
	}

	msgBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var message Message
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	return &message, nil
}

// SendMessage sends a message to a channel
func (c *Client) SendMessage(req SendMessageRequest) (*Message, error) {
	resp, err := c.makeRequest("sendMessage", req)
	if err != nil {
		return nil, err
	}

	return parseMessageResult(resp.Result)
}

// EditMessageText edits the text of a message in a channel
func (c *Client) EditMessageText(req EditMessageTextRequest) (*Message, error) {
	resp, err := c.makeRequest("editMessageText", req)
	if err != nil {
		return nil, err
	}

	return parseMessageResult(resp.Result)
}

// EditMessageCaption edits the caption of a message in a channel
func (c *Client) EditMessageCaption(req EditMessageCaptionRequest) (*Message, error) {
	resp, err := c.makeRequest("editMessageCaption", req)
	if err != nil {
		return nil, err
	}

	return parseMessageResult(resp.Result)
}

// EditMessageMedia edits the media of a message in a channel
func (c *Client) EditMessageMedia(req EditMessageMediaRequest) (*Message, error) {
	resp, err := c.makeRequest("editMessageMedia", req)
	if err != nil {
		return nil, err
	}

	return parseMessageResult(resp.Result)
}

// DeleteMessage deletes a message from a channel
func (c *Client) DeleteMessage(req DeleteMessageRequest) (bool, error) {
	resp, err := c.makeRequest("deleteMessage", req)
	if err != nil {
		return false, err
	}

	// deleteMessage returns a boolean result
	if resp.Result == nil {
		// If result is nil but OK is true, deletion was successful
		return resp.OK, nil
	}

	return true, nil
}

// ForwardMessage forwards a message to a channel
func (c *Client) ForwardMessage(req ForwardMessageRequest) (*Message, error) {
	resp, err := c.makeRequest("forwardMessage", req)
	if err != nil {
		return nil, err
	}

	return parseMessageResult(resp.Result)
}

// CopyMessage copies a message to a channel
func (c *Client) CopyMessage(req CopyMessageRequest) (*Message, error) {
	resp, err := c.makeRequest("copyMessage", req)
	if err != nil {
		return nil, err
	}

	return parseMessageResult(resp.Result)
}

// PinChatMessage pins a message in a channel
func (c *Client) PinChatMessage(req PinChatMessageRequest) (bool, error) {
	resp, err := c.makeRequest("pinChatMessage", req)
	if err != nil {
		return false, err
	}

	return resp.OK, nil
}

// UnpinChatMessage unpins a specific message in a channel
// If MessageID is 0, it will unpin all messages
func (c *Client) UnpinChatMessage(req UnpinChatMessageRequest) (bool, error) {
	resp, err := c.makeRequest("unpinChatMessage", req)
	if err != nil {
		return false, err
	}

	return resp.OK, nil
}

// UnpinAllChatMessages unpins all messages in a channel
func (c *Client) UnpinAllChatMessages(req UnpinAllChatMessagesRequest) (bool, error) {
	resp, err := c.makeRequest("unpinAllChatMessages", req)
	if err != nil {
		return false, err
	}

	return resp.OK, nil
}
