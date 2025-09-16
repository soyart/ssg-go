package ssg_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/soyart/ssg-go"
)

func TestPrependHooks(t *testing.T) {
	var hook1 ssg.Hook = func(_ string, _ []byte) (_ []byte, _ error) { return []byte("hook1"), nil }
	var hook2 ssg.Hook = func(_ string, _ []byte) (_ []byte, _ error) { return []byte("hook2"), nil }
	var hook3 ssg.Hook = func(_ string, _ []byte) (_ []byte, _ error) { return []byte("hook3"), nil }
	var hook4 ssg.Hook = func(_ string, _ []byte) (_ []byte, _ error) { return []byte("hook4"), nil }

	assert := func(t *testing.T, s *ssg.Ssg) {
		hooks := s.Options().Hooks()
		for i := range hooks {
			var result []byte

			switch i {
			case 0:
				result, _ = hooks[i]("", nil)
				if string(result) == "hook1" {
					continue
				}
			case 1:
				result, _ = hooks[i]("", nil)
				if string(result) == "hook2" {
					continue
				}
			case 2:
				result, _ = hooks[i]("", nil)
				if string(result) == "hook3" {
					continue
				}
			case 3:
				result, _ = hooks[i]("", nil)
				if string(result) == "hook4" {
					continue
				}

			default:
				t.Errorf("unexpected index %d", i)
			}

			t.Errorf("unexpected value for hooks[%d]: %s", i, result)
		}
	}

	t.Run("imperative", func(t *testing.T) {
		s := new(ssg.Ssg)
		s.With(ssg.WithHooks(hook3, hook4))
		prepend := ssg.PrependHooks(hook1, hook2)
		s.With(prepend)
		assert(t, s)
	})

	t.Run("option slice", func(t *testing.T) {
		var opts []ssg.Option
		original := ssg.WithHooks(hook3, hook4)
		prepend := ssg.PrependHooks(hook1, hook2)
		opts = append(opts, original, prepend)

		s := new(ssg.Ssg)
		s.With(opts...)
		assert(t, s)
	})

	t.Run("option slice rev", func(t *testing.T) {
		var opts []ssg.Option
		original := ssg.WithHooks(hook3, hook4)
		prepend := ssg.PrependHooks(hook1, hook2)
		opts = append(opts, prepend, original)

		s := new(ssg.Ssg)
		s.With(opts...)
		assert(t, s)
	})
}

// inputHasher computes and stamps hash for input files.
// If the hash cannot be computed, the hasher output is ignored
//
// For the output file ${dst}/foo/bar.html generated
// from ${src}/foo/bar.md, a hash file ${dst}/foo/bar.md.sha256
// will be generated.
func inputHasher(s *ssg.Ssg) ssg.Pipeline {
	return func(path string, data []byte, d fs.DirEntry) (string, []byte, fs.DirEntry, error) {
		if d.IsDir() {
			return path, data, d, nil
		}

		var err error
		hashPath := path + ".sha256"
		hashPath, err = mirrorPath(s.Src, s.Dst, hashPath)
		if err != nil {
			return path, data, d, nil
		}
		hash, err := hashData(data)
		if err != nil {
			return path, data, d, nil
		}

		s.Outputs().Add(ssg.Output(hashPath, path, []byte(hash), 0o644))
		return path, data, d, nil
	}
}

func hashData(data []byte) (string, error) {
	// TODO: init once
	hash := sha256.New().Sum(data)
	if len(hash) == 0 {
		return "", nil
	}
	return hex.EncodeToString(hash), nil
}

func mirrorPath(src, dst, path string) (string, error) {
	path, err := filepath.Rel(src, path)
	if err != nil {
		return "", err
	}
	return filepath.Join(dst, path), nil
}

// Change all filenames except for index.md(s)
func pipeChangeFilename(path string, data []byte, d fs.DirEntry) (string, []byte, fs.DirEntry, error) {
	if d.IsDir() {
		return path, data, d, nil
	}
	base := filepath.Base(path)
	if base == "index.md" {
		return path, data, d, nil
	}
	return changeFilename(path), data, d, nil
}

func changeFilename(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)

	base = strings.TrimSuffix(base, ext) + "_mw_change_filename"
	parent := filepath.Dir(path)

	return filepath.Join(parent, fmt.Sprintf("%s%s", base, ext))
}

func TestChainPipelines(t *testing.T) {
	src := "./ssg-testdata/myblog/src"
	dst := "./ssg-testdata/myblog/dst-chain-pipes"

	err := os.RemoveAll(dst)
	if err != nil {
		panic(err)
	}
	ignore, err := ssg.ParseSsgIgnore(filepath.Join(src, ssg.SsgIgnore))
	if err != nil {
		panic(err)
	}

	ssg.Generate(src, dst, "TestChainPipelines", "https://chain.pipes",
		ssg.WithPipelines(
			pipeChangeFilename,
			inputHasher,

			// The last pipeline skips output for svg files
			func(path string, data []byte, d fs.DirEntry) (string, []byte, fs.DirEntry, error) {
				if d.IsDir() {
					return path, data, d, nil
				}
				if filepath.Ext(path) == ".svg" {
					return path, data, d, ssg.ErrSkipCore
				}
				return path, data, d, nil
			},
		),
	)

	// For file in src (e.g. name.old_ext),
	// we must have 2 corresponding files:
	//
	// ${name}_mw_change_filename.${old_ext}
	// ${name}_mw_change_filename.${old_ext}.sha256
	filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if ignore.Ignore(path) {
			return nil
		}
		mirrored, err := mirrorPath(src, dst, path)
		if err != nil {
			panic(err)
		}

		base := filepath.Base(path)
		switch base {
		case
			ssg.MarkerHeader,
			ssg.MarkerFooter,
			ssg.SsgIgnore:

			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		rehash, err := hashData(data)
		if err != nil {
			panic(err)
		}

		// index.md(s) names are not changed
		if base == "index.md" {
			// dst/foo/path/index.md.sha256
			hashPath := mirrored + ".sha256"
			stat, err := os.Stat(hashPath)
			if err != nil {
				t.Fatalf("failed to stat hash %s", hashPath)
			}
			if stat.IsDir() {
				t.Fatalf("unexpected dir at hash path %s", hashPath)
			}
			hashed, err := os.ReadFile(hashPath)
			if err != nil {
				t.Fatalf("failed to read hash %s", hashPath)
			}

			if !bytes.Equal(hashed, []byte(rehash)) {
				t.Fatalf("unexpected hash")
			}
			return nil
		}

		// file whose name was changed
		target := changeFilename(mirrored)
		target1 := target
		if strings.HasSuffix(target, ".md") {
			target1 = strings.TrimSuffix(target, ".md") + ".html"
		}

		if filepath.Ext(target1) == ".svg" {
			_, err = os.Stat(target1)
			if !os.IsNotExist(err) {
				t.Fatalf("expecting to not exist: originator='%s', mirrored='%s', target='%s'", path, mirrored, target)
			}
			return nil
		}

		_, err = os.Stat(target1)
		if err != nil {
			t.Fatalf("failed to stat file: originator='%s', mirrored='%s', target='%s'", path, mirrored, target)
		}

		hashPath := target + ".sha256"
		hashed, err := os.ReadFile(hashPath)
		if err != nil {
			t.Fatalf("failed to read file: originator='%s', mirrored='%s', target='%s'", path, mirrored, target)
		}
		if !bytes.Equal(hashed, []byte(rehash)) {
			t.Fatalf("unexpected hash")
		}

		return nil
	})

	err = os.RemoveAll(dst)
	if err != nil {
		panic(err)
	}
}
