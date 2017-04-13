# Before talk
This was made by CubityFirst.  
Thanks a lot!

# Welcome
... to the markdowntestcubity wiki!

# Commands : 

## Basic :

**help** - Show's the help menu

**exit** - Exit DiscordConsole

**exec** - Execute a shell command (Somewhat-advanced)


## Channel/Guild Selectors :

**guilds** - List guilds/servers this bot is added to.

**guild** *id* - Select a guild to use for further commands.

**channels** - List channels in your selected guild.

**channel** *id* - Select a channel to use for further commands.

**pchannels** - List private channels a.k.a. 'DMs'.

**dm** **id** - Create a DM with specific user.

## Messaging Commands :
**say** *message* - Send a message in your selected channel.

**file** *filepath* - Upload file to selected channel.

**edit** *message id* *stuff* - Edit a message in your selected channel.

**del** *message id* - Delete a message in the selected channel.

**delall** *since message id* - Bulk delete messages since a specific message

**log** *output file* - Log the last few messages in console or to a file.

**typing** - Simulate typing in selected channel... (last's for about 10 seconds)

**nick** *nickname* Change own nickname.

## Member Bar : 

**playing** *game* - Set your playing status.

**streaming** *twitchurl* *gamename* - Set your streaming status

**members** - List (max 100) members in selected guild

**invite** - Create (permanent) instant invite. (Exports as the small code such as 83GZDqv)

**leave** - Leave's the current server

## Roles : 

**roles**   List all roles in selected guild.

**roleadd** <user id> <role id>     Add role to user

**roledel** <user id> <role id>     Remove role from user

**roleedit** *id* *flag* *value* - Changes role depending on flag used, see below.

### Role Flags : 
**name** - Changes the name of the roles.

**color** - Needs Hexadecimal (For example \#0FADED)

**separate** - Seperates role from others (Needs yes or no)

**perms** - Use https://discordapi.com/permissions.html to get the correct permissions

**mention** - Allows role to be mentionable or not. (Needs yes or no)

## Moderation : 

**kick** *id* - Kick's the ID

**ban** *id* - Ban's the ID

**unban** *id* - Unban's the ID

**nickall** *name* - Nickname's everyone (use no arguments to clear)


## QuickMove + ConsoleLog
**reply** - Jump to the channel of the last received message (Shown in the console)

**back** - Jump to previous guild and/or channel (from moving with **channel** or **guild**)

**enablemessages** - Enable intercepting messages

**disablemessages** - Reverts the above.
