// Package bouncerbot implements the bouncer bot.
package bouncerbot

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/cobaltspeech/log"
	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/kylrth/disco-bouncer/pkg/encrypt"
)

// Bot assigns roles to users once they send the correct code to the bot in a DM. The correct code
// is determined by attempting to decrypt the names in the database using the code as the key. If
// the decryption is successful, the user is given the name as a server nickname and the appropriate
// roles are assigned.
type Bot struct {
	*discordgo.Session
	l log.Logger
	d Decrypter

	// If left blank, this info is set when the first user joins the server. Currently the bot does
	// not support serving more than one guild.
	Guild *GuildInfo

	guildInfoCallbacks []func(*GuildInfo)
}

// New creates a new bouncer bot using the provided bot token, backed by the provided UserTable.
func New(l log.Logger, token string, users *db.UserTable) (*Bot, error) {
	return NewWithDecrypter(l, token, tableDecrypter{users})
}

// NewWithDecrypter sets up the bot using the provided Decrypter, instead of the default Decrypter
// backed with a db.UserTable.
func NewWithDecrypter(l log.Logger, token string, d Decrypter) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	b := Bot{
		Session: dg,
		l:       l,
		d:       d,
	}

	if err != nil {
		return &b, err
	}

	b.AddHandler(b.memberJoin)
	b.AddHandler(b.userDM)

	return &b, nil
}

// AddGuildInfoCallback ensures f will be called when the b.Guild is filled in.
func (b *Bot) AddGuildInfoCallback(f func(*GuildInfo)) {
	b.guildInfoCallbacks = append(b.guildInfoCallbacks, f)
}

var welcomeMessages = []string{
	"Welcome to the ACME Discord server! Please send me your unique code here to gain access to " +
		"the rest of the server.",
	"By sending your code, you're allowing BYU to give the server admins your real name and the " +
		"year you finished the senior cohort.",
	"I'll use this information to set which channels you'll be able to see, and to set your " +
		"nickname on the server.",
}

func (b *Bot) memberJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if b.Guild == nil {
		roles, err := s.GuildRoles(m.GuildID)
		if err != nil {
			b.l.Error("msg", "failed to get guild info", "error", err)
		}

		b.Guild = GetGuildInfo(b.l, roles, m.GuildID)

		for _, cb := range b.guildInfoCallbacks {
			cb(b.Guild)
		}
	}

	channel, err := s.UserChannelCreate(m.User.ID)
	if err != nil {
		b.l.Error("msg", "failed to create DM", "error", err)

		return
	}

	for _, msg := range welcomeMessages {
		b.message(channel.ID, "send welcome message", msg)
	}

	b.l.Debug("msg", "sent welcome DM", "user", m.User.ID, "username", m.User.Username)
}

func (b *Bot) userDM(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != "" {
		// not a DM
		return
	}
	if m.Author.ID == s.State.User.ID {
		// a message from us
		return
	}

	u, err := b.d.Decrypt(m.Content)
	if err != nil {
		if errors.As(err, &encrypt.ErrBadKey{}) {
			b.l.Info("msg", "DM did not provide acceptable key", "key", m.Content, "error", err)

			b.message(m.ChannelID, "reply to bad key",
				"Sorry, that key did not work. Send just the key in plain text, nothing else.",
			)

			return
		}
		if errors.Is(err, ErrNotFound) {
			b.l.Info("msg", "key did not decrypt any current user", "key", m.Content, "error", err)

			b.message(m.ChannelID, "reply to unused key",
				"Sorry, that key did not work. Reach out to the admins for help!",
			)

			return
		}

		b.l.Error("msg", "error decrypting with key", "key", m.Content, "error", err)

		b.message(m.ChannelID, "reply about decryption error",
			"There was an error on my end (`"+err.Error()+"`). Reach out to the admins for help!",
		)
	}

	b.message(m.ChannelID, "inform about successful decryption",
		"I found your info! I'll let you in now. :)",
	)

	err = b.admit(u, m.Author.ID)
	if err != nil {
		b.l.Error("msg", "failed to admit new user", "error", err)
		b.message(m.ChannelID, "reply about admit error",
			"There was an error on my end (`"+err.Error()+"`). Reach out to the admins for help!",
		)
	}

	b.l.Info("msg", "admitted new user", "userID", m.Author.ID, "username", m.Author.Username)
}

func (b *Bot) message(channelID, whatFor, msg string) {
	_, err := b.ChannelMessageSend(channelID, msg)
	if err != nil {
		b.l.Error("msg", "failed to "+whatFor, "error", err)
	}
}

func (b *Bot) admit(u *db.User, dID string) error {
	if b.Guild == nil {
		return errors.New("guild info not discovered yet")
	}

	err := b.GuildMemberNickname(b.Guild.GuildID, dID, u.Name)
	if err != nil {
		return fmt.Errorf("set nick: %w", err)
	}

	rolesToAdd := b.Guild.GetRoleIDsForUser(b.l, u)

	for _, roleID := range rolesToAdd {
		err = b.GuildMemberRoleAdd(b.Guild.GuildID, dID, roleID)
		if err != nil {
			return fmt.Errorf("set role: %w", err)
		}
	}

	err = b.GuildMemberRoleRemove(b.Guild.GuildID, dID, b.Guild.NewbieRole)
	if err != nil {
		return fmt.Errorf("remove newbie role: %w", err)
	}

	return nil
}
