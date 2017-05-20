# DiscordConsole rewrite

DiscordConsole has got some hairy code recently.  
It's time for a rewrite.

Don't worry, the old version will still be supported.

## Changes

First big change:  
**You will be navigating in a UI.**  
That said, the commands system will still exist, and it is how it all functions behind the scenes.  
You will also be able to disable the UI and get a normal commands dialog.  
Speaking of that, commands will also use quotes to separate arguments.  
*Example: `nick @me "some good nickname"`*

Secondly,
you will be able to pass multiple `-t` or `--token` (long option is also new) parameters at startup, to be able to  
switch between tokens in one instance.  
However, this would slow down startup time.

To reduce this problem, you will be able to run it with `--novalidation`, to disable account validation on startup.  
Please keep in mind however, this will let invalid tokens fail silently.  
~~There will also be stuff like `--nowebsocket` and similar, to speedup startup time even more.~~  
EDIT: No, actually. There will **MAYBE** not be a `--nowebsocket`. It's because we're gonna use that to make less web requests!

Because of the startup time fixes, support for LUA, API and selfbot *might* be dropped.  
Instead, you would be expected to run commands by starting a new process.  
We might even add official scripts to do these things.

User tokens are will now be the default, instead of bots. You can still specify `user ` before though, it won't break.

### Requirements

Unfortunately, this means the requirements changes.  
They will be the same as [discord-rs](https://github.com/SpaceManiac/discord-rs)'s.
