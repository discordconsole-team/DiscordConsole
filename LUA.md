# LUA Scripting in DiscordConsole
Welcome to LUA: A popular and modifiable scripting language.  
We use LUA in DiscordConsole to allow scripting repetetive tasks and other conveniences.  

To script in LUA is just like any other LUA file.  
When you're done, run them using:  
```run <lua file>```

Easy, huh? Now let's get into the details!

## LUA Specs
We use LUA 5.2 using the [go-lua](https://github.com/Shopify/go-lua) library.  

Added methods:  
`exec(string command): string` - Execute a DiscordConsole command.  
`replace(string original, string search, string replacement): string` - Replaces `search` with `replacement` in `original` and returns.  
`sleep(int seconds)` - Waits for `seconds` seconds.  
