default:
	go install
dgo:
	go get github.com/bwmarrin/discordgo
	cd $(GOPATH)/src/github.com/bwmarrin/discordgo; \
		git checkout develop; \
		go install
