package store

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/n3wscott/bujo/pkg/entry"
	"github.com/peterbourgon/diskv/v3"
	"strings"
)

type Persistence interface {
	ListAll(ctx context.Context) []*entry.Entry
	List(ctx context.Context, collection string) []*entry.Entry
	Store(e *entry.Entry) error
}

func Load(cfg Config) (Persistence, error) {
	if cfg == nil {
		var err error
		cfg, err = LoadConfig()
		if err != nil {
			return nil, err
		}
	}

	return &persistence{d: diskv.New(diskv.Options{
		BasePath:          cfg.BasePath(),
		AdvancedTransform: keyToPathTransform,
		InverseTransform:  pathToKeyTransform,
		CacheSizeMax:      1024 * 1024, // 1MB
	})}, nil
}

type persistence struct {
	d *diskv.Diskv
}

func (p *persistence) read(key string) (*entry.Entry, error) {
	val, err := p.d.Read(key)
	if err != nil {
		return nil, err
	}
	e := entry.Entry{}
	if err := json.Unmarshal(val, &e); err != nil {
		return nil, err
	}
	return &e, nil
}
func (p *persistence) ListAll(ctx context.Context) []*entry.Entry {
	all := make([]*entry.Entry, 0)
	for key := range p.d.Keys(ctx.Done()) {
		e, err := p.read(key)
		if err != nil {
			fmt.Printf("%s: %s\n", key, err) // TODO: print this to STDERR
		}
		all = append(all, e)
	}
	// TODO: sort these based on ?
	return all
}

func (p *persistence) List(ctx context.Context, collection string) []*entry.Entry {
	ck := toCollection(collection)
	all := make([]*entry.Entry, 0)
	for key := range p.d.Keys(ctx.Done()) {
		if pk := keyToPathTransform(key); pk.Path[0] == ck {
			e, err := p.read(key)
			if err != nil {
				fmt.Printf("%s: %s\n", key, err) // TODO: print this to STDERR
			}
			all = append(all, e)
		}
	}
	// TODO: sort these based on created.
	// TODO: add a filter for done?
	return all
}

func (p *persistence) Store(e *entry.Entry) error {
	key := toKey(e)
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if err := p.d.Write(key, data); err != nil {
		return err
	}
	return nil
}

const (
	layoutISO = "2006-01-02"
)

func keyToPathTransform(s string) *diskv.PathKey {
	parts := strings.Split(s, "-")
	return &diskv.PathKey{
		Path:     parts[:len(parts)-1],
		FileName: parts[len(parts)-1],
	}
}

func pathToKeyTransform(pathKey *diskv.PathKey) string {
	return fmt.Sprintf("%s-%s", strings.Join(pathKey.Path, "-"), pathKey.FileName)
}

// toKey makes `collection-date-id`
func toKey(e *entry.Entry) string {
	collection := toCollection(e.Collection)
	then := e.Created.Time.Format(layoutISO)

	b, _ := json.Marshal(e)
	id := md5.Sum(b)

	return fmt.Sprintf("%s-%s-%x", collection, then, id[:8])
}

// toKey makes `collection-date-id`
func toCollection(s string) string {
	collection := base64.StdEncoding.EncodeToString([]byte(s))
	return fmt.Sprintf("%s", collection)
}
