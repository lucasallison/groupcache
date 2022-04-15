package prefetcher

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func fileExists(dst string) bool {
	_, err := os.Stat(dst)
	return !errors.Is(err, os.ErrNotExist)
}

// creates subdirectories if they do not exists, otherwise does nothing
// TODO call this func pls
func createBaseDir() error {
	root, err := getRootSaveDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(root, os.ModePerm)
}

func getRootSaveDir() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}

	root = filepath.Join(root, "traces")
	return root, nil
}

func getUserTracePath(uid string) (string, error) {
	root, err := getRootSaveDir()
	// TODO extension?
	return filepath.Join(root, uid+".trace"), err
}

func deleteUserTrace(uid string) error {

	fpath, err := getUserTracePath(uid)
	if err != nil {
		return err
	}

	/* Nothing to remove */
	if !fileExists(fpath) {
		return nil
	}

	os.Remove(fpath)
	return nil
}

func writeUserTrace(uid string, trace []string) error {
	fpath, err := getUserTracePath(uid)

	file, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, t := range trace {
		fmt.Fprintln(w, t)
	}
	return w.Flush()

}

func getUserTrace(uid string) ([]string, error) {

	fpath, err := getUserTracePath(uid)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var trace []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		trace = append(trace, scanner.Text())
	}
	return trace, scanner.Err()
}
