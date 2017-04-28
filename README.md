# DiscordConsole
The console that allows you to control both your user and bot account in interesting ways.

[Download Win/Mac/Linux 64-bit binaries](https://github.com/LEGOlord208/DiscordConsole/releases)

Or compile it yourself:
```
go get github.com/legolord208/DiscordConsole
```

![Imgur](http://i.imgur.com/ilOhYGb.png)

Type "help" at the prompt for a full list of commands.

I would be thrilled if you [joined the semi-official discord server](https://discord.gg/xvQV8bT)!

## Getting Started
To get started, simply
[Download Win/Mac/Linux 64-bit binaries](https://github.com/LEGOlord208/DiscordConsole/releases).

**If** you want to get the absolutely latest development update, you'll have to compile it yourself.  
Do that using
```
go get github.com/legolord208/DiscordConsole
```
.

You'll also need the DiscordGo development version. To get that, type `make dgo`.  
Then just type `make`, or `go install`. You choose.

## Special features
Set playing status, simulate typing, bulk delete and more.

In addition, you are even able to delete bot defined roles.  
If you try to do this in discord, it just says:  
![Imgur](http://i.imgur.com/Ubr2OMZ.png)

This was also discovered recently, by a friend of mine:  
You can bypass the black background in an avatar when setting the bot avatar.  
![Imgur](http://i.imgur.com/Q0GQR8d.png)

### Bulk delete
DiscordConsole lets you BULK DELETE messages. This allows you to delete a bunch of messages at once, without needing to write any code!  
Unfortunately, discord still requires you to have a bot account for this. Shame on you, discord!  

### Log
Log the last 100 messages to a file, or just view the last 10 directly in the console!

## Command line
DiscordConsole has full command line support. You can supply a bot/user token or email/password on the command line, and even specify commands to run, so you can use DiscordConsole for scripting.

Having it start with a specific server automatically selected? Making a cron job to automatically message how many days until Trump leaves? Easy!  
![Imgur](http://i.imgur.com/2mst4pH.png)  
*This could also be done with webhooks, but hush now :P*

# Have fun!
That is the most important of all.
