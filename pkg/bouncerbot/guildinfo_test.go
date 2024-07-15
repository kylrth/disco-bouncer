package bouncerbot_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/cobaltspeech/log/pkg/testinglog"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/kylrth/disco-bouncer/pkg/bouncerbot"
)

var (
	adminRole = discordgo.Role{
		ID:   "a",
		Name: "Admin",
	}
	cohort2016 = discordgo.Role{
		ID:   "b",
		Name: "2016 𝛽",
	}
	cohort2019 = discordgo.Role{
		ID:   "c",
		Name: "2019 𝜀",
	}
	cohort2022 = discordgo.Role{
		ID:   "d",
		Name: "2022 𝜃",
	}
	profRole = discordgo.Role{
		ID:   "e",
		Name: "professor",
	}
	taRole = discordgo.Role{
		ID:   "f",
		Name: "TA",
	}
	slRole = discordgo.Role{
		ID:   "g",
		Name: "student leadership",
	}
	boardRole = discordgo.Role{
		ID:   "h",
		Name: "alumni board",
	}
	newbieRole = discordgo.Role{
		ID:   "i",
		Name: "newbie",
	}
	preACMERole = discordgo.Role{
		ID:   "k",
		Name: "pre-core ACME",
	}
	trickyRole = discordgo.Role{
		ID:   "j",
		Name: "teehee (l00l)",
	}
)

func TestGetGuildInfo(t *testing.T) {
	t.Parallel()

	const guildID = "guildddd"

	type testCase struct {
		roles []*discordgo.Role
		want  bouncerbot.GuildInfo
	}
	tests := map[string]testCase{
		"empty": {nil, bouncerbot.GuildInfo{GuildID: guildID}},
		"everything": {
			[]*discordgo.Role{
				&adminRole, &cohort2016, &cohort2019, &cohort2022, &profRole, &taRole, &slRole,
				&boardRole, &newbieRole, &preACMERole,
			},
			bouncerbot.GuildInfo{
				GuildID:               guildID,
				ProfessorRole:         profRole.ID,
				TARole:                taRole.ID,
				StudentLeadershipRole: slRole.ID,
				AlumniBoardRole:       boardRole.ID,
				NewbieRole:            newbieRole.ID,
				PreACMERole:           preACMERole.ID,
				RolesByYear: map[string]string{
					"2016": cohort2016.ID, "2019": cohort2019.ID, "2022": cohort2022.ID,
				},
			},
		},
		"missingSome": {
			[]*discordgo.Role{
				&adminRole, &cohort2016, &cohort2019, &cohort2022, &taRole,
				&boardRole, &preACMERole,
			},
			bouncerbot.GuildInfo{
				GuildID:         guildID,
				TARole:          taRole.ID,
				AlumniBoardRole: boardRole.ID,
				PreACMERole:     preACMERole.ID,
				RolesByYear: map[string]string{
					"2016": cohort2016.ID, "2019": cohort2019.ID, "2022": cohort2022.ID,
				},
			},
		},
		"tricky": {
			[]*discordgo.Role{
				&adminRole, &trickyRole, &profRole, &taRole, &slRole,
				&boardRole, &newbieRole, &preACMERole,
			},
			bouncerbot.GuildInfo{
				GuildID:               guildID,
				ProfessorRole:         profRole.ID,
				TARole:                taRole.ID,
				StudentLeadershipRole: slRole.ID,
				AlumniBoardRole:       boardRole.ID,
				NewbieRole:            newbieRole.ID,
				PreACMERole:           preACMERole.ID,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			l := testinglog.NewConvenientLogger(t)
			defer l.Done()

			info := bouncerbot.GetGuildInfo(l, tc.roles, guildID)

			if diff := cmp.Diff(&tc.want, info, cmpopts.EquateEmpty()); diff != "" {
				t.Error("unexpected output (-want +got):\n" + diff)
			}
		})
	}
}

func TestGuildInfo_GetRoleIDsForUser(t *testing.T) {
	t.Parallel()

	info := bouncerbot.GuildInfo{
		GuildID:               "guildy",
		ProfessorRole:         profRole.ID,
		TARole:                taRole.ID,
		StudentLeadershipRole: slRole.ID,
		AlumniBoardRole:       boardRole.ID,
		NewbieRole:            newbieRole.ID,
		PreACMERole:           preACMERole.ID,
		RolesByYear: map[string]string{
			"2016": cohort2016.ID, "2019": cohort2019.ID, "2022": cohort2022.ID,
		},
	}

	type testCase struct {
		in   *db.User
		want []string
	}
	tests := map[string]testCase{
		"2022": {
			&db.User{FinishYear: "2022", StudentLeadership: true},
			[]string{cohort2022.ID, slRole.ID},
		},
		"2021": {
			&db.User{FinishYear: "2021"},
			nil,
		},
		"TA_alum": {
			&db.User{FinishYear: "2019", TA: true, AlumniBoard: true},
			[]string{cohort2019.ID, taRole.ID, boardRole.ID},
		},
		"prof": {
			&db.User{Professor: true},
			[]string{profRole.ID},
		},
		"pre-core": {
			&db.User{},
			[]string{preACMERole.ID},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			l := testinglog.NewConvenientLogger(t)
			defer l.Done()

			got := info.GetRoleIDsForUser(l, tc.in)

			if diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Error("unexpected output (-want +got):\n" + diff)
			}
		})
	}
}
