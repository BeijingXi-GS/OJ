package filejudge

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestPreflightRefusesMissingIsolation(t *testing.T) {
	for _, name := range []string{"docker_unavailable", "windows_engine", "wrong_cgroup", "declared_volume", "leftover_container"} {
		t.Run(name, func(t *testing.T) {
			e := &Engine{image: "sha256:" + strings.Repeat("a", 64)}
			e.command = func(_ context.Context, args []string, out, _ io.Writer) error {
				if name == "docker_unavailable" {
					return io.ErrClosedPipe
				}
				var data string
				switch args[0] {
				case "version":
					data = "linux amd64"
					if name == "windows_engine" {
						data = "windows amd64"
					}
				case "info":
					data = "2"
					if name == "wrong_cgroup" {
						data = "1"
					}
				case "image":
					data = "null"
					if name == "declared_volume" {
						data = `{"/data":{}}`
					}
				case "ps":
					if name == "leftover_container" {
						data = "container-id"
					}
				default:
					t.Fatalf("preflight ran beyond failed isolation: %v", args)
				}
				_, _ = io.WriteString(out, data)
				return nil
			}
			if err := e.Preflight(context.Background()); err == nil {
				t.Fatal("missing isolation accepted")
			}
		})
	}
}
