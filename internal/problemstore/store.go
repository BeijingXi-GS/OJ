package problemstore

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sort"

	"oj.local/oj/internal/problems"
)

type Store struct{ BuiltinRoot, UploadRoot string }
type Location struct {
	Definition problems.Definition
	Root       string
}
type registration struct {
	Version    int                 `json:"version"`
	ID         string              `json:"id"`
	Definition problems.Definition `json:"definition"`
}

func clone(p problems.Definition) problems.Definition {
	p.Samples = append([]problems.Sample(nil), p.Samples...)
	p.StarterCodes = problems.CloneStarterCodes(p.StarterCodes)
	return p
}
func unavailable() error { return errors.New("problem_store_unavailable") }

func openRoot(path string) (*os.Root, error) {
	st, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, unavailable()
	}
	if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return nil, unavailable()
	}
	r, err := os.OpenRoot(path)
	if err != nil {
		return nil, unavailable()
	}
	return r, nil
}

func read(root *os.Root, name string) ([]byte, error) {
	st, err := root.Lstat(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, unavailable()
	}
	if !st.Mode().IsRegular() {
		return nil, unavailable()
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, unavailable()
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(st, opened) {
		return nil, unavailable()
	}
	b, err := io.ReadAll(io.LimitReader(f, MaxFile+1))
	if err != nil || len(b) > MaxFile {
		return nil, unavailable()
	}
	return b, nil
}

type registrationRoot interface {
	Lstat(string) (os.FileInfo, error)
	Open(string) (*os.File, error)
	OpenRoot(string) (*os.Root, error)
}

func pendingName(id string) string { return ".staging-" + id }

// A pending publication is an empty, non-link directory. A malformed marker
// must not silently hide corruption or permit an incomplete publication.
func pending(root registrationRoot, id string) (bool, error) {
	name := pendingName(id)
	st, err := root.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return false, unavailable()
	}
	f, err := root.Open(name)
	if errors.Is(err, os.ErrNotExist) {
		// A concurrent successful commit may remove the marker after Lstat.
		if _, err := root.Lstat(name); errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
	}
	if err != nil {
		return false, unavailable()
	}
	opened, statErr := f.Stat()
	if statErr != nil || !opened.IsDir() || !os.SameFile(st, opened) {
		_ = f.Close()
		return false, unavailable()
	}
	entries, readErr := f.ReadDir(1)
	closeErr := f.Close()
	if len(entries) != 0 || !errors.Is(readErr, io.EOF) || closeErr != nil {
		return false, unavailable()
	}
	return true, nil
}

func registered(root registrationRoot, id string) (definition problems.Definition, found bool, resultErr error) {
	// Check on every exit, including errors caused by an in-progress write. This
	// keeps a partial registration from turning List into a transient failure.
	defer func() {
		isPending, err := pending(root, id)
		if err != nil || isPending {
			definition, found, resultErr = problems.Definition{}, false, err
		}
	}()
	if isPending, err := pending(root, id); err != nil || isPending {
		return problems.Definition{}, false, err
	}
	st, err := root.Lstat(id)
	if errors.Is(err, os.ErrNotExist) {
		return problems.Definition{}, false, nil
	}
	if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return problems.Definition{}, false, unavailable()
	}
	p, err := root.OpenRoot(id)
	if err != nil {
		return problems.Definition{}, false, unavailable()
	}
	defer p.Close()
	// The directory may have been created after the initial pending check.
	// Do not read any files until its publication state has been checked again.
	if isPending, err := pending(root, id); err != nil || isPending {
		return problems.Definition{}, false, err
	}
	raw, err := read(p, "problem.json")
	if errors.Is(err, os.ErrNotExist) {
		return problems.Definition{}, false, nil
	}
	if err != nil {
		return problems.Definition{}, false, unavailable()
	}
	var r registration
	if strictJSON(raw, &r) != nil || r.Version != 1 || r.ID != id || r.Definition.ID != id || validateDefinition(r.Definition) != nil {
		return problems.Definition{}, false, unavailable()
	}
	manifestRaw, err := read(p, "manifest.json")
	if err != nil {
		return problems.Definition{}, false, unavailable()
	}
	var manifest Manifest
	if strictJSON(manifestRaw, &manifest) != nil {
		return problems.Definition{}, false, unavailable()
	}
	manifest, err = NormalizeManifest(manifest)
	if err != nil || manifest.Type != r.Definition.Type {
		return problems.Definition{}, false, unavailable()
	}
	return clone(r.Definition), true, nil
}

func (s *Store) Resolve(id string) (Location, bool, error) {
	if p, ok := problems.Get(id); ok {
		return Location{p, s.BuiltinRoot}, true, nil
	}
	if s.UploadRoot == "" || !uploadIDRE.MatchString(id) {
		return Location{}, false, nil
	}
	r, err := openRoot(s.UploadRoot)
	if errors.Is(err, os.ErrNotExist) {
		return Location{}, false, nil
	}
	if err != nil {
		return Location{}, false, err
	}
	defer r.Close()
	p, ok, err := registered(r, id)
	return Location{p, s.UploadRoot}, ok, err
}

