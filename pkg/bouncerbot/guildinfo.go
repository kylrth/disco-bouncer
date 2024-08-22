package bouncerbot

import (
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
	"github.com/cobaltspeech/log"
	"github.com/kylrth/disco-bouncer/internal/db"
)

const (
	professorRole         = "professor"
	taRole                = "TA"
	studentLeadershipRole = "student leadership"
	alumniBoardRole       = "alumni board"
	newbieRole            = "newbie"
	preCoreRole           = "pre-core ACME"
)

// GuildInfo contains IDs necessary for the bot to interact with roles and users in the guild.
type GuildInfo struct {
	GuildID string `json:"guild_id"`

	ProfessorRole         string `json:"professor_role"`
	TARole                string `json:"ta_role"`
	StudentLeadershipRole string `json:"student_leadership_role"`
	AlumniBoardRole       string `json:"alumni_board_role"`
	NewbieRole            string `json:"newbie_role"`
	PreCoreRole           string `json:"preacme_role"` // keep old JSON key

	RolesByYear map[string]string `json:"roles_by_year"`
}

// GetGuildInfo collects the guild information using the discordgo session.
func GetGuildInfo(l log.Logger, roles []*discordgo.Role, guildID string) *GuildInfo {
	var out GuildInfo
	out.GuildID = guildID

	out.RolesByYear = make(map[string]string)
	for _, role := range roles {
		if year := getYearIfPresent(role.Name); year != "" {
			out.RolesByYear[year] = role.ID

			continue
		}

		switch role.Name {
		case professorRole:
			out.ProfessorRole = role.ID
		case taRole:
			out.TARole = role.ID
		case studentLeadershipRole:
			out.StudentLeadershipRole = role.ID
		case alumniBoardRole:
			out.AlumniBoardRole = role.ID
		case newbieRole:
			out.NewbieRole = role.ID
		case preCoreRole:
			out.PreCoreRole = role.ID
		}
	}

	checkRoleFilled(l, out.ProfessorRole, professorRole)
	checkRoleFilled(l, out.TARole, taRole)
	checkRoleFilled(l, out.StudentLeadershipRole, studentLeadershipRole)
	checkRoleFilled(l, out.AlumniBoardRole, alumniBoardRole)
	checkRoleFilled(l, out.NewbieRole, newbieRole)
	checkRoleFilled(l, out.PreCoreRole, preCoreRole)

	l.Debug("msg", "Collected guild info.", "RolesByYear", out.RolesByYear)

	return &out
}

// getYearIfPresent returns "" if the string doesn't start with a year, otherwise it returns up to
// the first space character " ".
func getYearIfPresent(s string) string {
	if len(s) < 4 {
		return ""
	}

	// first 4 runes must be a year
	for _, c := range s[:4] {
		if !unicode.IsDigit(c) {
			return ""
		}
	}

	yearEnd := strings.Index(s, " ")
	if yearEnd == -1 {
		return s
	}

	return s[:yearEnd]
}

func checkRoleFilled(l log.Logger, field, name string) {
	if field == "" {
		l.Error("msg", "role info not found", "role", name)
	}
}

// GetRoleIDsForUser returns the role IDs that the user should be given.
func (i *GuildInfo) GetRoleIDsForUser(l log.Logger, u *db.User) []string {
	roleIDs := []string{}
	if u.FinishYear != "" {
		if role, ok := i.RolesByYear[u.FinishYear]; ok {
			roleIDs = append(roleIDs, role)
		} else {
			l.Info("msg", "no role for finish year", "finishYear", u.FinishYear)
		}
	} else if !u.Professor {
		roleIDs = append(roleIDs, i.PreCoreRole)
	}

	if u.Professor {
		roleIDs = append(roleIDs, i.ProfessorRole)
	}
	if u.TA {
		roleIDs = append(roleIDs, i.TARole)
	}
	if u.StudentLeadership {
		roleIDs = append(roleIDs, i.StudentLeadershipRole)
	}
	if u.AlumniBoard {
		roleIDs = append(roleIDs, i.AlumniBoardRole)
	}

	return roleIDs
}
