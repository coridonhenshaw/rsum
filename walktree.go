package main

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type WalkTreeCallbackInterface interface {
	PreDirectory(Path string) (Accept bool)
	Directory(Path string) (Continue bool)
	File(Path string, f os.FileInfo) (Continue bool)
	DirectoryError(Path string, err error) (Continue bool)
	NonFileObject(Path string, f os.FileInfo) (Continue bool)
}

func WalkTree(Path string, FollowSymlinks bool, Recurse bool, Callback WalkTreeCallbackInterface) bool {

	if !Callback.PreDirectory(Path) {
		return true
	}
	f, err := os.Open(Path)
	if err != nil {
		return Callback.DirectoryError(Path, err)
	}
	defer f.Close()

	Callback.Directory(Path)

	for {
		files, err := f.Readdir(1)

		if err != nil {
			if err == io.EOF {
				break
			}
			//			fmt.Println(files[0].Name())

			return Callback.DirectoryError(Path, err)
		}

		file := files[0]
		Fullname := filepath.Join(Path, file.Name())
		Mode := file.Mode()

		const RejectMask = fs.ModeCharDevice | fs.ModeDevice | fs.ModeNamedPipe | fs.ModeSocket

		if (Mode & RejectMask) != 0 {
			Callback.NonFileObject(Fullname, file)
			continue
		}

		if (Mode & fs.ModeSymlink) != 0 {
			if !FollowSymlinks {
				continue
			}

			fn, file, err := ParseSymlink(Fullname, file)
			if err != nil {
				Callback.DirectoryError(Fullname, err)
			}
			if fn == false {
				continue
			}

			Mode = file.Mode()
		}

		if (Mode & fs.ModeDir) != 0 {
			if Recurse {
				rc := WalkTree(Fullname, FollowSymlinks, Recurse, Callback)
				if rc == false {
					return false
				}
			}
		} else {
			Callback.File(Fullname, file)
		}
	}

	return true
}

func ParseSymlink(Filename string, f os.FileInfo) (bool, os.FileInfo, error) {
	Target, err := os.Readlink(Filename)
	if err != nil {
		return false, nil, err
	}

	if Target == "." {
		// Circular reference
		return false, nil, nil
	}

	Target = filepath.Clean(Target)

	f, err = os.Stat(Target)
	if err != nil {
		return false, nil, err
	}

	if f.Mode().IsDir() || f.Mode().IsRegular() {
		return true, f, nil
	}

	return false, nil, nil
}
