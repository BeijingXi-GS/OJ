package problemstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"oj.local/oj/internal/problems"
)

func publicationFixture(t *testing.T, id string) (problems.Definition, map[string][]byte) {
	t.Helper()
	u := validUpload()
	p, manifest, err := validateUpload(u)
	if err != nil {
		t.Fatal(err)
	}
	p.ID = id
	u.Files["manifest.json"], err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	u.Files["problem.json"], err = json.Marshal(registration{Version: 1, ID: id, Definition: p})
	if err != nil {
		t.Fatal(err)
	}
	return p, u.Files
}

func writePublicationFiles(root *os.Root, files map[string][]byte) error {
	for name, data := range files {
		if err := write(root, name, data); err != nil {
			return err
		}
	}
	return nil
}

func testStoreRoot(t *testing.T) (*Store, *os.Root) {
	t.Helper()
	s := &Store{UploadRoot: t.TempDir()}
	r, err := os.OpenRoot(s.UploadRoot)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return s, r
}

func assertHidden(t *testing.T, store *Store, id string) {
	t.Helper()
	if _, ok, err := store.Get(id); err != nil || ok {
		t.Fatalf("pending publication exposed or poisoned Get: exists=%v error=%v", ok, err)
	}
	list, err := store.List()
	if err != nil || len(list) != len(problems.List()) {
		t.Fatalf("pending publication exposed or poisoned List: count=%d error=%v", len(list), err)
	}
}

func TestPublicationHidesPartialAndCompleteFilesUntilCommit(t *testing.T) {
	s, root := testStoreRoot(t)
	id := "upload-" + strings.Repeat("1", 32)
	_, files := publicationFixture(t, id)
	partial, complete := make(chan struct{}), make(chan struct{})
	finishPartial, finishComplete := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	var partialOnce, completeOnce sync.Once
	releasePartial := func() { partialOnce.Do(func() { close(finishPartial) }) }
	releaseComplete := func() { completeOnce.Do(func() { close(finishComplete) }) }
	defer releasePartial()
	defer releaseComplete()
	go func() {
		done <- publish(root, id, func(final *os.Root) error {
			if err := write(final, "problem.json", []byte(`{"version":`)); err != nil {
				return err
			}
			close(partial)
			<-finishPartial
			if err := final.Remove("problem.json"); err != nil {
				return err
			}
			if err := writePublicationFiles(final, files); err != nil {
				return err
			}
			close(complete)
			<-finishComplete
			return nil
		})
	}()
	select {
	case <-partial:
	case err := <-done:
		t.Fatal("publication failed before partial phase", err)
	}
	for range 8 {
		assertHidden(t, s, id)
	}
	releasePartial()
	select {
	case <-complete:
	case err := <-done:
		t.Fatal("publication failed before complete phase", err)
	}
	for range 8 {
		assertHidden(t, s, id)
	}
	releaseComplete()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	restarted := &Store{UploadRoot: s.UploadRoot}
	if p, ok, err := restarted.Get(id); err != nil || !ok || p.ID != id {
		t.Fatal("committed registration unavailable after restart", ok, err)
	}
	entries, err := os.ReadDir(s.UploadRoot)
	if err != nil || len(entries) != 1 || entries[0].Name() != id {
		t.Fatal("successful publication left a marker", err)
	}
}

type publicationFaultRoot struct {
	*os.Root
	openErr, commitErr, cleanupErr error
	removed                        []string
	cleaned                        []string
	opened                         *os.Root
	closedBeforeCommit             bool
}

func (r *publicationFaultRoot) OpenRoot(name string) (*os.Root, error) {
	if r.openErr != nil {
		return nil, r.openErr
	}
	opened, err := r.Root.OpenRoot(name)
	r.opened = opened
	return opened, err
}

func (r *publicationFaultRoot) Remove(name string) error {
	r.removed = append(r.removed, name)
	if r.opened != nil {
		_, err := r.opened.Lstat(".")
		r.closedBeforeCommit = errors.Is(err, os.ErrClosed)
	}
	if r.commitErr != nil {
		return r.commitErr
	}
	return r.Root.Remove(name)
}

func (r *publicationFaultRoot) RemoveAll(name string) error {
	r.cleaned = append(r.cleaned, name)
	if r.cleanupErr != nil {
		return r.cleanupErr
	}
	return r.Root.RemoveAll(name)
}

