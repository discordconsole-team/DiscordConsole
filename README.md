# DiscordConsole rewrite

DiscordConsole has got some hairy code recently.  
It's time for a rewrite.

Don't worry, the old version will still be supported.

## Changes

First big change:  
**You are now navigating in a UI.**  
That said, the commands system still exists, and it is how it all functions behind the scenes.

Secondly,
you can now pass multiple `-t` or `--token` (long option is also new) parameters at startup, to be able to  
switch between tokens in one instance.  
However, this slows down startup time.

To reduce this problem, you can run it with `--novalidation`, to disable account validation on startup.  
Please keep in mind however, this will let other functions fail silently.

User tokens are now the default, instead of bots. You can still specify `user ` before though, it won't break.

### Requirements

Unfortunately, this means the requirements changes.  
These are the same as [discord-rs](https://github.com/SpaceManiac/discord-rs)'s.
