package compiler

import (
	"context"
	"fmt"
	"sync"
)

type Manager struct {
	mu        sync.RWMutex
	languages map[string]*Language
	executors map[string]Executor
}

func NewManager() *Manager {
	return &Manager{
		languages: make(map[string]*Language),
		executors: make(map[string]Executor),
	}
}

func (m *Manager) RegisterLanguage(lang *Language, executor Executor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.languages[lang.ID] = lang
	m.executors[lang.ID] = executor
}

func (m *Manager) GetLanguage(id string) (*Language, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	lang, ok := m.languages[id]
	if !ok {
		return nil, fmt.Errorf("language %q not registered", id)
	}
	return lang, nil
}

func (m *Manager) GetExecutor(langID string) (Executor, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	exec, ok := m.executors[langID]
	if !ok {
		return nil, fmt.Errorf("executor for language %q not registered", langID)
	}
	return exec, nil
}

func (m *Manager) ListLanguages() []*Language {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Language, 0, len(m.languages))
	for _, lang := range m.languages {
		result = append(result, lang)
	}
	return result
}

func (m *Manager) Compile(ctx context.Context, code string, langID string) (*CompilationResult, error) {
	lang, err := m.GetLanguage(langID)
	if err != nil {
		return nil, err
	}
	exec, err := m.GetExecutor(langID)
	if err != nil {
		return nil, err
	}
	return exec.Compile(ctx, code, lang)
}

func (m *Manager) Execute(ctx context.Context, code string, input string, langID string) (*ExecutionResult, error) {
	lang, err := m.GetLanguage(langID)
	if err != nil {
		return nil, err
	}
	exec, err := m.GetExecutor(langID)
	if err != nil {
		return nil, err
	}
	return exec.Execute(ctx, code, input, lang)
}
