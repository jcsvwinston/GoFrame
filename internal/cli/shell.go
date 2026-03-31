package cli

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"
)

func runShell(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("shell", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configPath := fs.String("config", "", "Path to goframe config file")
	command := fs.String("command", "", "Execute one SQL command and exit")
	fs.StringVar(command, "c", "", "Shorthand for --command")
	timeout := fs.Duration("timeout", 10*time.Second, "Per-statement timeout")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if len(fs.Args()) > 0 {
		return fmt.Errorf("shell does not accept positional arguments")
	}
	if *timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}

	_, database, cleanup, err := newDatabase(*configPath)
	if err != nil {
		return err
	}
	defer cleanup()

	sqlDB, err := database.SqlDB()
	if err != nil {
		return fmt.Errorf("open sql handle: %w", err)
	}

	if strings.TrimSpace(*command) != "" {
		return executeSQLScriptWithOutput(sqlDB, *command, *timeout, stdout)
	}

	if !isTerminalReader(stdin) {
		body, err := io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		return executeSQLScriptWithOutput(sqlDB, string(body), *timeout, stdout)
	}

	fmt.Fprintln(stdout, "Entering GoFrame SQL shell. Type 'exit' or 'quit' to leave.")
	scanner := bufio.NewScanner(stdin)
	for {
		fmt.Fprint(stdout, "goframe-sql> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		switch strings.ToLower(line) {
		case "exit", "quit", "\\q":
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		err := executeSQLStatement(ctx, sqlDB, line, stdout)
		cancel()
		if err != nil {
			fmt.Fprintf(stderr, "statement error: %v\n", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read shell input: %w", err)
	}
	return nil
}

func executeSQLScriptWithOutput(sqlDB *sql.DB, script string, timeout time.Duration, out io.Writer) error {
	statements := splitSQLStatements(script)
	if len(statements) == 0 {
		return nil
	}
	for i, stmt := range statements {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		err := executeSQLStatement(ctx, sqlDB, stmt, out)
		cancel()
		if err != nil {
			return fmt.Errorf("statement #%d failed: %w", i+1, err)
		}
	}
	return nil
}