func TestPublicationFailureRollbackAndOwnership(t *testing.T) {
	for _, name := range []string{"open", "write", "commit", "write_cleanup", "commit_cleanup", "existing_final", "existing_pending"} {
		t.Run(name, func(t *testing.T) {
			s, root := testStoreRoot(t)
			id := "upload-" + strings.Repeat("2", 32)
			_, files := publicationFixture(t, id)
			fault := errors.New("injected private path must not escape")
			wrapped := &publicationFaultRoot{Root: root}
			if name == "open" {
				wrapped.openErr = fault
			}
			if strings.HasPrefix(name, "commit") {
				wrapped.commitErr = fault
			}
			if strings.HasSuffix(name, "cleanup") {
				wrapped.cleanupErr = fault
			}
			existing := ""
			if name == "existing_final" {
				existing = id
			} else if name == "existing_pending" {
				existing = pendingName(id)
			}
			if existing != "" {
				if err := root.Mkdir(existing, 0700); err != nil {
					t.Fatal(err)
				}
				if err := write(root, existing+"/keep", []byte("existing data")); err != nil {
					t.Fatal(err)
				}
			}
			err := publish(wrapped, id, func(final *os.Root) error {
				if strings.HasPrefix(name, "write") {
					if err := write(final, "problem.json", []byte(`{"version":`)); err != nil {
						return err
					}
					return fault
				}
				return writePublicationFiles(final, files)
			})
			if err == nil || err.Error() != "problem_store_unavailable" {
				t.Fatal("failed publication exposed internal error or succeeded", err)
			}
			if existing != "" {
				b, err := os.ReadFile(filepath.Join(s.UploadRoot, existing, "keep"))
				if err != nil || string(b) != "existing data" || len(wrapped.cleaned) != 0 {
					t.Fatal("publication touched an existing directory", err, wrapped.cleaned)
				}
				if name == "existing_pending" && len(wrapped.removed) != 0 {
					t.Fatal("publication removed an unowned marker")
				}
				if name == "existing_final" {
					if _, err := root.Lstat(pendingName(id)); !errors.Is(err, os.ErrNotExist) {
						t.Fatal("owned marker retained after final directory collision", err)
					}
				}
				return
			}
			if len(wrapped.cleaned) != 1 || wrapped.cleaned[0] != id {
				t.Fatal("rollback failed to clean exactly the owned final directory", wrapped.cleaned)
			}
			if strings.HasSuffix(name, "cleanup") {
				if _, err := root.Lstat(id); err != nil {
					t.Fatal("failed rollback lost retained directory", err)
				}
				if st, err := root.Lstat(pendingName(id)); err != nil || !st.IsDir() {
					t.Fatal("failed rollback lost pending marker", err)
				}
			} else if _, err := root.Lstat(id); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("rollback left final directory", err)
			}
			if name == "write_cleanup" && len(wrapped.removed) != 0 {
				t.Fatal("failed rollback attempted visibility commit")
			}
			assertHidden(t, &Store{UploadRoot: s.UploadRoot}, id)
		})
	}
}

func TestPublicationClosesFinalRootBeforeCommit(t *testing.T) {
	_, root := testStoreRoot(t)
	id := "upload-" + strings.Repeat("3", 32)
	_, files := publicationFixture(t, id)
	wrapped := &publicationFaultRoot{Root: root}
	if err := publish(wrapped, id, func(final *os.Root) error { return writePublicationFiles(final, files) }); err != nil {
		t.Fatal(err)
	}
	if !wrapped.closedBeforeCommit || len(wrapped.cleaned) != 0 || len(wrapped.removed) != 1 {
		t.Fatal("commit did not follow final root closure or removed published files")
	}
}

func TestUnmarkedVersionOneRegistrationRemainsReadable(t *testing.T) {
	s, root := testStoreRoot(t)
	id := "upload-" + strings.Repeat("4", 32)
	p, files := publicationFixture(t, id)
	if err := root.Mkdir(id, 0700); err != nil {
		t.Fatal(err)
	}
	final, err := root.OpenRoot(id)
	if err != nil {
		t.Fatal(err)
	}
	writeErr := writePublicationFiles(final, files)
	closeErr := final.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal(writeErr, closeErr)
	}
	if got, ok, err := s.Get(id); err != nil || !ok || got.Title != p.Title {
		t.Fatal("legacy version one registration rejected", err)
	}
	list, err := s.List()
	if err != nil || len(list) != len(problems.List())+1 {
		t.Fatal("legacy registration absent from list", err)
	}
}

func TestMalformedPendingMarkersFailClosed(t *testing.T) {
	for _, kind := range []string{"file", "nonempty", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			s, root := testStoreRoot(t)
			p, err := s.Create(validUpload())
			if err != nil {
				t.Fatal(err)
			}
			marker := pendingName(p.ID)
			switch kind {
			case "file":
				err = write(root, marker, nil)
			case "nonempty":
				err = root.Mkdir(marker, 0700)
				if err == nil {
					err = write(root, marker+"/unexpected", nil)
				}
			case "symlink":
				err = os.Symlink(p.ID, filepath.Join(s.UploadRoot, marker))
				if err != nil {
					t.Skip("OS does not permit symlink creation")
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if _, ok, err := s.Get(p.ID); err == nil || ok || err.Error() != "problem_store_unavailable" {
				t.Fatal("malformed marker was accepted or exposed private error", ok, err)
			}
			if _, err := s.List(); err == nil {
				t.Fatal("malformed marker ignored by List")
			}
		})
	}
}

