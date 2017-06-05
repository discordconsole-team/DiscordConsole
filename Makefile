default:
	go install
race:
	go install --race
dgo:
	go get github.com/bwmarrin/discordgo
	cd $(GOPATH)/src/github.com/bwmarrin/discordgo; \
		git checkout develop; \
		go install

build:
	# This requires a script only I have.
	# This will not work for you.
	cd $(GOPATH)/bin; \
		./Cross\ Compile.sh DiscordConsole discordconsole-team
