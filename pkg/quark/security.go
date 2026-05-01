package quark

import (
	"fmt"
	"regexp"
	"strings"
)

// SQLGuard provides security validations for SQL queries.
// It prevents SQL injection by validating identifiers and enforcing safe practices.
type SQLGuard struct {
	identifierPattern *regexp.Regexp
	reservedKeywords  map[string]bool
	maxIdentifierLen  int
}

// reservedSQLKeywords contains SQL keywords that should not be used as identifiers.
// This is not exhaustive but covers the most common and dangerous ones.
var reservedSQLKeywords = map[string]bool{
	"SELECT": true, "INSERT": true, "UPDATE": true, "DELETE": true,
	"DROP": true, "CREATE": true, "ALTER": true, "TRUNCATE": true,
	"EXEC": true, "EXECUTE": true, "UNION": true, "UNION ALL": true,
	"OR": true, "AND": true, "WHERE": true, "FROM": true, "JOIN": true,
	"LEFT": true, "RIGHT": true, "INNER": true, "OUTER": true,
	"ORDER": true, "GROUP": true, "HAVING": true, "LIMIT": true,
	"OFFSET": true, "VALUES": true, "SET": true, "INTO": true,
	"TABLE": true, "DATABASE": true, "SCHEMA": true, "INDEX": true,
	"VIEW": true, "TRIGGER": true, "PROCEDURE": true, "FUNCTION": true,
}

// identifierRegex matches valid SQL identifiers.
// Must start with letter or underscore, followed by letters, numbers, or underscores.
var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// NewSQLGuard creates a new SQLGuard with default settings.
func NewSQLGuard() *SQLGuard {
	return &SQLGuard{
		identifierPattern: identifierRegex,
		reservedKeywords:  reservedSQLKeywords,
		maxIdentifierLen:  64,
	}
}

// ValidateIdentifier checks if a table or column identifier is safe to use.
// Returns error if the identifier:
// - Is empty or too long (>64 chars)
// - Contains invalid characters
// - Is a reserved SQL keyword
func (g *SQLGuard) ValidateIdentifier(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("%w: identifier is empty", ErrInvalidIdentifier)
	}

	if len(name) > g.maxIdentifierLen {
		return fmt.Errorf("%w: identifier %q exceeds maximum length of %d characters",
			ErrInvalidIdentifier, name, g.maxIdentifierLen)
	}

	if !g.identifierPattern.MatchString(name) {
		return fmt.Errorf("%w: identifier %q contains invalid characters. Only letters, numbers, and underscores allowed",
			ErrInvalidIdentifier, name)
	}

	upper := strings.ToUpper(name)
	if g.reservedKeywords[upper] {
		return fmt.Errorf("%w: identifier %q is a reserved SQL keyword",
			ErrInvalidIdentifier, name)
	}

	return nil
}

// ValidateIdentifiers checks multiple identifiers at once.
func (g *SQLGuard) ValidateIdentifiers(names ...string) error {
	for _, name := range names {
		if err := g.ValidateIdentifier(name); err != nil {
			return err
		}
	}
	return nil
}

// QuoteIdentifier validates and quotes an identifier using the dialect.
func (g *SQLGuard) QuoteIdentifier(dialect Dialect, name string) (string, error) {
	if err := g.ValidateIdentifier(name); err != nil {
		return "", err
	}
	return dialect.Quote(name), nil
}

// ValidateOperator checks if an operator is in the allowed whitelist.
func (g *SQLGuard) ValidateOperator(op string) error {
	allowedOperators := map[string]bool{
		"=": true, "!=": true, "<>": true, "<": true, ">": true,
		"<=": true, ">=": true,
		"LIKE": true, "NOT LIKE": true,
		"IN": true, "NOT IN": true,
		"IS": true, "IS NOT": true,
		"BETWEEN": true, "NOT BETWEEN": true,
	}

	upper := strings.ToUpper(strings.TrimSpace(op))
	if !allowedOperators[upper] {
		return fmt.Errorf("%w: operator %q is not allowed", ErrInvalidQuery, op)
	}
	return nil
}

// HasPlaceholders checks if a query string contains parameter placeholders.
// This is used to enforce that raw queries use parameterized statements.
func HasPlaceholders(query string) bool {
	patterns := []string{
		`\?`,    // MySQL, SQLite: ?
		`\$\d+`, // PostgreSQL: $1, $2
		`@p\d+`, // MSSQL: @p1, @p2
		`:\d+`,  // Oracle: :1, :2
		`:\w+`,  // Oracle named: :name
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, query)
		if matched {
			return true
		}
	}

	return false
}

// ValidateRawQuery performs basic validation on a raw SQL query.
// It checks for placeholders and suspicious patterns.
func (g *SQLGuard) ValidateRawQuery(query string, requirePlaceholders bool) error {
	if requirePlaceholders && !HasPlaceholders(query) {
		return fmt.Errorf("%w: raw queries must use placeholders (?, $1, @p1, :1, etc.)",
			ErrInvalidQuery)
	}

	// Additional checks for obvious injection attempts
	suspiciousPatterns := []string{
		`;\s*DROP\s`,
		`;\s*DELETE\s`,
		`;\s*UPDATE\s+\w+\s+SET\s+\w+\s*=`,
		`UNION\s+SELECT`,
		`OR\s+1\s*=\s*1`,
		`OR\s+'\s*1\s*'\s*=\s*'\s*1`,
	}

	upper := strings.ToUpper(query)
	for _, pattern := range suspiciousPatterns {
		matched, _ := regexp.MatchString(pattern, upper)
		if matched {
			return fmt.Errorf("%w: query contains suspicious patterns that may indicate SQL injection",
				ErrInvalidQuery)
		}
	}

	return nil
}
