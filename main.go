package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/smantic/tengen"
)

var list map[string]time.Time = make(map[string]time.Time)

type session struct {
	discordgo.Session
}

type Config struct {
	BotToken string
}

func main() {

	c := Config{}
	tengen.Init(&c, os.Args)

	if c.BotToken == "" {
		log.Fatal("expected non empty bot token")
	}

	session, err := discordgo.New("Bot " + c.BotToken)
	if err != nil {
		log.Fatalf("failed to create discord session: %v\n", err)
	}

	session.AddHandler(listen)

	err = session.Open()
	if err != nil {
		log.Fatalf("failed to open discord session: %v\n", err)
	}
	defer session.Close()
	log.Printf("starting rocbot...")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func listen(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	sesh := session{*s}

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		log.Printf("unable to get channel: %v\n", err)
		return
	}

	content := m.Content
	log.Println(content)

	switch channel.Type {
	case 0:
		if needsReminder(m.Author.Username) {
			sesh.WriteMsg(channel.ID, m.Author.Mention())
			return
		}

		tokens := strings.Split(content, " ")
		if len(tokens) != 2 {
			return
		}

		//log.Printf("%s == shutup", tokens[0])
		if tokens[0] == "shutup" {

			if isBanned(m.Author.Username) {
				sesh.WriteMsg(channel.ID, m.Author.Mention())
				return
			}

			username := strings.Trim(tokens[1], "<@>!")
			sesh.WriteMsg(channel.ID, tokens[1])
			list[username] = time.Now()
		}
	}
}

func needsReminder(username string) bool {
	time, exist := list[username]
	return exist && time.Before(time.Local().AddDate(0, 0, -1))
}

func isBanned(username string) bool {
	_, exist := list[username]
	return exist
}

func (s *session) WriteMsg(cid string, mention string) {

	msg := mention + " SHUT UP BITCH.\n" + "https://www.youtube.com/watch?v=V9O94UTDAJQ"
	_, err := s.ChannelMessageSend(cid, msg)
	if err != nil {
		log.Printf("failed to write msg: %s\n", err.Error())
	}
}
