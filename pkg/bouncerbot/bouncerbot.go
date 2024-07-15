// Package bouncerbot implements the bouncer bot.
package bouncerbot

import (
	"errors"
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/cobaltspeech/log"
	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/kylrth/disco-bouncer/pkg/encrypt"
)

// Bot assigns roles to users once they send the correct code to the bot in a DM. The correct code
// is determined by attempting to decrypt the names in the database using the code as the key. If
// the decryption is successful, the user is given the name as a server nickname and the appropriate
// roles are assigned.
//
// Currently the bot does not support serving more than one guild.
type Bot struct {
	*discordgo.Session
	l log.Logger
	d Decrypter

	gi     *GuildInfo
	giLock sync.RWMutex

	guildInfoCallbacks []func(*GuildInfo)
}

// New creates a new bouncer bot using the provided bot token, backed by the provided UserTable.
func New(l log.Logger, token string, users *db.UserTable) (*Bot, error) {
	return NewWithDecrypter(l, token, TableDecrypter{users})
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

	dg.Identify.Intents = 0

	b.AddHandler(b.handleMemberJoin)
	dg.Identify.Intents |= discordgo.IntentGuildMembers

	b.AddHandler(b.handleMessage)
	dg.Identify.Intents |= discordgo.IntentDirectMessages

	b.AddHandler(b.handleRoleCreate)
	b.AddHandler(b.handleRoleUpdate)
	b.AddHandler(b.handleRoleDelete)
	dg.Identify.Intents |= discordgo.IntentGuilds

	return &b, nil
}

// AddGuildInfoCallback ensures f will be called when the guild info is filled in. A read lock will
// be held while the callbacks are called.
func (b *Bot) AddGuildInfoCallback(f func(*GuildInfo)) {
	b.guildInfoCallbacks = append(b.guildInfoCallbacks, f)
}

func (b *Bot) guildInfoIsNil() bool {
	b.giLock.RLock()
	defer b.giLock.RUnlock()

	return b.gi == nil
}

func (b *Bot) handleMemberJoin(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if b.guildInfoIsNil() {
		b.GetGuildInfo(m.GuildID)
	}

	channel, err := s.UserChannelCreate(m.User.ID)
	if err != nil {
		b.l.Error("msg", "failed to create DM", "error", err)

		return
	}

	b.message(channel.ID, messageWelcome)
	b.l.Debug("msg", "sent welcome DM", "user", m.User.ID, "username", m.User.Username)
}

func (b *Bot) GetGuildInfo(guildID string) {
	roles, err := b.GuildRoles(guildID)
	if err != nil {
		b.l.Error("msg", "failed to get guild roles", "error", err)

		return
	}

	b.giLock.Lock()
	b.gi = GetGuildInfo(b.l, roles, guildID)
	b.giLock.Unlock()

	b.giLock.RLock()
	defer b.giLock.RUnlock()
	for _, cb := range b.guildInfoCallbacks {
		cb(b.gi)
	}
}

func (b *Bot) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != "" {
		// not a DM

		if b.guildInfoIsNil() {
			// let's get the guild info while we're here
			b.GetGuildInfo(m.GuildID)
		}

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
			b.message(m.ChannelID, messageBadKey)

			return
		}
		if errors.Is(err, ErrNotFound) {
			b.l.Info("msg", "key did not decrypt any current user", "key", m.Content, "error", err)
			b.message(m.ChannelID, messageNotFound)

			return
		}

		b.l.Error("msg", "error decrypting with key", "key", m.Content, "error", err)
		b.message(m.ChannelID, messageDecryptionError)

		return
	}

	b.message(m.ChannelID, messageSuccessful)

	err = b.admit(u, m.Author.ID)
	if err != nil {
		if err.Error() != errNick403 {
			b.l.Error("msg", "failed to admit new user", "error", err)
			b.message(m.ChannelID, messageAdmitError)

			return
		}

		b.message(m.ChannelID, messageNickPerm)
	}

	b.l.Info(
		"msg", "admitted new user", "userID", m.Author.ID, "username", m.Author.Username,
		"name", u.Name, "finishYear", u.FinishYear, "isProf", u.Professor, "isTA", u.TA,
		"isSL", u.StudentLeadership, "isAB", u.AlumniBoard,
	)

	// Delete the user now that we've successfully admitted them.
	err = b.d.Delete(u.ID)
	if err != nil {
		b.l.Error("msg", "failed to delete user after admitting", "id", u.ID)
	}
}

const (
	messageWelcome         = "send welcome message"
	messageSuccessful      = "inform about successful decryption"
	messageBadKey          = "reply to bad key"
	messageNotFound        = "reply to unused key"
	messageDecryptionError = "reply about decryption error"
	messageNickPerm        = "reply about nickname permissions"
	messageAdmitError      = "reply about admit error"
	messageOtherError      = "inform about internal error"
)

