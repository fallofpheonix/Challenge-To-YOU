package content

import (
	"sort"
	"strings"
	"sync"
)

type Index struct {
	mu           sync.RWMutex
	challenges   map[string]*Challenge
	byCategory   map[string][]string
	byDifficulty map[string][]string
	byLanguage   map[string][]string
	byTag        map[string][]string
	byEra        map[string][]string
}

func NewIndex() *Index {
	return &Index{
		challenges:   make(map[string]*Challenge),
		byCategory:   make(map[string][]string),
		byDifficulty: make(map[string][]string),
		byLanguage:   make(map[string][]string),
		byTag:        make(map[string][]string),
		byEra:        make(map[string][]string),
	}
}

func (idx *Index) Add(ch *Challenge) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.challenges[ch.ID] = ch

	idx.byCategory[ch.Category] = append(idx.byCategory[ch.Category], ch.ID)

	diffBucket := difficultyBucket(ch.Difficulty)
	idx.byDifficulty[diffBucket] = append(idx.byDifficulty[diffBucket], ch.ID)

	for _, lang := range ch.SupportedLanguages {
		idx.byLanguage[lang] = append(idx.byLanguage[lang], ch.ID)
	}

	for _, tag := range ch.Tags {
		idx.byTag[tag] = append(idx.byTag[tag], ch.ID)
	}

	if era, ok := ch.Metadata["era"].(string); ok {
		idx.byEra[era] = append(idx.byEra[era], ch.ID)
	}
}

func (idx *Index) Remove(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	ch, ok := idx.challenges[id]
	if !ok {
		return
	}

	delete(idx.challenges, id)

	idx.byCategory[ch.Category] = removeFromSlice(idx.byCategory[ch.Category], id)

	diffBucket := difficultyBucket(ch.Difficulty)
	idx.byDifficulty[diffBucket] = removeFromSlice(idx.byDifficulty[diffBucket], id)

	for _, lang := range ch.SupportedLanguages {
		idx.byLanguage[lang] = removeFromSlice(idx.byLanguage[lang], id)
	}

	for _, tag := range ch.Tags {
		idx.byTag[tag] = removeFromSlice(idx.byTag[tag], id)
	}

	if era, ok := ch.Metadata["era"].(string); ok {
		idx.byEra[era] = removeFromSlice(idx.byEra[era], id)
	}
}

func (idx *Index) Get(id string) (*Challenge, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	ch, ok := idx.challenges[id]
	return ch, ok
}

func (idx *Index) List() []*Challenge {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	result := make([]*Challenge, 0, len(idx.challenges))
	for _, ch := range idx.challenges {
		result = append(result, ch)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

func (idx *Index) ByCategory(category string) []*Challenge {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	ids := idx.byCategory[category]
	return idx.getByIDs(ids)
}

func (idx *Index) ByDifficulty(min, max float64) []*Challenge {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	var result []*Challenge
	for _, ch := range idx.challenges {
		if ch.Difficulty >= min && ch.Difficulty <= max {
			result = append(result, ch)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Difficulty < result[j].Difficulty
	})
	return result
}

func (idx *Index) ByLanguage(lang string) []*Challenge {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	ids := idx.byLanguage[lang]
	return idx.getByIDs(ids)
}

func (idx *Index) ByTag(tag string) []*Challenge {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	ids := idx.byTag[tag]
	return idx.getByIDs(ids)
}

func (idx *Index) ByEra(era string) []*Challenge {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	ids := idx.byEra[era]
	return idx.getByIDs(ids)
}

func (idx *Index) Search(query string) []*Challenge {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	query = lowercase(query)
	var result []*Challenge
	for _, ch := range idx.challenges {
		if contains(ch.ID, query) || contains(ch.Title, query) || contains(ch.Description, query) || contains(ch.Category, query) {
			result = append(result, ch)
		}
	}
	return result
}

func (idx *Index) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.challenges)
}

func (idx *Index) Categories() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	cats := make([]string, 0, len(idx.byCategory))
	for cat := range idx.byCategory {
		cats = append(cats, cat)
	}
	sort.Strings(cats)
	return cats
}

func (idx *Index) Languages() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	langs := make([]string, 0, len(idx.byLanguage))
	for lang := range idx.byLanguage {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}

func (idx *Index) Tags() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	tags := make([]string, 0, len(idx.byTag))
	for tag := range idx.byTag {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func (idx *Index) getByIDs(ids []string) []*Challenge {
	result := make([]*Challenge, 0, len(ids))
	for _, id := range ids {
		if ch, ok := idx.challenges[id]; ok {
			result = append(result, ch)
		}
	}
	return result
}

func difficultyBucket(d float64) string {
	switch {
	case d < 0.2:
		return "tutorial"
	case d < 0.4:
		return "beginner"
	case d < 0.6:
		return "easy"
	case d < 0.7:
		return "medium"
	case d < 0.85:
		return "hard"
	default:
		return "expert"
	}
}

func removeFromSlice(slice []string, val string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != val {
			result = append(result, s)
		}
	}
	return result
}

func lowercase(s string) string {
	return strings.ToLower(s)
}

func contains(s, substr string) bool {
	return strings.Contains(lowercase(s), substr)
}
