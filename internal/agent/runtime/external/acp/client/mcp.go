package client

import (
	"strings"

	acp "github.com/coder/acp-go-sdk"

	mcpgw "github.com/kindredzhang/ami/internal/mcp"
)

const (
	amiToolsMCPServerName       = "Ami Tools"
	amiToolsMCPServerSlug       = "Ami_Tools"
	amiHeaderBotID              = mcpgw.ToolHeaderBotID
	amiHeaderChatID             = mcpgw.ToolHeaderChatID
	amiHeaderRuntimeID          = mcpgw.ToolHeaderRuntimeID
	amiHeaderRuntimeToken       = mcpgw.ToolHeaderRuntimeToken
	amiHeaderSessionID          = mcpgw.ToolHeaderSessionID
	amiHeaderRunID              = mcpgw.ToolHeaderRunID
	amiHeaderSessionType        = mcpgw.ToolHeaderSessionType
	amiHeaderRouteID            = mcpgw.ToolHeaderRouteID
	amiHeaderChannelIdentityID  = mcpgw.ToolHeaderChannelIdentityID
	amiHeaderCurrentPlatform    = mcpgw.ToolHeaderCurrentPlatform
	amiHeaderReplyTarget        = mcpgw.ToolHeaderReplyTarget
	amiHeaderConversationType   = mcpgw.ToolHeaderConversationType
	amiHeaderIsSubagent         = mcpgw.ToolHeaderIsSubagent
	amiHeaderSupportsImageInput = mcpgw.ToolHeaderSupportsImageInput
)

// AmiToolsHTTPMCPServer names the Ami tool gateway at rawURL, with the
// session's identity in headers for an endpoint that resolves it from them.
// It carries the identifier form of the name: an agent that qualifies tool
// names by server (`server__tool`) drops every tool of a server whose name is
// not an identifier. Permission preflight recognizes both forms. An empty URL
// yields the zero server.
func AmiToolsHTTPMCPServer(rawURL string, session mcpgw.ToolSessionContext) acp.McpServer {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return acp.McpServer{}
	}
	return acp.McpServer{
		Http: &acp.McpServerHttpInline{
			Name:    amiToolsMCPServerSlug,
			Url:     rawURL,
			Headers: amiToolsHTTPHeaders(session),
		},
	}
}

func amiToolsHTTPHeaders(session mcpgw.ToolSessionContext) []acp.HttpHeader {
	headers := make([]acp.HttpHeader, 0, 11)
	add := func(name, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		headers = append(headers, acp.HttpHeader{Name: name, Value: value})
	}

	add(amiHeaderBotID, session.BotID)
	add(amiHeaderChatID, session.ChatID)
	add(amiHeaderRuntimeID, session.RuntimeID)
	add(amiHeaderRuntimeToken, session.RuntimeToken)
	add(amiHeaderSessionID, session.SessionID)
	add(amiHeaderRunID, session.RunID)
	add(amiHeaderSessionType, session.SessionType)
	add(amiHeaderRouteID, session.RouteID)
	add(amiHeaderChannelIdentityID, session.ChannelIdentityID)
	add(amiHeaderCurrentPlatform, session.CurrentPlatform)
	add(amiHeaderReplyTarget, session.ReplyTarget)
	add(amiHeaderConversationType, session.ConversationType)
	if session.IsSubagent {
		add(amiHeaderIsSubagent, "true")
	}
	if session.SupportsImageInput {
		add(amiHeaderSupportsImageInput, "true")
	}
	return headers
}

// IsAmiToolsMCPServerName reports a name an agent may use for the Ami
// tool gateway, in either of the forms AmiToolsMCPServer and the
// identity-header variant give it.
func IsAmiToolsMCPServerName(name string) bool { return isAmiToolsMCPServerName(name) }

func isAmiToolsMCPServerName(name string) bool {
	name = strings.TrimSpace(name)
	return strings.EqualFold(name, amiToolsMCPServerName) ||
		strings.EqualFold(name, amiToolsMCPServerSlug)
}
