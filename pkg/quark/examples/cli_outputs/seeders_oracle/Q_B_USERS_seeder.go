package seeders

import (
	"context"
	"github.com/jcsvwinston/GoFrame/pkg/quark"
)

func SeedQBUsers(ctx context.Context, client *quark.Client) error {
	// Write your seeding logic here
	// Example:
	// return quark.For[User](ctx, client).Create(&User{Name: "Admin"})
	return nil
}
