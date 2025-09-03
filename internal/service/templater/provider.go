package templater

import (
	"embed"
	"path"
)

type EmbeddedProvider struct {
	FS    embed.FS
	Paths []string
}

func (o *EmbeddedProvider) Get(name string) (string, error) {
	var filename string

	var paths []string
	if o.Paths != nil {
		paths = o.Paths
	} else {
		paths = []string{""}
	}

	exts := []string{".mustache"}

	for _, p := range paths {
		for _, e := range exts {
			name := path.Join(p, name+e)
			f, err := o.FS.Open(name)
			if err == nil {
				filename = name
				f.Close()
				break
			}
		}
	}

	if filename == "" {
		return "", nil
	}

	data, err := o.FS.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
