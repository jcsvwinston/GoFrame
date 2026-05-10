// Command admin-server is the standalone observability admin for Nucleus.
//
// In Phase 1 this is a placeholder that exits immediately after printing a
// banner. Phase 4 of the refactor plan replaces this main with the real
// AgentService + ControlService server.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
)

const skeletonBanner = `nucleus admin-server (Phase 1 skeleton).

This binary is intentionally a no-op until Phase 4. It exists today so the
monorepo skeleton, go.work, Makefile, and CI pipeline can build and test
the server module from day one.

To follow progress, see the refactor plan checked in at the repository root.
`

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	versionFlag := flag.Bool("version", false, "print build phase and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println("nucleus-admin-server skeleton (phase 1)")
		return
	}

	logger.Info("admin-server skeleton invoked", "phase", 1)
	fmt.Print(skeletonBanner)
}