type registrationHookRoot struct {
	*os.Root
	stat func(string) (os.FileInfo, error)
	open func(string) (*os.File, error)
}

func (r registrationHookRoot) Lstat(name string) (os.FileInfo, error) {
	if r.stat != nil {
		return r.stat(name)
	}
	return r.Root.Lstat(name)
}

func (r registrationHookRoot) Open(name string) (*os.File, error) {
	if r.open != nil {
		return r.open(name)
	}
	return r.Root.Open(name)
}

func TestPendingDisappearanceAndIOFailure(t *testing.T) {
	_, root := testStoreRoot(t)
	id := "upload-" + strings.Repeat("5", 32)
	marker := pendingName(id)
	if err := root.Mkdir(marker, 0700); err != nil {
		t.Fatal(err)
	}
	wrapped := registrationHookRoot{Root: root, open: func(name string) (*os.File, error) {
		if err := root.Remove(name); err != nil {
			return nil, err
		}
		return nil, os.ErrNotExist
	}}
	if found, err := pending(wrapped, id); err != nil || found {
		t.Fatal("concurrent commit treated as corruption", found, err)
	}
	for _, phase := range []string{"stat", "open", "disappearance_recheck"} {
		t.Run(phase, func(t *testing.T) {
			fault := errors.New("private filesystem detail")
			wrapped := registrationHookRoot{Root: root}
			if phase == "stat" {
				wrapped.stat = func(string) (os.FileInfo, error) { return nil, fault }
			} else {
				if err := root.Mkdir(marker, 0700); err != nil {
					t.Fatal(err)
				}
				defer func() { _ = root.Remove(marker) }()
				wrapped.open = func(string) (*os.File, error) { return nil, fault }
				if phase == "disappearance_recheck" {
					wrapped.open = func(string) (*os.File, error) { return nil, os.ErrNotExist }
				}
			}
			if _, found, err := registered(wrapped, id); err == nil || found || err.Error() != "problem_store_unavailable" {
				t.Fatal("pending I/O failure did not fail closed", found, err)
			}
		})
	}
}

func TestRegistrationRechecksPendingAroundDirectoryAndResult(t *testing.T) {
	for _, phase := range []string{"after_initial_check", "before_success", "before_error"} {
		t.Run(phase, func(t *testing.T) {
			s, root := testStoreRoot(t)
			p, err := s.Create(validUpload())
			if err != nil {
				t.Fatal(err)
			}
			if phase != "before_success" {
				if err := os.WriteFile(filepath.Join(s.UploadRoot, p.ID, "problem.json"), []byte(`{"version":`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			checks := 0
			wrapped := registrationHookRoot{Root: root, stat: func(name string) (os.FileInfo, error) {
				if name == pendingName(p.ID) {
					checks++
					if (phase == "after_initial_check" && checks == 2) || (phase != "after_initial_check" && checks == 3) {
						if err := root.Mkdir(name, 0700); err != nil {
							return nil, err
						}
					}
				}
				return root.Lstat(name)
			}}
			if _, ok, err := registered(wrapped, p.ID); err != nil || ok || checks != 3 {
				t.Fatal("publication crossed a registration visibility boundary", ok, err, checks)
			}
		})
	}
}

func TestConcurrentPublicationReaders(t *testing.T) {
	s, root := testStoreRoot(t)
	const count = 12
	ids := make(chan string, count)
	stop, readerDone := make(chan struct{}), make(chan error, 1)
	go func() {
		for {
			select {
			case <-stop:
				readerDone <- nil
				return
			default:
			}
			if _, err := s.List(); err != nil {
				readerDone <- err
				return
			}
			select {
			case id := <-ids:
				if _, _, err := s.Get(id); err != nil {
					readerDone <- err
					return
				}
			default:
			}
		}
	}()
	var publishErr error
	for i := range count {
		id := fmt.Sprintf("upload-%032x", i)
		_, files := publicationFixture(t, id)
		ids <- id
		if err := publish(root, id, func(final *os.Root) error { return writePublicationFiles(final, files) }); err != nil {
			publishErr = err
			break
		}
	}
	close(stop)
	if readErr := <-readerDone; readErr != nil || publishErr != nil {
		t.Fatal("concurrent reader or publisher failed", readErr, publishErr)
	}
	list, err := s.List()
	if err != nil || len(list) != len(problems.List())+count {
		t.Fatal("concurrent publication lost registrations", len(list), err)
	}
}
