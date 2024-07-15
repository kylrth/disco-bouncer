package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cobaltspeech/log"
	"github.com/kylrth/disco-bouncer/internal/db"
	"github.com/kylrth/disco-bouncer/pkg/client"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload user information",
	Long: `Upload user information to the server, encrypting the name field and then printing the keys.

The user info is accepted in CSV format on stdin. The first line may optionally be exactly the
header below:

	name,finish_year,professor,ta,student_leadership,alumni_board
	John Doe,2016,0,1,0,1
	...

If no data is provided on stdin, the information will be prompted for in the terminal.

The finish year may be set to -1 if the user is pre-ACME.

As users are uploaded, the name and key will be printed to stdout like this:

	id,name,key
	1,John Doe,1234567890abcdef1234567890abcdef

The key is a hex string that should be provided to the user. They will be able to use it to gain
access to the Discord server.
`,
	Args: cobra.NoArgs,
	Run: withLAndC(func(l log.Logger, c *client.Client, _ []string) error {
		return upload(l, c)
	}),
}

func upload(l log.Logger, c *client.Client) error {
	ch := make(chan *db.User)

	go func() {
		err := getInput(ch)
		if err != nil {
			l.Error("msg", "error getting input", "error", err)
		}
	}()

	fmt.Println("id,name,key")
	for u := range ch {
		plainName := u.Name

		id, key, err := c.Users.Upload(context.Background(), u)
		if err != nil {
			return fmt.Errorf("upload user: %w", err)
		}

		fmt.Printf("%d,%s,%s\n", id, plainName, key)
	}

	return nil
}

func getInput(c chan<- *db.User) error {
	// check if info is on stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return getInputFromStdin(c)
	}

	// get info by prompting on the terminal
	promptUser(c)

	return nil
}

func getInputFromStdin(c chan<- *db.User) error {
	r := csv.NewReader(os.Stdin)
	r.ReuseRecord = true
	r.FieldsPerRecord = 6

	defer close(c)

	lineNum := 0
	for {
		line, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}
		lineNum++

		if lineNum == 1 && isUploadHeader(line) {
			continue
		}

		u, err := parseUser(line)
		if err != nil {
			return err
		}
		c <- u
	}
}

func promptUser(c chan<- *db.User) {
	scanner := bufio.NewScanner(os.Stdin)
	defer close(c)

	for {
		var u db.User

		fmt.Fprint(os.Stderr, "Name (leave empty to quit): ")
		if scanner.Scan() {
			u.Name = scanner.Text()
			if u.Name == "" {
				return
			}
		} else {
			return
		}

		fmt.Fprint(os.Stderr, "Finish year (leave empty for pre-ACME or prof): ")
		if scanner.Scan() {
			u.FinishYear = scanner.Text()
		}

		fmt.Fprint(os.Stderr, "Professor? (y/N): ")
		if scanner.Scan() {
			u.Professor = strings.EqualFold(scanner.Text(), "y")
		} else {
			fmt.Fprintln(os.Stderr, "unexpected end of input")
		}

		fmt.Fprint(os.Stderr, "TA? (y/N): ")
		if scanner.Scan() {
			u.TA = strings.EqualFold(scanner.Text(), "y")
		} else {
			fmt.Fprintln(os.Stderr, "unexpected end of input")
		}

		fmt.Fprint(os.Stderr, "Student leadership? (y/N): ")
		if scanner.Scan() {
			u.StudentLeadership = strings.EqualFold(scanner.Text(), "y")
		} else {
			fmt.Fprintln(os.Stderr, "unexpected end of input")
		}

		fmt.Fprint(os.Stderr, "Alumni board? (y/N): ")
		if scanner.Scan() {
			u.AlumniBoard = strings.EqualFold(scanner.Text(), "y")
		} else {
			fmt.Fprintln(os.Stderr, "unexpected end of input")
		}

		c <- &u
	}
}

func isUploadHeader(line []string) bool {
	return line[0] == "name" &&
		line[1] == "finish_year" &&
		line[2] == "professor" &&
		line[3] == "ta" &&
		line[4] == "student_leadership" &&
		line[5] == "alumni_board"
}

func parseUser(line []string) (*db.User, error) {
	var u db.User
	u.Name = line[0]
	u.FinishYear = line[1]

	var err error
	u.Professor, err = parseBool(line[2])
	if err != nil {
		return &u, err
	}
	u.TA, err = parseBool(line[3])
	if err != nil {
		return &u, err
	}
	u.StudentLeadership, err = parseBool(line[4])
	if err != nil {
		return &u, err
	}
	u.AlumniBoard, err = parseBool(line[5])
	if err != nil {
		return &u, err
	}

	return &u, nil
}

func parseBool(s string) (bool, error) {
	switch s {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("invalid bool: %s", s)
	}
}
