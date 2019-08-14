# These are a few scripts for your convenience. To run them, use Just:
# https://github.com/casey/just

# -*- mode: make-gmake-mode; -*-

debug:
    go install --race

release:
    go install

fix-dgo:
    #!/usr/bin/env sh
    go get github.com/bwmarrin/discordgo
    cd "$GOPATH/src/github.com/bwmarrin/discordgo"
    git checkout develop
    go install

cross-compile:
    #!/usr/bin/env sh

    cd "$GOPATH/bin"
    ./CrossCompile.sh DiscordConsole discordconsole-team
