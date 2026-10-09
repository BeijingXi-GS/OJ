package filejudge

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// Called only after the container lifecycle has completed without taint.
// Windows Go RemoveAll can fail on read-only artifacts. Preserve their read-only
// state throughout execution, then clear it only on the four owned artifact paths.
func removeJobFiles(jobsPath, name string) (err error) {
	if runtime.GOOS != "windows" {
		return os.RemoveAll(filepath.Join(jobsPath, name))
	}
	st, err := os.Lstat(jobsPath)
	if err != nil {
		return err
	}
	if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return errors.New("cleanup_root_invalid")
	}
	parent, err := os.OpenRoot(jobsPath)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, parent.Close()) }()
	opened, err := parent.Stat(".")
	if err != nil || !os.SameFile(st, opened) {
		return errors.New("cleanup_root_changed")
	}
	job, identity, err := cleanupChildRoot(parent, name)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	prepareErr := prepareJobArtifacts(job)
	closeErr := job.Close()
	if prepareErr != nil || closeErr != nil {
		return errors.Join(prepareErr, closeErr)
	}
	current, err := parent.Lstat(name)
	if err != nil || !current.IsDir() || current.Mode()&os.ModeSymlink != 0 || !os.SameFile(identity, current) {
		return errors.New("cleanup_job_changed")
	}
	return parent.RemoveAll(name)
}

func cleanupChildRoot(parent *os.Root, name string) (*os.Root, os.FileInfo, error) {
	st, err := parent.Lstat(name)
	if err != nil {
		return nil, nil, err
	}
	if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return nil, nil, errors.New("cleanup_directory_invalid")
	}
	child, err := parent.OpenRoot(name)
	if err != nil {
		return nil, nil, err
	}
	opened, err := child.Stat(".")
	if err != nil || !os.SameFile(st, opened) {
		return nil, nil, errors.Join(errors.New("cleanup_directory_changed"), child.Close())
	}
	return child, st, nil
}

func prepareJobArtifacts(job *os.Root) error {
	if err := makeArtifactRemovable(job); err != nil {
		return err
	}
	for _, name := range []string{"checker-build", "sender-build", "receiver-build"} {
		child, _, err := cleanupChildRoot(job, name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		prepareErr := makeArtifactRemovable(child)
		closeErr := child.Close()
		if prepareErr != nil || closeErr != nil {
			return errors.Join(prepareErr, closeErr)
		}
	}
	return nil
}

func makeArtifactRemovable(dir *os.Root) (err error) {
	before, err := dir.Lstat("program")
	if errors.Is(err, os.ErrNotExist) {
		return nil // A failed or unattempted compilation need not have an artifact.
	}
	if err != nil {
		return err
	}
	if !before.Mode().IsRegular() {
		return errors.New("cleanup_artifact_invalid")
	}
	f, err := dir.Open("program")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	opened, err := f.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(before, opened) {
		return errors.New("cleanup_artifact_changed")
	}
	// On Windows Root.Chmod opens the final component with O_OPEN_REPARSE,
	// so even a replaced final symlink is not followed to its target.
	if err := dir.Chmod("program", 0600); err != nil {
		return err
	}
	after, err := dir.Lstat("program")
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(opened, after) {
		return errors.New("cleanup_artifact_changed")
	}
	held, err := f.Stat()
	if err != nil || !held.Mode().IsRegular() || !os.SameFile(after, held) {
		return errors.New("cleanup_artifact_changed")
	}
	return nil
}
