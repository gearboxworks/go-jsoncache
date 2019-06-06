package jsoncache

import (
	"github.com/gearboxworks/go-status"
	"github.com/gearboxworks/go-status/only"
	"github.com/mitchellh/go-homedir"
	"os"
	"strings"
)

func ExtractPath(filepath Filepath, basedir Dir) (path Path, sts status.Status) {
	for range only.Once {
		var err error
		filepath, err = homedir.Expand(filepath)
		if err != nil {
			sts = status.Wrap(err).
				SetMessage("filepath '%s' could not be expanded: %s",
					filepath,
					err.Error(),
				)
		}
		basedir, err = homedir.Expand(basedir)
		if err != nil {
			sts = status.Wrap(err).
				SetMessage("filepath '%s' could not be expanded: %s",
					basedir,
					err.Error(),
				)
		}
		if strings.HasPrefix(filepath, basedir) {
			path = string([]byte(filepath)[len(basedir):])
		} else {
			path = filepath
		}
	}
	return path, sts
}

func DirExists(d string) bool {
	_, err := os.Stat(d)
	return !os.IsNotExist(err)
}
