package services

import (
	"container/list"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/tracewayapp/traceway/backend/app/cache"
	"github.com/tracewayapp/traceway/backend/app/models"
	"github.com/tracewayapp/traceway/backend/app/storage"

	"github.com/go-sourcemap/sourcemap"
	"github.com/google/uuid"
	traceway "go.tracewayapp.com"
)

// parsedSourceMapCache holds already-parsed *sourcemap.Consumer values so
// concurrent stack-trace resolutions can reuse one parse. A raw 20 MB
// source map balloons to ~100-200 MB once parsed; re-parsing on every
// stack frame (or every request) was the OOM cause on the 1 GB t4g.micro.
//
// Consumers are safe for concurrent reads per go-sourcemap docs.
type parsedSourceMapCache struct {
	mu    sync.Mutex
	items map[string]*list.Element
	order *list.List
	max   int
}

type parsedSourceMapEntry struct {
	key      string
	consumer *sourcemap.Consumer
}

var parsedSourceMaps = &parsedSourceMapCache{
	items: make(map[string]*list.Element),
	order: list.New(),
	max:   5,
}

func (c *parsedSourceMapCache) get(key string) (*sourcemap.Consumer, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		return el.Value.(*parsedSourceMapEntry).consumer, true
	}
	return nil, false
}

func (c *parsedSourceMapCache) put(key string, consumer *sourcemap.Consumer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		el.Value.(*parsedSourceMapEntry).consumer = consumer
		c.order.MoveToFront(el)
		return
	}
	el := c.order.PushFront(&parsedSourceMapEntry{key: key, consumer: consumer})
	c.items[key] = el
	for c.order.Len() > c.max {
		back := c.order.Back()
		if back == nil {
			break
		}
		evicted := c.order.Remove(back).(*parsedSourceMapEntry)
		delete(c.items, evicted.key)
	}
}

func ParsedSourceMapStats() (items int, max int) {
	parsedSourceMaps.mu.Lock()
	defer parsedSourceMaps.mu.Unlock()
	return parsedSourceMaps.order.Len(), parsedSourceMaps.max
}

var stackFrameRe = regexp.MustCompile(`^(\s{4})(.+):(\d+):(\d+)$`)
var jsFuncDeclRe = regexp.MustCompile(
	`(?:(?:export\s+(?:default\s+)?)?function\s+(\w+)` +
		`|(?:const|let|var)\s+(\w+)\s*=` +
		`|^\s*(?:async\s+)?(\w+)\s*\([^)]*\)\s*\{)`,
)

var jsControlFlowKeywords = map[string]bool{
	"if": true, "for": true, "while": true, "switch": true,
	"catch": true, "return": true, "throw": true, "else": true,
}

