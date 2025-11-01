package walkcache

import (
	"io/fs"
	"os"
)

type proxy struct {
	Cache
}

func (p *proxy) ReadFile(path string) ([]byte, error) {
	data, info, ok := p.GetFile(path)
	if ok {
		return data, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p.SetFile(path, data, info)
	return data, nil
}

func (p *proxy) ReadDir(path string) ([]fs.DirEntry, error) {
	data, ok := p.GetDir(path)
	if ok {
		return data, nil
	}
	data, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	p.SetDir(path, data)
	return data, nil
}

func (p *proxy) Stat(path string) (fs.FileInfo, error) {
	data, info, ok := p.GetFile(path)
	if ok {
		return info, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	p.SetFile(path, data, info)
	return info, nil
}