var messages = map[string][]string{
	messageWelcome: {
		"Welcome to the ACME Discord server! Please send me your unique code here to gain access " +
			"to the rest of the server.",
		"By sending your code, you're allowing BYU to give the server admins your real name and " +
			"the year you finished the senior cohort.",
		"I'll use this information to set which channels you'll be able to see, and to set your " +
			"nickname on the server.",
		"If you're new to Discord, don't send me the code until you've set a password for your " +
			"new account! Otherwise, you'll lose access once you close your browser window and " +
			"your code will not work next time.",
	},
	messageSuccessful: {"I found your info! I'll let you in now. :)"},
	messageBadKey: {
		"Sorry, that key did not work. The key should be 64 hexadecimal characters, sent as " +
			"plain text in a single message by itself.",
		"If you still have trouble, ask for help in the waiting room channel.",
	},
	messageNotFound: {"Sorry, that key did not work. Ask for help in the waiting room channel!"},
	messageDecryptionError: {
		"There was a decryption error with that key. Ask for help in the waiting room channel!",
	},
	messageNickPerm: {
		"Everything worked except I wasn't able to set your nickname because of your high role.",
		"Please set your nickname by sending `/nick FIRST LAST` in one of the channels.",
	},
	messageAdmitError: {
		"There was an error while trying to admit you. Ask for help in the waiting room channel!",
	},
	messageOtherError: {
		"There was an error with a message I tried to send. Complain in the waiting room channel!",
	},
}

func (b *Bot) message(channelID, whatFor string) {
	msgs, ok := messages[whatFor]
	if !ok {
		b.l.Error("msg", "unknown message type", "type", whatFor)
		b.message(channelID, messageOtherError)

		return
	}

	for _, msg := range msgs {
		_, err := b.ChannelMessageSend(channelID, msg)
		if err != nil {
			b.l.Error("msg", "failed to "+whatFor, "error", err)

			break
		}
	}
}

func (b *Bot) admit(u *db.User, dID string) error {
	if b.guildInfoIsNil() {
		return errors.New("guild info not discovered yet")
	}

	var errs []error

	b.giLock.RLock()
	rolesToAdd := b.gi.GetRoleIDsForUser(b.l, u)
	guildID := b.gi.GuildID
	newbieRole := b.gi.NewbieRole
	b.giLock.RUnlock()

	for _, roleID := range rolesToAdd {
		err := b.GuildMemberRoleAdd(guildID, dID, roleID)
		if err != nil {
			errs = append(errs, fmt.Errorf("set role '%s': %w", roleID, err))
		}
	}

	err := b.GuildMemberRoleRemove(guildID, dID, newbieRole)
	if err != nil {
		errs = append(errs, fmt.Errorf("remove newbie role: %w", err))
	}

	err = b.GuildMemberNickname(guildID, dID, u.Name)
	if err != nil {
		errs = append(errs, fmt.Errorf("set nick: %w", err))
	}

	return errors.Join(errs...)
}

const errNick403 = `set nick: HTTP 403 Forbidden, {"message": "Missing Permissions", "code": 50013}`

func (b *Bot) handleRoleCreate(_ *discordgo.Session, m *discordgo.GuildRoleCreate) {
	b.l.Info("msg", "role created; updating guild info", "id", m.Role.ID, "name", m.Role.Name)
	b.GetGuildInfo(m.GuildID)
}

func (b *Bot) handleRoleUpdate(_ *discordgo.Session, m *discordgo.GuildRoleUpdate) {
	b.l.Info("msg", "role updated; updating guild info", "id", m.Role.ID, "name", m.Role.Name)
	b.GetGuildInfo(m.GuildID)
}

func (b *Bot) handleRoleDelete(_ *discordgo.Session, m *discordgo.GuildRoleDelete) {
	b.l.Info("msg", "role deleted; updating guild info", "id", m.RoleID)
	b.GetGuildInfo(m.GuildID)
}

var (
	// ErrUnknownYear is returned by Migrate when the specified cohort year doesn't have a role on
	// the Discord server.
	ErrUnknownYear = errors.New("unknown cohort year")

	// ErrNoUser is returned by Migrate when the specified user isn't found on the server.
	ErrNoUser = errors.New("user not found")
)

// Migrate moves the specified user by name from the pre-ACME role to their new cohort role.
func (b *Bot) Migrate(name, year string) error {
	if b.guildInfoIsNil() {
		b.l.Error(
			"msg", "failed to migrate user due to missing guild info", "name", name, "year", year)

		return errors.New("guild info not discovered yet")
	}

	b.giLock.RLock()
	guildID := b.gi.GuildID
	preACME := b.gi.PreACMERole
	cohort, ok := b.gi.RolesByYear[year]
	b.giLock.RUnlock()

	if !ok {
		b.l.Info("msg", "requested migration to unknown year", "name", name, "year", year)

		return ErrUnknownYear
	}

	found, err := b.GuildMembersSearch(guildID, name, 1)
	if err != nil {
		b.l.Error("msg", "failed to search for Discord user to migrate", "error", err)

		return fmt.Errorf("search for guild user: %w", err)
	}
	if len(found) == 0 {
		b.l.Info("msg", "found no matching Discord user to migrate", "name", name)

		return ErrNoUser
	}
	user := found[0]
	if user.Nick != name && user.User.Username != name {
		b.l.Info(
			"msg", "found inexact match for Discord user migration",
			"name", name, "foundNick", user.Nick, "foundUsername", user.User.Username)

		return ErrNoUser
	}

	b.l.Info(
		"msg", "found matching Discord user for migration",
		"name", name, "year", year, "user", user.User.ID)

	err = b.GuildMemberRoleAdd(guildID, user.User.ID, cohort)
	if err != nil {
		b.l.Error(
			"msg", "failed to add new cohort role", "year", year, "cohort", cohort, "user", user)

		return fmt.Errorf("add new cohort role: %w", err)
	}

	err = b.GuildMemberRoleRemove(guildID, user.User.ID, preACME)
	if err != nil {
		b.l.Error("msg", "failed to remove pre-ACME role", "user", user)

		return fmt.Errorf("remove pre-ACME role: %w", err)
	}

	return nil
}
