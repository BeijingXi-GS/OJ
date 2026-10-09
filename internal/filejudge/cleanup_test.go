package filejudge

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRemoveJobRejectsUnownedAndTaintedPaths(t *testing.T) {
	root := t.TempDir()
	e := &Engine{jobsRoot: root}
	for _, target := range []string{root, filepath.Dir(root), filepath.Join(root, "other-job"), filepath.Join(root, "run-nested", "run-child")} {
		if err := e.removeJob(target); err == nil {
			t.Fatalf("unowned cleanup accepted: %s", target)
		}
	}
	dir := filepath.Join(root, "run-retain")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "program")
	if err := os.WriteFile(p, []byte("trusted artifact"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(p, 0600) })
	e.tainted.Store(true)
	if err := e.removeJob(dir); err == nil {
		t.Fatal("tainted job cleanup accepted")
	}
	st, err := os.Stat(p)
	if err != nil || st.Mode().Perm()&0200 != 0 {
		t.Fatal("tainted artifact modified", err)
	}
}

func TestWindowsRemoveJobReadOnlyArtifacts(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows attribute compatibility")
	}
	root := t.TempDir()
	dir := filepath.Join(root, "run-readonly")
	for _, build := range []string{"", "checker-build", "sender-build", "receiver-build"} {
		folder := filepath.Join(dir, build)
		if err := os.MkdirAll(folder, 0700); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(folder, "program")
		if err := os.WriteFile(p, []byte("trusted fake artifact"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(p, 0555); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(p, 0600) })
	}
	e := &Engine{jobsRoot: root}
	if err := e.removeJob(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(dir); !os.IsNotExist(err) {
		t.Fatal("job remains", err)
	}
	if err := e.removeJob(dir); err != nil {
		t.Fatal("absent job cleanup", err)
	}
	empty := filepath.Join(root, "run-uncompiled")
	if err := os.Mkdir(empty, 0700); err != nil {
		t.Fatal(err)
	}
	if err := e.removeJob(empty); err != nil {
		t.Fatal("absent roles", err)
	}
}

func TestWindowsRemoveJobRejectsUnexpectedArtifactType(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows attribute compatibility")
	}
	for _, relative := range []string{"program", filepath.Join("sender-build", "program")} {
		t.Run(relative, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "run-invalid")
			if err := os.MkdirAll(filepath.Join(dir, relative), 0700); err != nil {
				t.Fatal(err)
			}
			e := &Engine{jobsRoot: root}
			if err := e.removeJob(dir); err == nil {
				t.Fatal("nonregular artifact accepted")
			}
			if _, err := os.Stat(dir); err != nil {
				t.Fatal("failure evidence removed", err)
			}
		})
	}
}

func TestWindowsRemoveJobDoesNotFollowLinks(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows attribute compatibility")
	}
	for _, kind := range []string{"job", "build", "artifact"} {
		t.Run(kind, func(t *testing.T) {
			root, outside := t.TempDir(), t.TempDir()
			artifact := filepath.Join(outside, "program")
			if err := os.WriteFile(artifact, []byte("outside evidence"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(artifact, 0555); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(artifact, 0600) })
			dir := filepath.Join(root, "run-linked")
			link, target := dir, outside
			if kind != "job" {
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
				link = filepath.Join(dir, "sender-build")
				if kind == "artifact" {
					link, target = filepath.Join(dir, "program"), artifact
				}
			}
			if err := os.Symlink(target, link); err != nil {
				t.Skipf("Windows symlink unavailable: %v", err)
			}
			t.Cleanup(func() { _ = os.Remove(link) })
			e := &Engine{jobsRoot: root}
			if err := e.removeJob(dir); err == nil {
				t.Fatal("linked cleanup accepted")
			}
			st, err := os.Stat(artifact)
			if err != nil || st.Mode().Perm()&0200 != 0 {
				t.Fatal("external artifact changed", err)
			}
			data, err := os.ReadFile(artifact)
			if err != nil || string(data) != "outside evidence" {
				t.Fatal("external artifact damaged", err)
			}
		})
	}
}

func TestWindowsRemoveJobDoesNotChmodUnknownPaths(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows attribute compatibility")
	}
	root := t.TempDir()
	dir := filepath.Join(root, "run-unknown")
	folder := filepath.Join(dir, "unknown-build")
	if err := os.MkdirAll(folder, 0700); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(folder, "program")
	if err := os.WriteFile(p, []byte("unexpected artifact"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(p, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(p, 0600) })
	e := &Engine{jobsRoot: root}
	if err := e.removeJob(dir); err == nil {
		t.Fatal("unexpected read-only path silently made writable")
	}
	st, err := os.Stat(p)
	if err != nil || st.Mode().Perm()&0200 != 0 {
		t.Fatal("unknown artifact changed", err)
	}
}
