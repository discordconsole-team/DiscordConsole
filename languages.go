package main

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

var ErrLangCorrupt = errors.New("Corrupt language file")
var lang map[string]string

func loadLang(reader io.Reader) error {
	lang = make(map[string]string)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		parts := strings.SplitN(text, "=", 2)
		if len(parts) != 2 {
			return ErrLangCorrupt
		}
		key := parts[0]
		val := parts[1]

		if strings.HasSuffix(key, ".dev") && DevVersion {
			key = key[:len(key)-len(".dev")]
		}

		lang[key] = val
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
func loadLangString(lang string) error {
	return loadLang(strings.NewReader(lang))
}
func loadLangDefault() {
	loadLangString(LangEn)
}

// Here is just some long data.
// This comment is a separator, btw.

var LangEn = `
update.checking=Checking for updates...
update.error=Error checking for updates
update.available=Update available! Version
update.available.dev=Latest stable release:
update.download=Download from
update.none=No updates found.

loading.bookmarks=Reading bookmarks...

failed.reading=Could not read
failed.realine.start=Could not start readline library
failed.realine.read=Could not read line
failed.auth=Couldn't authenticate
failed.session.start=Could not open session
failed.session.close=Could not close session
failed.perms=No permissions to perform this action.
failed.path.home=Could not determine value of ~
failed.user=Couldn't query user
failed.user.edit=Couldn't edit user data
failed.channel=Could not query channel
failed.guild=Could not query guild
failed.timestamp=Couldn't parse timestamp
failed.channel.create=Could not create channel
failed.msg.query=Could not get message
failed.msg.send=Could not send message
failed.msg.edit=Couldn't edit message
failed.msg.delete=Couldn't delete message
failed.msg.notfound=Message not found!
failed.lua.run=Could not run lua
failed.lua.event=Recovered from LUA error
failed.voice.connect=Could not connect to voice channel
failed.voice.speak=Could not start speaking
failed.voice.disconnect=Could not disconnect
failed.exec=Could not execute
failed.fixpath=Could not 'fix' filepath
failed.file.open=Couldn't open file
failed.file.write=Could not write file
failed.file.read=Could not read file
failed.file.load=Could not load file.
failed.file.save=Could not save file.
failed.status=Couldn't update status
failed.typing=Couldn't start typing
failed.members=Could not list members
failed.invite.accept=Could not accept invite
failed.invite.create=Invite could not be created
failed.roles=Could not get roles
failed.role.change=Could not add/remove role
failed.role.create=Could not create role
failed.role.edit=Could not edit role
failed.role.delete=Could not delete role!
failed.nick=Could not set nickname
failed.ban.create=Could not ban user
failed.ban.delete=Could not unban user
failed.ban.list=Could not list bans
failed.kick=Could not kick user
failed.leave=Could not leave
failed.block=Couldn't block user
failed.friends=Couldn't get friends
failed.json=Could not parse json
failed.base64=Couldn't convert to Base64
failed.react=Could not react to message
failed.react.used=Emoji used already, skipping
failed.webrequest=Could not make web request
failed.avatar=Couldn't set avatar
failed.status=Could not set status

invalid.yn=Please type either 'y' or 'n'.
invalid.webhook=Webhook format invalid. Format: id/token
invalid.webhook.command=Not an allowed webhook command
invalid.limit.message=Message exceeds character limit
invalid.channel=No channel selected!
invalid.guild=No guild selected!
invalid.value=No such value
invalid.role=No role with that ID
invalid.number=Not a number
invalid.cache=No cache available!
invalid.onlyfor.users=This only works for users.
invalid.onlyfor.bots=This command only works for bot users.
invalid.music.playing=Already playing something
invalid.bookmark=Bookmark doesn't exist
invalid.status.offline=The offline status exists, but cannot be set through the API
invalid.command=Unknown command. Do 'help' for help

login.detect=You are logged into Discord. Use that login? (y/n):
login.token=Please paste your bot 'token' here, or leave blank for a username/password prompt.
login.token.user=User tokens are prefixed with 'user '
login.token.webhook=Webhook tokens are prefixed with 'webhook ', and their URL or id/token
login.starting=Authenticating...
login.finish=Logged in with user ID
intro.help=Write 'help' for help
intro.exit=Press Ctrl+D or type 'exit' to exit.

pointer.unknown=Unknown
pointer.private=Private

status.msg.create=Created message with ID
status.msg.intercept=Messages will now be intercepted.
status.msg.nointercept=Messages will no longer be intercepted.
status.cmd.intercept='console.' commands will now be intercepted.
status.cmd.nointercept='console.' commands will no longer be intercepted.
status.channel=Selected channel with ID
status.invite.accept=Accepted invite.
status.invite.create=Created invite with code:
status.cache=Message cached!
status.loading=Loading...
status.avatar=Avatar set!
status.name=Name set!
status.status=Status set!

restarting.session=Restarting session...
restarting.cache.loc=Reloading location cache...
restarting.cache.vars=Deleting cache variables...
`