func (s *Store) Get(id string) (problems.Definition, bool, error) {
	location, ok, err := s.Resolve(id)
	return location.Definition, ok, err
}

func (s *Store) List() ([]problems.Definition, error) {
	list := problems.List()
	if s.UploadRoot == "" {
		return list, nil
	}
	r, err := openRoot(s.UploadRoot)
	if errors.Is(err, os.ErrNotExist) {
		return list, nil
	}
	if err != nil {
		return nil, err
	}
	defer r.Close()
	d, err := r.Open(".")
	if err != nil {
		return nil, unavailable()
	}
	defer d.Close()
	entries, err := d.ReadDir(-1)
	if err != nil {
		return nil, unavailable()
	}
	ids := []string{}
	for _, entry := range entries {
		if uploadIDRE.MatchString(entry.Name()) {
			ids = append(ids, entry.Name())
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		p, ok, err := registered(r, id)
		if err != nil {
			return nil, err
		}
		if ok {
			list = append(list, p)
		}
	}
	return list, nil
}

func write(root *os.Root, name string, data []byte) error {
	f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return unavailable()
	}
	var n int
	if n, err = f.Write(data); err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		return unavailable()
	}
	return nil
}

type publicationRoot interface {
	Mkdir(string, os.FileMode) error
	OpenRoot(string) (*os.Root, error)
	Remove(string) error
	RemoveAll(string) error
}

// publish makes the directory visible only after populate and Close succeed.
// The marker also keeps an interrupted or failed rollback hidden after restart.
func publish(root publicationRoot, id string, populate func(*os.Root) error) error {
	if !uploadIDRE.MatchString(id) {
		return unavailable()
	}
	stage := pendingName(id)
	if err := root.Mkdir(stage, 0700); err != nil {
		return unavailable()
	}
	created, published := false, false
	defer func() {
		if !published {
			// An existing ID is never ours to clean. Retain the marker whenever
			// cleanup of a directory that we did create cannot be confirmed.
			if !created || root.RemoveAll(id) == nil {
				_ = root.Remove(stage)
			}
		}
	}()
	if err := root.Mkdir(id, 0700); err != nil {
		return unavailable()
	}
	created = true
	final, err := root.OpenRoot(id)
	if err != nil {
		return unavailable()
	}
	writeErr := populate(final)
	closeErr := final.Close()
	if writeErr != nil || closeErr != nil {
		return unavailable()
	}
	if err := root.Remove(stage); err != nil {
		return unavailable()
	}
	published = true
	return nil
}

// Create publishes only complete immutable problem directories. Its generated ID
// is never accepted from a descriptor and cannot overwrite a built-in problem.
func (s *Store) Create(upload Upload) (problems.Definition, error) {
	files := make(map[string][]byte, len(upload.Files))
	for name, data := range upload.Files {
		files[name] = append([]byte(nil), data...)
	}
	upload.Files = files
	upload.Descriptor.Cases = append([]CaseFile(nil), upload.Descriptor.Cases...)
	upload.Descriptor.Samples = append([]problems.Sample(nil), upload.Descriptor.Samples...)
	upload.Descriptor.StarterCodes = problems.CloneStarterCodes(upload.Descriptor.StarterCodes)
	if upload.Descriptor.Communication != nil {
		config := *upload.Descriptor.Communication
		upload.Descriptor.Communication = &config
	}
	p, m, err := validateUpload(upload)
	if err != nil {
		return problems.Definition{}, err
	}
	if s.UploadRoot == "" {
		return problems.Definition{}, unavailable()
	}
	if err := os.MkdirAll(s.UploadRoot, 0700); err != nil {
		return problems.Definition{}, unavailable()
	}
	r, err := openRoot(s.UploadRoot)
	if err != nil {
		return problems.Definition{}, unavailable()
	}
	defer r.Close()
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return problems.Definition{}, unavailable()
	}
	p.ID = "upload-" + hex.EncodeToString(random[:])
	manifestRaw, err := json.Marshal(m)
	if err != nil || len(manifestRaw) > MaxFile {
		return problems.Definition{}, invalid("manifest_limit")
	}
	registrationRaw, err := json.Marshal(registration{Version: 1, ID: p.ID, Definition: p})
	if err != nil || len(registrationRaw) > MaxFile {
		return problems.Definition{}, invalid("registration_limit")
	}
	if err := publish(r, p.ID, func(final *os.Root) error {
		for name, data := range files {
			if err := write(final, name, data); err != nil {
				return err
			}
		}
		if err := write(final, "manifest.json", manifestRaw); err != nil {
			return err
		}
		return write(final, "problem.json", registrationRaw)
	}); err != nil {
		return problems.Definition{}, err
	}
	return clone(p), nil
}
