# LUA Scripting in DiscordConsole
Welcome to LUA: A popular and modifiable scripting language.  
We use LUA in DiscordConsole to allow scripting repetetive tasks and other conveniences.  

To script in LUA is just like any other LUA file.  
When you're done, run them using:  
```
run <lua file>
```

Easy, huh? Now let's get into the details!

## LUA Specs
We use LUA 5.2 using the [go-lua](https://github.com/Shopify/go-lua) library.  

Added methods:  
`exec(string command): string` - Execute a DiscordConsole command.  
`replace(string original, string search, string replacement): string` - Replaces `search` with `replacement` in `original` and returns.  
`sleep(int seconds)` - Waits for `seconds` seconds.  

## Events!
Another feature of DiscordConsole's LUA are events.  
The only event available yet is on message event.  

To register an event, you first create a function.  
This is quite simply what will execute once the event happens.  

Now, call the following function:  
```Lua
registerEvent(string id, string handler)
```
The ID is anything. Anything at all. Well, it gotta be of the type `string`.  
It is only there, so when you reload the script, it doesn't create a new event.  

The handler is the name of your function.  

### Activating the events
Since the events are registered through the `registerEvent` function, you simply need to run them!  
`run your file.lua`  

### Example event
```Lua
function myFunc(event)
	if event.Content == ":>" then
		exec("channel "..event.ChannelID);
		exec("say =)*");
	end
end

registerEvent("my very awesome unique id", "myFunc");
```

### Event params!
Oi there, wait a second... What the heck am I doing? Where did i get that `event.Content` from?  
Explain myself!  

So, I forgot to tell you the function... has a parameter.  
It includes information about the message.  
Here it is:  

| Name         | Description              |
| ------------ | ------------------------ |
| ID           | The message ID.          |
| Content      | The message content.     |
| ChannelID    | The channel ID.          |
| Timestamp    | The timestamp (ANSIC).   |
| AuthorID     | The author's user ID.    |
| AuthorBot    | Either true or false.    |
| AuthorAvatar | The author's avatar key. |
| AuthorName   | The author's name.       |
