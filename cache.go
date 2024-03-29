package jsoncache

import (
	"encoding/json"
	"fmt"
	"github.com/gearboxworks/go-status"
	"github.com/gearboxworks/go-status/only"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const ErrCacheMiss = "cache key not found"
const ErrExpired = "cache key expired"

type Cache struct {
	Dir     Dir
	Disable bool
}

type Wrapper struct {
	Expires string `json:"expires"`
	Data    string `json:"data"`
}

func New(dir Dir) *Cache {
	var err error
	for range only.Once {
		dir, err = homedir.Expand(dir)
		if err == nil {
			break
		}
		basedir := filepath.Base(dir)
		if basedir == "" {
			basedir = DefaultCacheDir
		}
		dir = fmt.Sprintf("%s%c%s",
			os.TempDir(),
			os.PathSeparator,
			basedir,
		)
	}
	return &Cache{
		Dir: dir,
	}
}

func (me *Cache) GetCacheFilepath(key Key) Filepath {
	return fmt.Sprintf("%s%c%s.json",
		me.Dir,
		os.PathSeparator,
		key,
	)
}

func (me *Cache) VerifyCacheFile(key Key) (fp Filepath, sts status.Status) {
	var f *os.File
	var err error
	for range only.Once {
		fp = me.GetCacheFilepath(key)
		f, err = os.Open(fp)
		if err != nil {
			pe, ok := err.(*os.PathError)
			if !ok {
				break
			}
			if pe.Err == syscall.ENOENT && pe.Op == "open" {
				err = fmt.Errorf(ErrCacheMiss)
				break
			}
			break
		}
	}
	me.close(f)
	if err != nil {
		var msg string
		if err.Error() == ErrCacheMiss {
			msg = fmt.Sprintf("no cache file for key '%s'", key)
		} else {
			msg = fmt.Sprintf("cannot open cache file for key '%s'", key)
		}
		sts = status.Wrap(err, &status.Args{
			Message: msg,
		})
	}
	return fp, sts
}

func (me *Cache) Clear(key Key) (sts status.Status) {
	for range only.Once {
		if me.Disable {
			break
		}
		err := os.Remove(string(me.GetCacheFilepath(key)))
		if err != nil {
			pe, ok := err.(*os.PathError)
			if !ok {
				break
			}
			if pe.Err == syscall.ENOENT && pe.Op == "open" {
				sts = status.Wrap(err, &status.Args{
					Message: fmt.Sprintf("unable to clear cache '%s'", key),
				})
				break
			}
		}
	}
	return sts
}

func (me *Cache) Get(key Key) (data []byte, ok bool, sts status.Status) {
	for range only.Once {
		if me.Disable {
			break
		}
		var fp Filepath
		fp, sts = me.VerifyCacheFile(key)
		if status.IsError(sts) {
			break
		}
		var b []byte
		b, err := ioutil.ReadFile(string(fp))
		if err != nil {
			sts = status.Wrap(err, &status.Args{
				Message: fmt.Sprintf("could not read file '%s'", fp),
				Help:    fmt.Sprintf("ensure you have permissions to read '%s'", fp),
			})
			break
		}
		w := Wrapper{}
		err = json.Unmarshal(b, &w)
		if err != nil {
			sts = status.Wrap(err, &status.Args{
				Message: fmt.Sprintf("could not unmarshal JSON in file '%s'", fp),
				Help:    fmt.Sprintf("try deleting the files your cache at '%s'", filepath.Dir(fp)),
			})
			break
		}
		data = []byte(w.Data)
		expires, err := time.Parse(time.RFC3339, w.Expires)
		if err != nil {
			sts = status.Wrap(err, &status.Args{
				Message: fmt.Sprintf("failed to calculate cache expiration for file '%s'", fp),
				Help:    fmt.Sprintf("try deleting the files your cache at '%s'", filepath.Dir(fp)),
			})
			break
		}
		if expires.Before(time.Now()) {
			//_ = os.Remove(fp)
			sts = status.Wrap(fmt.Errorf(ErrExpired), &status.Args{
				Success:    true,
				Message:    fmt.Sprintf("cache expired for key '%s'", key),
				HttpStatus: http.StatusOK,
			})
			break
		}
		sts = status.Success("cache retrieved for key '%s'", key)
		ok = true
	}
	return data, ok, sts
}

func (me *Cache) Set(key Key, b []byte, duration string) (sts status.Status) {

	for range only.Once {
		dur, err := time.ParseDuration(duration)
		if err != nil {
			break
		}
		w := &Wrapper{
			Expires: time.Now().Add(dur).Format(time.RFC3339),
			Data:    string(b),
		}
		b, err := json.Marshal(w)
		if err != nil {
			sts = status.Wrap(err).
				SetHelp(status.AllHelp, "this should never happen, so try rebooting. Or contacting support").
				SetMessage("could not marshal JSON to cache key '%s'", key)
			break
		}
		fp := me.GetCacheFilepath(key)
		d := filepath.Dir(string(fp))
		if !DirExists(d) {
			err = os.MkdirAll(filepath.Dir(string(fp)), 0777)
			if err != nil {
				sts = status.Wrap(err).
					SetMessage("unable to create cache directory '%s'", d).
					SetHelp(status.AllHelp, "ensure you have permissions to '%s'", filepath.Dir(d))
				break
			}
		}
		err = ioutil.WriteFile(string(fp), b, 0777)
		if err != nil {
			sts = status.Wrap(err, &status.Args{
				Message: fmt.Sprintf("unable to write to cache file '%s'", fp),
				Help:    fmt.Sprintf("ensure you have permissions to '%s'", filepath.Dir(d)),
			})
			break
		}
		sts = status.Success("cache set for key '%s'", key)

	}
	return sts
}

func (me *Cache) close(f *os.File) {
	_ = f.Close()
}
