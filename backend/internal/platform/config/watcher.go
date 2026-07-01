package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

// OnConfigChanged 是配置变更回调。返回 error 时记录日志但继续运行。
type OnConfigChanged func(newCfg *Config) error

// Watcher 监控配置文件变化并触发重载。
type Watcher struct {
	path    string
	onChange OnConfigChanged
	watcher *fsnotify.Watcher
	debounce time.Duration
}

// NewWatcher 创建配置热加载器。
// path 是配置文件路径（如 configs/config.local.yaml）。
// onChange 在文件变更且 debounce 到期后被调用。
func NewWatcher(path string, onChange OnConfigChanged, debounce time.Duration) (*Watcher, error) {
	if path == "" {
		return nil, fmt.Errorf("config path is required")
	}
	if onChange == nil {
		return nil, fmt.Errorf("onChange callback is required")
	}
	if debounce == 0 {
		debounce = 500 * time.Millisecond
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create fsnotify watcher: %w", err)
	}

	if err := watcher.Add(path); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("watch config file: %w", err)
	}

	return &Watcher{
		path:     path,
		onChange: onChange,
		watcher:  watcher,
		debounce: debounce,
	}, nil
}

// Start 启动监听循环。阻塞直到 ctx 取消或 Stop 被调用。
func (w *Watcher) Start(ctx context.Context) error {
	var timer *time.Timer
	for {
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return nil
		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}
			if event.Name != w.path {
				continue
			}
			if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Create == 0 {
				continue
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(w.debounce, func() {
				os.Setenv("APP_CONFIG_PATH", w.path)
				cfg, err := LoadConfig("")
				if err != nil {
					fmt.Fprintf(os.Stderr, "hot-reload config failed: %v\n", err)
					return
				}
				if err := w.onChange(cfg); err != nil {
					fmt.Fprintf(os.Stderr, "apply reloaded config failed: %v\n", err)
				}
			})
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "config watcher error: %v\n", err)
		}
	}
}

// Stop 停止监听并释放资源。
func (w *Watcher) Stop() error {
	return w.watcher.Close()
}
