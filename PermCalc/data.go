package permcalc

// First few permission constants
// (All permission constants are from discordgo)
const (
	PermCreateInstantInvite = 1 << iota
	PermKickMembers
	PermBanMembers
	PermAdministrator
	PermManageChannels
	PermManageServer
	PermAddReactions
	PermViewAuditLogs
)

// Second batch of permission constants
const (
	PermChangeNickname = 1 << (iota + 26)
	PermManageNicknames
	PermManageRoles
	PermManageWebhooks
	PermManageEmojis
)

// Third batch of permission constants
const (
	PermReadMessages = 1 << (iota + 10)
	PermSendMessages
	PermSendTTSMessages
	PermManageMessages
	PermEmbedLinks
	PermAttachFiles
	PermReadMessageHistory
	PermMentionEveryone
	PermUseExternalEmojis
)

// Last few permission constants
const (
	PermConnect = 1 << (iota + 20)
	PermSpeak
	PermMuteMembers
	PermDeafenMembers
	PermMoveMembers
	PermUseVoiceActivity
)

// PermOrder is all permissions, sorted by order.
var PermOrder = []int{
	// General Permissions
	PermAdministrator,
	PermViewAuditLogs,
	PermManageRoles,
	PermKickMembers,
	PermCreateInstantInvite,
	PermManageNicknames,
	PermManageWebhooks,

	PermManageServer,
	PermManageChannels,
	PermBanMembers,
	PermChangeNickname,
	PermManageEmojis,

	// Text Permissions
	PermReadMessages,
	PermSendTTSMessages,
	PermEmbedLinks,
	PermReadMessageHistory,
	PermUseExternalEmojis,

	PermSendMessages,
	PermManageMessages,
	PermAttachFiles,
	PermMentionEveryone,
	PermAddReactions,

	// Voice Permissions
	PermConnect,
	PermMuteMembers,
	PermMoveMembers,

	PermSpeak,
	PermDeafenMembers,
	PermUseVoiceActivity,
}

// PermStrings the name for all permissions.
var PermStrings = map[int]string{
	// General Permissions
	PermAdministrator:       "Administrator",
	PermViewAuditLogs:       "View Audit Logs",
	PermManageRoles:         "Manage Roles",
	PermKickMembers:         "Kick Members",
	PermCreateInstantInvite: "Create Instant Invite",
	PermManageNicknames:     "Manage Nicknames",
	PermManageWebhooks:      "Manage Webhooks",

	PermManageServer:   "Manage Server",
	PermManageChannels: "Manage Channels",
	PermBanMembers:     "Ban Members",
	PermChangeNickname: "Change Nickname",
	PermManageEmojis:   "Manage Emojis",

	// Text Permissions
	PermReadMessages:       "Read Messages",
	PermSendTTSMessages:    "Send TTS Messages",
	PermEmbedLinks:         "Embed Links",
	PermReadMessageHistory: "Read Message History",
	PermUseExternalEmojis:  "Use External Emojis",

	PermSendMessages:    "Send Messages",
	PermManageMessages:  "Manage Messages",
	PermAttachFiles:     "Attach Files",
	PermMentionEveryone: "Mention Everyone",
	PermAddReactions:    "Add Reactions",

	// Voice Permissions
	PermConnect:     "Connect",
	PermMuteMembers: "Mute Members",
	PermMoveMembers: "Move Members",

	PermSpeak:            "Speak",
	PermDeafenMembers:    "Deafen Members",
	PermUseVoiceActivity: "Use Voice Activity",
}
