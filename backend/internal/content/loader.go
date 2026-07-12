package content

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Loader struct {
	mu           sync.RWMutex
	dataDir      string
	index        *Index
	loaded       map[string]*Challenge
	packs        map[string]*ChallengePack
	onChange     func(*Challenge)
	lastScan     time.Time
	scanInterval time.Duration
}

func NewLoader(dataDir string) *Loader {
	return &Loader{
		dataDir:      dataDir,
		index:        NewIndex(),
		loaded:       make(map[string]*Challenge),
		packs:        make(map[string]*ChallengePack),
		scanInterval: 30 * time.Second,
	}
}

func (l *Loader) LoadAll() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.index = NewIndex()
	l.loaded = make(map[string]*Challenge)
	l.packs = make(map[string]*ChallengePack)

	entries, err := os.ReadDir(l.dataDir)
	if err != nil {
		return fmt.Errorf("read data dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if err := l.loadDirectory(filepath.Join(l.dataDir, entry.Name())); err != nil {
				log.Printf("Warning: skipping directory %s: %v", entry.Name(), err)
			}
		} else if strings.HasSuffix(entry.Name(), ".json") && entry.Name() != "pack.json" {
			if err := l.loadFile(filepath.Join(l.dataDir, entry.Name())); err != nil {
				log.Printf("Warning: skipping file %s: %v", entry.Name(), err)
			}
		}
	}

	l.lastScan = time.Now()
	log.Printf("Loaded %d challenges, %d packs", len(l.loaded), len(l.packs))
	return nil
}

func (l *Loader) loadDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if err := l.loadDirectory(filepath.Join(dir, entry.Name())); err != nil {
				log.Printf("Warning: skipping subdirectory %s: %v", entry.Name(), err)
			}
		} else if strings.HasSuffix(entry.Name(), ".json") {
			path := filepath.Join(dir, entry.Name())
			if entry.Name() == "pack.json" {
				if err := l.loadPack(path); err != nil {
					log.Printf("Warning: skipping pack %s: %v", path, err)
				}
			} else {
				if err := l.loadFile(path); err != nil {
					log.Printf("Warning: skipping challenge %s: %v", path, err)
				}
			}
		}
	}
	return nil
}

func (l *Loader) loadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	ch, errs := ValidateChallenge(data)
	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %s", errs.Error())
	}

	ch.Metadata["source_file"] = path

	l.loaded[ch.ID] = ch
	l.index.Add(ch)

	return nil
}

func (l *Loader) loadPack(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	pack, err := ValidatePack(data)
	if err != nil {
		return fmt.Errorf("validate pack: %w", err)
	}

	l.packs[pack.ID] = pack
	return nil
}

func (l *Loader) Get(id string) (*Challenge, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	ch, ok := l.loaded[id]
	return ch, ok
}

func (l *Loader) Index() *Index {
	return l.index
}

func (l *Loader) Packs() map[string]*ChallengePack {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.packs
}

func (l *Loader) Pack(id string) (*ChallengePack, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	pack, ok := l.packs[id]
	return pack, ok
}

func (l *Loader) Reload() error {
	return l.LoadAll()
}

func (l *Loader) Stats() LoaderStats {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return LoaderStats{
		TotalChallenges: len(l.loaded),
		TotalPacks:      len(l.packs),
		LastScan:        l.lastScan,
		Categories:      l.index.Categories(),
		Languages:       l.index.Languages(),
		Tags:            l.index.Tags(),
	}
}

type LoaderStats struct {
	TotalChallenges int
	TotalPacks      int
	LastScan        time.Time
	Categories      []string
	Languages       []string
	Tags            []string
}

func (l *Loader) Watch(callback func(*Challenge)) {
	l.onChange = callback
}

func (l *Loader) LoadFromJSON(data []byte) (*Challenge, error) {
	ch, errs := ValidateChallenge(data)
	if len(errs) > 0 {
		return ch, errs
	}

	l.mu.Lock()
	l.loaded[ch.ID] = ch
	l.index.Add(ch)
	l.mu.Unlock()

	if l.onChange != nil {
		l.onChange(ch)
	}

	return ch, nil
}

func (l *Loader) LoadPackFromJSON(data []byte) (*ChallengePack, error) {
	pack, err := ValidatePack(data)
	if err != nil {
		return nil, err
	}

	l.mu.Lock()
	l.packs[pack.ID] = pack
	l.mu.Unlock()

	return pack, nil
}

func (l *Loader) ExportChallenge(id string) ([]byte, error) {
	ch, ok := l.Get(id)
	if !ok {
		return nil, fmt.Errorf("challenge %s not found", id)
	}
	return json.MarshalIndent(ch, "", "  ")
}

func (l *Loader) ExportPack(id string) ([]byte, error) {
	pack, ok := l.Pack(id)
	if !ok {
		return nil, fmt.Errorf("pack %s not found", id)
	}
	return json.MarshalIndent(pack, "", "  ")
}
