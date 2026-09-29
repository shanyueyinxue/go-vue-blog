package scheduler

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-co-op/gocron/v2"
	"github.com/robfig/cron/v3" // 仅用于 cron 表达式校验 
)

type Task struct {
	ScanInterval string // 扫描间隔（Cron表达式，含秒字段）
	TaskFunc     func() // 任务函数
	Name         string // 任务名称

	scheduled bool       // 任务是否已添加到调度器
	job       gocron.Job // gocron 中的 job 对象
}

func NewTask(name string, scanInterval string, taskFunc func()) *Task {
	return &Task{
		ScanInterval: scanInterval,
		TaskFunc:     taskFunc,
		Name:         name,
	}
}

type Scheduler struct {
	scheduler gocron.Scheduler // gocron 调度器实例
	tasks     []*Task          // 任务列表
	isStart   bool             // 调度器是否已启动
	mu        sync.RWMutex
}

// NewScheduler 创建调度器，可传入一组初始任务
func NewScheduler(tasks ...*Task) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create gocron scheduler: %w", err)
	}
	return &Scheduler{
		scheduler: s,
		tasks:     tasks,
		isStart:   false,
	}, nil
}

var (
	ErrTaskIsRunning      = errors.New("task is running")
	ErrScanIntervalFormat = errors.New("scan interval format error")
	ErrTaskFuncNil        = errors.New("task function is nil")
	ErrTaskNameEmpty      = errors.New("task name is empty")
	ErrTaskNameDuplicate  = errors.New("task name duplicate")
)

// runTaskLocked 将具体任务添加到 gocron 调度器中，需持有锁
func (s *Scheduler) runTaskLocked(task *Task) error {
	if task.scheduled {
		return ErrTaskIsRunning
	}
	// 使用 CronJob，第二个参数 true 表示表达式包含秒字段（6字段）
	job, err := s.scheduler.NewJob(
		gocron.CronJob(task.ScanInterval, true),
		gocron.NewTask(task.TaskFunc),
		gocron.WithName(task.Name),
	)
	if err != nil {
		return fmt.Errorf("add job failed: %w", err)
	}
	task.job = job
	task.scheduled = true
	return nil
}

// Start 启动调度器，同时将尚未调度的任务全部添加
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isStart {
		return nil // 已启动
	}

	if len(s.tasks) == 0 {
		fmt.Println("warning: task list is empty")
	}

	// 遍历所有任务，将未调度的添加到 gocron
	for _, task := range s.tasks {
		if task.scheduled {
			continue
		}
		err := s.runTaskLocked(task)
		if err != nil && err != ErrTaskIsRunning {
			// 回滚：移除所有已成功添加的任务
			for _, t := range s.tasks {
				if t.scheduled && t.job != nil {
					_ = s.scheduler.RemoveJob(t.job.ID())
					t.scheduled = false
					t.job = nil
				}
			}
			return err
		}
	}
	s.scheduler.Start()
	s.isStart = true
	return nil
}

// Stop 停止调度器，释放资源
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.scheduler == nil {
		return
	}
	_ = s.scheduler.Shutdown() // Shutdown 会等待正在执行的任务完成
	s.isStart = false
}

// AddTask 动态添加新任务（若调度器已启动，任务会立即开始调度）
func (s *Scheduler) AddTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.ValidateTask(task)
	if err != nil {
		return err
	}

	// 如果调度器已经启动，立即将任务加入调度
	if s.isStart {
		if err := s.runTaskLocked(task); err != nil {
			return err
		}
	}

	s.tasks = append(s.tasks, task)
	return nil
}

// RemoveTask 根据任务名称移除任务
func (s *Scheduler) RemoveTask(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newTasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if task.Name == name {
			if task.scheduled && task.job != nil {
				_ = s.scheduler.RemoveJob(task.job.ID())
				task.scheduled = false
				task.job = nil
			}
			continue
		}
		newTasks = append(newTasks, task)
	}
	s.tasks = newTasks
}

// ValidateTask 校验任务参数是否合法
func (s *Scheduler) ValidateTask(task *Task) error {
	if task.Name == "" {
		return ErrTaskNameEmpty
	}
	// 检查名称重复
	for _, t := range s.tasks {
		if t.Name == task.Name {
			return ErrTaskNameDuplicate
		}
	}
	if task.ScanInterval == "" {
		return ErrScanIntervalFormat
	}
	// 校验 cron 表达式格式（沿用 robfig/cron 解析器，保证与 gocron 行为一致）
	err := s.parseStandard(task.ScanInterval)
	if err != nil {
		return err
	}
	if task.TaskFunc == nil {
		return ErrTaskFuncNil
	}
	return nil
}

// parseStandard 校验带秒字段的 Cron 表达式（6字段）
func (s *Scheduler) parseStandard(interval string) error {
	p := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	_, err := p.Parse(interval)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrScanIntervalFormat, err)
	}
	return nil
}

// ValidateStandard 对外暴露的表达式校验工具方法
func (s *Scheduler) ValidateStandard(interval string) bool {
	err := s.parseStandard(interval)
	return err == nil
}
