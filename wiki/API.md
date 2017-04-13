# DiscordConsole API

TIP: If you're a developer thinking about using the API, please consider Discord's own.  
This API has literally no advantages, and is sort of only here to used by DiscordConsole itself.

## For users

Hi! What's this cool thing? How do you get started?

Easy!  
First, start the API!
```
api_start
```
*(stop it with `api_stop`, or exit the console)*  
You will get a file as output.  
Copy that file path to your clipboard, you'll need it later.

Now, open another console!
```
api_start <file>
```
with *<file>* being the name of the file.

All set up?  
Cool!

Now type
```
broadcast say hi
```

WHAT? Did you see that? All consoles ran `say hi`!  
Amazing, huh?

Yeah that's it, you can now broadcast stuff between consoles and accounts. Pretty cool, 'ey?

# For developers

The file given by `api_start` is a JSON file.  
It contains 2 fields. `Command` and `SentAt`.

`Command` being your command.  
`SentAt` being the current time. Well, it technically doesn't have to be. It just has to be unique.

Ok, that was easy! Yay!
