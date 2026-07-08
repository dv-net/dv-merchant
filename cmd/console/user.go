package console

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"syscall"

	"github.com/dv-net/dv-merchant/internal/cache"
	"github.com/dv-net/dv-merchant/internal/service"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_users"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

func prepareUsersCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "set-email",
			Usage: "change user email",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "old_email",
					Aliases:  []string{"old"},
					Usage:    "user email",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "new_email",
					Aliases:  []string{"new"},
					Usage:    "new email",
					Required: true,
				},
			},
			Action: func(ctx context.Context, c *cli.Command) error {
				conf, err := loadConfig(c.Args().Slice(), c.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				st, err := storage.InitStore(ctx, conf)
				if err != nil {
					return fmt.Errorf("storage init: %w", err)
				}

				const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
				re := regexp.MustCompile(emailRegex)

				email := c.String("old_email")
				if !re.MatchString(email) {
					return fmt.Errorf("invalid email address: %s", email)
				}

				newEmail := c.String("new_email")
				if !re.MatchString(newEmail) {
					return fmt.Errorf("invalid email address: %s", newEmail)
				}

				usr, err := st.Users().GetByEmail(ctx, email)
				if err != nil {
					return fmt.Errorf("failed to get user by email: %w", err)
				}

				return st.Users().SetEmail(ctx, repo_users.SetEmailParams{
					ID:    usr.ID,
					Email: newEmail,
				})
			},
		}, // change email
		{
			Name:  "reset-password",
			Usage: "reset user password",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "email",
					Usage: "Email user",
				},
				&cli.StringFlag{
					Name:  "password",
					Usage: "New password",
				},
			},
			Action: resetPassword,
		},
	}
}

func resetPassword(ctx context.Context, c *cli.Command) error {
	conf, err := loadConfig(c.Args().Slice(), c.StringSlice("configs"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	st, err := storage.InitStore(ctx, conf)
	if err != nil {
		return fmt.Errorf("storage init: %w", err)
	}
	l := logger.New("local", conf.Log)

	ca := cache.InitCache()
	services, err := service.NewServices(ctx, conf, st, ca, l, "local", "local")
	if err != nil {
		return fmt.Errorf("new services: %w", err)
	}

	email := c.String("email")
	password := c.String("password")

	if email == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Fprint(os.Stdout, "Enter your email: ")
		email, _ = reader.ReadString('\n')
		email = strings.TrimSpace(email)
	}

	usr, err := st.Users().GetByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user by email %w", err)
	}

	if password == "" {
		fmt.Fprint(os.Stdout, "Enter your new password: ")
		bytePassword, _ := term.ReadPassword(syscall.Stdin)
		password = string(bytePassword)
	}

	err = services.UserCredentialsService.ChangeUserPassword(ctx, usr.ID, password)
	if err != nil {
		return fmt.Errorf("change user password: %w", err)
	}

	fmt.Println("Password successfully reset")
	return nil
}