func ResolveStackTrace(ctx context.Context, projectId uuid.UUID, stackTrace string, sourceMaps []*models.SourceMap) string {
	if len(sourceMaps) == 0 {
		return stackTrace
	}

	// Source-map resolution is opt-in. A 20+ MB minified map parses into
	// ~100-200 MB of in-memory state; under concurrent load this was pushing
	// a 1 GB EC2 instance over the OOM line. Default: off. Flip on with
	// TRACEWAY_SOURCEMAP_RESOLUTION=on in the environment when the host has
	// enough RAM.
	if os.Getenv("TRACEWAY_SOURCEMAP_RESOLUTION") != "on" {
		return stackTrace
	}

	// Build lookup: basename of source map's file_name (without .map) -> source map
	// Also keep a map of file_name -> storage_key for direct lookup
	smByBasename := make(map[string]*models.SourceMap)
	for _, sm := range sourceMaps {
		smByBasename[sm.FileName] = sm
		// Also index by basename
		base := filepath.Base(sm.FileName)
		smByBasename[base] = sm
	}

	lines := strings.Split(stackTrace, "\n")
	resolved := make([]string, 0, len(lines))
	framesResolved := 0
	maxFrames := 50

	// Per-call consumer map — a given source map is parsed at most once per
	// ResolveStackTrace invocation even if not in the global cache.
	localConsumers := make(map[string]*sourcemap.Consumer)

	for _, line := range lines {
		if framesResolved >= maxFrames {
			resolved = append(resolved, line)
			continue
		}

		matches := stackFrameRe.FindStringSubmatch(line)
		if matches == nil {
			resolved = append(resolved, line)
			continue
		}

		indent := matches[1]
		fileName := matches[2]
		lineNum, _ := strconv.Atoi(matches[3])
		colNum, _ := strconv.Atoi(matches[4])

		sm := findSourceMap(fileName, smByBasename)
		if sm == nil {
			resolved = append(resolved, line)
			continue
		}

		consumer, err := getParsedSourceMap(ctx, sm.StorageKey, localConsumers)
		if err != nil || consumer == nil {
			resolved = append(resolved, line)
			continue
		}

		origFile, origName, origLine, origCol, ok := consumer.Source(lineNum, colNum)
		if !ok || origFile == "" {
			resolved = append(resolved, line)
			continue
		}

		if content := consumer.SourceContent(origFile); content != "" {
			if extracted := extractFunctionName(content, origLine); extracted != "" {
				origName = extracted
			}
		}

		resolved = append(resolved, fmt.Sprintf("%s%s:%d:%d", indent, origFile, origLine, origCol))
		framesResolved++

		if origName != "" && len(resolved) >= 2 {
			prev := resolved[len(resolved)-2]
			if strings.HasSuffix(strings.TrimSpace(prev), "()") {
				trimmed := strings.TrimSpace(prev)
				indent := prev[:len(prev)-len(trimmed)]
				resolved[len(resolved)-2] = indent + origName + "()"
			}
		}
	}

	return strings.Join(resolved, "\n")
}

func findSourceMap(stackFile string, smByBasename map[string]*models.SourceMap) *models.SourceMap {
	// Try file.map directly
	mapName := stackFile + ".map"
	if sm, ok := smByBasename[mapName]; ok {
		return sm
	}

	// Try basename.map
	base := filepath.Base(stackFile) + ".map"
	if sm, ok := smByBasename[base]; ok {
		return sm
	}

	// Try without query params
	cleanName := stackFile
	if idx := strings.IndexAny(cleanName, "?#"); idx != -1 {
		cleanName = cleanName[:idx]
	}
	mapName = filepath.Base(cleanName) + ".map"
	if sm, ok := smByBasename[mapName]; ok {
		return sm
	}

	return nil
}

// getParsedSourceMap returns a reusable *sourcemap.Consumer for the given
// storage key, parsing at most once per unique source map. Lookup order:
//
//  1. Per-call localConsumers (same ResolveStackTrace invocation)
//  2. Global parsedSourceMaps LRU (cross-request reuse, capped at 5 entries)
//  3. Raw-bytes cache, then storage.Store.Read — then parse and populate.
//
// The raw-bytes cache is still populated so the parse only pays one storage
// read even when the parsed cache evicts us.
func getParsedSourceMap(ctx context.Context, storageKey string, localConsumers map[string]*sourcemap.Consumer) (*sourcemap.Consumer, error) {
	if c, ok := localConsumers[storageKey]; ok {
		return c, nil
	}
	if c, ok := parsedSourceMaps.get(storageKey); ok {
		localConsumers[storageKey] = c
		return c, nil
	}

	data, ok := cache.SourceMapCache.Get(storageKey)
	if !ok {
		var err error
		data, err = storage.Store.Read(ctx, storageKey)
		if err != nil {
			traceway.CaptureException(fmt.Errorf("failed to read source map from storage (key=%s): %w", storageKey, err))
			return nil, err
		}
		cache.SourceMapCache.Put(storageKey, data)
	}

	consumer, err := sourcemap.Parse("", data)
	if err != nil {
		return nil, err
	}
	parsedSourceMaps.put(storageKey, consumer)
	localConsumers[storageKey] = consumer
	return consumer, nil
}

func extractFunctionName(sourceContent string, line int) string {
	lines := strings.Split(sourceContent, "\n")
	for i := line - 1; i >= 0 && i >= line-50; i-- {
		matches := jsFuncDeclRe.FindStringSubmatch(lines[i])
		if matches != nil {
			for _, m := range matches[1:] {
				if m != "" && !jsControlFlowKeywords[m] {
					return m
				}
			}
		}
	}
	return ""
}
