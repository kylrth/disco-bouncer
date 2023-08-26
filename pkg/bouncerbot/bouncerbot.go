// Package bouncerbot implements the bouncer bot.
package bouncerbot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cobaltspeech/log"
)

// New creates a new bouncer bot using the provided bot token. The bot assigns roles to users once
// they send the correct code to the bot in a DM. The correct code is determined by attempting to
// decrypt the names in the database using the code as the key. If the decryption is successful, the
// user is given the name as a server nickname and the appropriate roles are assigned.
func New(l log.Logger, token string) (*discordgo.Session, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return dg, err
	}

	dg.AddHandler(memberJoin(l))
	dg.AddHandler(userDM(l))

	return dg, nil
}

var welcomeMessages = []string{
	"Welcome to the ACME Discord server! Please enter your unique code in this DM to gain access " +
		"to the rest of the server.",
}

func memberJoin(l log.Logger) func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		channel, err := s.UserChannelCreate(m.User.ID)
		if err != nil {
			l.Error("msg", "failed to create DM", "error", err)

			return
		}

		for _, msg := range welcomeMessages {
			_, err = s.ChannelMessageSend(channel.ID, msg)
			if err != nil {
				l.Error("msg", "failed to send DM", "error", err)

				return
			}
		}

		l.Debug("msg", "sent welcome DM", "user", m.User.ID, "username", m.User.Username)
	}
}

func userDM(l log.Logger) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.GuildID != "" {
			// not a DM
			return
		}
		if m.Author.ID == s.State.User.ID {
			// a message from us
			return
		}

		_, err := s.ChannelMessageSend(m.ChannelID, "Hello! Nice DM you got there")
		if err != nil {
			l.Error("msg", "failed to reply to DM", "error", err)
		}

		l.Debug("msg", "replied to DM", "userID", m.Author.ID, "username", m.Author.Username)
	}
}
