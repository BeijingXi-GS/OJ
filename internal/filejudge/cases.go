package filejudge

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"

	"oj.local/oj/internal/problemstore"
)

type CaseFile = problemstore.CaseFile
type Manifest = problemstore.Manifest
type SnapshotCase struct {
	ID              string
	Input, Expected []byte
}
type Snapshot struct {
	Manifest      Manifest
	Cases         []SnapshotCase
	SHA256        string
	Interactor    string
	CheckerSource string
}

func regularRead(root *os.Root, name string) ([]byte, error) {
	st, err := root.Lstat(name)
	if err != nil || !st.Mode().IsRegular() {
		return nil, errors.New("testdata_not_regular")
	}
	f, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(st, opened) {
		return nil, errors.New("testdata_changed")
	}
	b, err := io.ReadAll(io.LimitReader(f, MaxFile+1))
	if err != nil || len(b) > MaxFile {
		return nil, errors.New("testdata_limit")
	}
	return b, nil
}
func safeFilename(s, suffix string) bool {
	return problemstore.SafeFilename(s, suffix)
}
func Load(rootPath, problem string) (Snapshot, error) {
	return LoadWithStore(&problemstore.Store{BuiltinRoot: rootPath}, problem)
}
func LoadWithStore(store *problemstore.Store, problem string) (Snapshot, error) {
	var result Snapshot
	if store == nil {
		return result, errors.New("problem_store_unavailable")
	}
	location, exists, err := store.Resolve(problem)
	if err != nil {
		return result, err
	}
	if !exists {
		return result, errors.New("unknown_problem")
	}
	rootPath := location.Root
	info, err := os.Lstat(rootPath)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return result, errors.New("testdata_root_invalid")
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return result, err
	}
	defer root.Close()
	st, err := root.Lstat(problem)
	if err != nil || !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return result, errors.New("problem_directory_invalid")
	}
	p, err := root.OpenRoot(problem)
	if err != nil {
		return result, err
	}
	defer p.Close()
	raw, err := regularRead(p, "manifest.json")
	if err != nil {
		return result, err
	}
	result.Manifest, err = problemstore.ParseManifest(raw)
	if err != nil {
		return result, errors.New("manifest_invalid")
	}
	m := result.Manifest
	if m.Type != location.Definition.Type {
		return result, errors.New("manifest_policy_invalid")
	}
	result.Interactor = m.Interactor
	h := sha256.New()
	add := func(b []byte) { _ = binary.Write(h, binary.LittleEndian, uint64(len(b))); _, _ = h.Write(b) }
	add(raw)
	if m.CheckerSource != "" {
		source, err := regularRead(p, m.CheckerSource)
		if err != nil {
			return result, err
		}
		if err = problemstore.ValidateCheckerSource(source); err != nil {
			return result, errors.New("checker_source_invalid")
		}
		result.CheckerSource = string(source)
		add(source)
	}
	for _, c := range m.Cases {
		input, err := regularRead(p, c.Input)
		if err != nil {
			return result, err
		}
		if err := problemstore.ValidateCaseInput(m, input); err != nil {
			return result, errors.New("input_invalid")
		}
		var expected []byte
		if c.Expected != "" {
			expected, err = regularRead(p, c.Expected)
			if err != nil {
				return result, err
			}
		}
		add([]byte(c.ID))
		add(input)
		add(expected)
		result.Cases = append(result.Cases, SnapshotCase{c.ID, input, expected})
	}
	result.SHA256 = hex.EncodeToString(h.Sum(nil))
	return result, nil
}
func hash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
