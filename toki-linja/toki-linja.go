package linja

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/badger"
	"gopkg.in/ini.v1"
)

var cfg *ini.File
var db *badger.DB
var bridgedChans []bridgedChan

type bridgedChan struct {
	ChanID    string
	WebhookID string
	GuildID   string
	Bridge    string
	Token     string
}
type cmd func(*discordgo.Session, *discordgo.MessageCreate)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.WebhookID != "" {
		return
	}
	for _, channel := range bridgedChans {
		if m.WebhookID == channel.WebhookID {
			return
		}
	}
	if strings.HasPrefix(m.Content, "tl!") && m.Author.ID == cfg.Section("Bot").Key("operator").String() {
		cmds := map[string]cmd{
			"add": addBridge,
			"rm":  rmBridge,
		}
		lex := strings.Split(strings.Split(m.Content, "!")[1], " ")
		if val, ok := cmds[lex[0]]; ok {
			go val(s, m)
		}
	}
	var brdg string
	for _, channel := range bridgedChans {
		if m.ChannelID == channel.ChanID {
			brdg = channel.Bridge
		}
	}
	if brdg == "" {
		return
	}
	for _, channel := range bridgedChans {
		if channel.Bridge == brdg && channel.ChanID != m.ChannelID {
			name := ""
			guild, err := s.State.Guild(m.GuildID)
			if err != nil {
				guild, err = s.Guild(m.GuildID)
				if err != nil {
					name = "ma"
				}
			} else {
				name = guild.Name
			}
			params := discordgo.WebhookParams{
				Username:  m.Author.Username + " lon " + name,
				AvatarURL: m.Author.AvatarURL(""),
				Content:   m.ContentWithMentionsReplaced(),
				Embeds:    m.Embeds,
			}
			s.WebhookExecute(
				channel.WebhookID,
				channel.Token,
				false,
				&params,
			)
		}
	}
	logMessage(m.Message, brdg)
}

func logMessage(m *discordgo.Message, bridge string) {
	data, _ := json.Marshal(m)
	f, err := os.OpenFile("./storage/"+bridge+".log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(string(data) + "\n"); err != nil {
		log.Println(err)
	}
}

func saveBridges() {
	data, _ := json.Marshal(bridgedChans)
	db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("bridges"), data)
		return err
	})
}

func loadBridges() {
	db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("bridges"))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			json.Unmarshal(val, &bridgedChans)
			return nil
		})
		return nil
	})
}

func Main() {
	var err error
	cfg, err = ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Failed to load config.ini")
		os.Exit(1)
	}
	db, err = badger.Open(badger.DefaultOptions("./storage/db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	loadBridges()
	defer saveBridges()

	discord, err := discordgo.New("Bot " + cfg.Section("Bot").Key("token").String())
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
	}

	discord.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening connection: ", err)
		return
	}

	fmt.Println("ilo Koje is now running.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}
