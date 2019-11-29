package linja

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func addBridge(s *discordgo.Session, m *discordgo.MessageCreate) {
	lex := strings.Split(strings.Split(m.Content, "!")[1], " ")
	alreadyExists := false
	if len(lex) < 2 {
		embed := NewEmbed().
			SetTitle("Please give a bridge name!").
			SetColor(0xff0000)
		s.ChannelMessageSendEmbed(m.ChannelID, embed.MessageEmbed)
		return
	}
	for _, channel := range bridgedChans {
		if channel.ChanID == m.ChannelID {
			alreadyExists = true
		}
	}
	if alreadyExists {
		embed := NewEmbed().
			SetTitle("Error!").
			SetDescription("This room is already bridged.").
			SetColor(0xff0000)
		s.ChannelMessageSendEmbed(m.ChannelID, embed.MessageEmbed)
		return
	} else {
		wh, err := s.WebhookCreate(m.ChannelID, "Bridge", "")
		if err != nil {
			embed := NewEmbed().
				SetTitle("Error!").
				SetDescription("I couldn't create a webhook. Please make sure I have the permissions.").
				SetColor(0xff0000)
			s.ChannelMessageSendEmbed(m.ChannelID, embed.MessageEmbed)
			return
		}
		bridgedChans = append(bridgedChans, bridgedChan{
			ChanID:    m.ChannelID,
			GuildID:   m.GuildID,
			Bridge:    lex[1],
			WebhookID: wh.ID,
			Token:     wh.Token,
		})
		embed := NewEmbed().
			SetTitle("Successful!").
			SetDescription("You are now bridged to " + lex[1]).
			SetColor(0x00ff00)
		s.ChannelMessageSendEmbed(m.ChannelID, embed.MessageEmbed)
		return
	}
}

func remove(s []bridgedChan, i int) []bridgedChan {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func rmBridge(s *discordgo.Session, m *discordgo.MessageCreate) {
	alreadyExists := false
	var bridge bridgedChan
	var bridgeIndex int
	for indx, channel := range bridgedChans {
		if channel.ChanID == m.ChannelID {
			alreadyExists = true
			bridge = channel
			bridgeIndex = indx
		}
	}
	if !alreadyExists {
		embed := NewEmbed().
			SetTitle("Error!").
			SetDescription("This room isn't bridged.").
			SetColor(0xff0000)
		s.ChannelMessageSendEmbed(m.ChannelID, embed.MessageEmbed)
		return
	} else {
		bridgedChans = remove(bridgedChans, bridgeIndex)
		embed := NewEmbed().
			SetTitle("Successful!").
			SetDescription("You are now no longer bridged to " + bridge.Bridge).
			SetColor(0x00ff00)
		s.ChannelMessageSendEmbed(m.ChannelID, embed.MessageEmbed)
		return
	}
}
