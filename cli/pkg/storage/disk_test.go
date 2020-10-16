package storage

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"replicate.ai/cli/pkg/files"
)

func TestDiskStorageGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(dir, "some-file"), []byte("hello"), 0644)
	require.NoError(t, err)

	_, err = storage.Get("does-not-exist")
	require.IsType(t, &DoesNotExistError{}, err)

	content, err := storage.Get("some-file")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), content)
}
func TestDiskGetPathTar(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)

	tmpDir, err := files.TempDir("test")
	require.NoError(t, err)
	err = storage.GetPathTar("does-not-exist.tar.gz", tmpDir)
	require.IsType(t, &DoesNotExistError{}, err)
}

func TestDiskStoragePut(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)

	err = storage.Put("some-file", []byte("hello"))
	require.NoError(t, err)

	content, err := ioutil.ReadFile(path.Join(dir, "some-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), content)

	err = storage.Put("subdirectory/another-file", []byte("hello again"))
	require.NoError(t, err)

	content, err = ioutil.ReadFile(path.Join(dir, "subdirectory/another-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello again"), content)
}

func TestDiskStorageList(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)

	err = storage.Put("some-file", []byte("hello"))
	require.NoError(t, err)
	err = storage.Put("dir/another-file", []byte("hello"))
	require.NoError(t, err)

	paths, err := storage.List("")
	require.NoError(t, err)
	require.Equal(t, []string{"some-file"}, paths)

	paths, err = storage.List("dir")
	require.NoError(t, err)
	require.Equal(t, []string{"dir/another-file"}, paths)

	paths, err = storage.List("dir-that-does-not-exist")
	require.NoError(t, err)
	require.Equal(t, []string{}, paths)
}

func TestPutPath(t *testing.T) {
	storageDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(storageDir)

	storage, err := NewDiskStorage(storageDir)
	require.NoError(t, err)

	workDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workDir)
	require.NoError(t, ioutil.WriteFile(path.Join(workDir, "some-file"), []byte("hello"), 0644))
	require.NoError(t, os.Mkdir(path.Join(workDir, "subdirectory"), 0755))
	require.NoError(t, ioutil.WriteFile(path.Join(workDir, "subdirectory/another-file"), []byte("hello again"), 0644))

	err = storage.PutPath(workDir, "parent")
	require.NoError(t, err)

	content, err := ioutil.ReadFile(path.Join(storageDir, "parent/some-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), content)

	content, err = ioutil.ReadFile(path.Join(storageDir, "parent/subdirectory/another-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello again"), content)
}

func TestDiskListRecursive(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Works with emty storage
	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)
	results := make(chan ListResult)
	go storage.ListRecursive(results, "checkpoints")
	require.Empty(t, <-results)

	// Lists stuff!
	require.NoError(t, storage.Put("checkpoints/abc123.json", []byte("yep")))
	require.NoError(t, storage.Put("experiments/def456.json", []byte("nope")))
	results = make(chan ListResult)
	go storage.ListRecursive(results, "checkpoints")
	require.Equal(t, ListResult{
		Path: "checkpoints/abc123.json",
		MD5:  []byte{0x93, 0x48, 0xae, 0x78, 0x51, 0xcf, 0x3b, 0xa7, 0x98, 0xd9, 0x56, 0x4e, 0xf3, 0x8, 0xec, 0x25},
	}, <-results)
	require.Empty(t, <-results)
}

func TestDiskMatchFilenamesRecursive(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Works with emty storage
	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)
	results := make(chan ListResult)
	go storage.MatchFilenamesRecursive(results, "checkpoints", "replicate-metadata.json")
	v := <-results
	require.Empty(t, v)
}
