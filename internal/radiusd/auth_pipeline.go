package radiusd

import (
	"context"
	"fmt"
	"sync"

	"github.com/talkincode/toughradius/v9/internal/domain"
	vendorparsers "github.com/talkincode/toughradius/v9/internal/radiusd/plugins/vendorparsers"
	"layeh.com/radius"
)

// AuthPipelineStage represents a pluggable processing unit inside ServeRADIUS.
type AuthPipelineStage interface {
	Name() string
	Execute(ctx *AuthPipelineContext) error
}

type stageFunc struct {
	name string
	fn   func(ctx *AuthPipelineContext) error
}

func (s *stageFunc) Name() string {
	return s.name
}

func (s *stageFunc) Execute(ctx *AuthPipelineContext) error {
	return s.fn(ctx)
}

func newStage(name string, fn func(ctx *AuthPipelineContext) error) AuthPipelineStage {
	return &stageFunc{name: name, fn: fn}
}

// AuthPipeline manages ordered stage execution.
type AuthPipeline struct {
	mu     sync.RWMutex
	stages []AuthPipelineStage
}

// NewAuthPipeline creates an empty pipeline instance.
func NewAuthPipeline() *AuthPipeline {
	return &AuthPipeline{stages: make([]AuthPipelineStage, 0)}
}

// Use appends a stage to the end of the pipeline.
func (p *AuthPipeline) Use(stage AuthPipelineStage) *AuthPipeline {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stages = append(p.stages, stage)
	return p
}

// InsertBefore inserts a stage before the target stage name.
func (p *AuthPipeline) InsertBefore(target string, stage AuthPipelineStage) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	idx := p.indexOf(target)
	if idx == -1 {
		return fmt.Errorf("stage %s not found", target)
	}
	p.stages = append(p.stages[:idx], append([]AuthPipelineStage{stage}, p.stages[idx:]...)...)
	return nil
}

// InsertAfter inserts a stage after the target stage name.
func (p *AuthPipeline) InsertAfter(target string, stage AuthPipelineStage) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	idx := p.indexOf(target)
	if idx == -1 {
		return fmt.Errorf("stage %s not found", target)
	}
	insertPos := idx + 1
	p.stages = append(p.stages[:insertPos], append([]AuthPipelineStage{stage}, p.stages[insertPos:]...)...)
	return nil
}

// Replace swaps the stage with the provided implementation.
func (p *AuthPipeline) Replace(target string, stage AuthPipelineStage) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	idx := p.indexOf(target)
	if idx == -1 {
		return fmt.Errorf("stage %s not found", target)
	}
	p.stages[idx] = stage
	return nil
}

// Remove deletes a stage from the pipeline.
func (p *AuthPipeline) Remove(target string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	idx := p.indexOf(target)
	if idx == -1 {
		return fmt.Errorf("stage %s not found", target)
	}
	p.stages = append(p.stages[:idx], p.stages[idx+1:]...)
	return nil
}

// Stages returns a copy of the registered stages.
func (p *AuthPipeline) Stages() []AuthPipelineStage {
	p.mu.RLock()
	defer p.mu.RUnlock()
	stages := make([]AuthPipelineStage, len(p.stages))
	copy(stages, p.stages)
	return stages
}

// Execute runs stages sequentially until completion or ctx.Stop() is invoked.
func (p *AuthPipeline) Execute(ctx *AuthPipelineContext) error {
	p.mu.RLock()
	stages := make([]AuthPipelineStage, len(p.stages))
	copy(stages, p.stages)
	p.mu.RUnlock()

	for _, stage := range stages {
		if ctx.IsStopped() {
			break
		}
		if err := stage.Execute(ctx); err != nil {
			return fmt.Errorf("stage %s failed: %w", stage.Name(), err)
		}
	}
	return nil
}

func (p *AuthPipeline) indexOf(name string) int {
	for idx, stage := range p.stages {
		if stage.Name() == name {
			return idx
		}
	}
	return -1
}

// AuthPipelineContext carries per-request mutable data across stages.
type AuthPipelineContext struct {
	Context context.Context
	Service *AuthService

	Writer   radius.ResponseWriter
	Request  *radius.Request
	Response *radius.Packet

	Username         string
	NasIdentifier    string
	CallingStationID string
	RemoteIP         string

	NAS                    *domain.NetNas
	VendorRequest          *VendorRequest
	VendorRequestForPlugin *vendorparsers.VendorRequest
	User                   *domain.RadiusUser

	IsEAP            bool
	EAPMethod        string
	IsMacAuth        bool
	RateLimitChecked bool

	stop bool
}

// NewAuthPipelineContext builds a context with sane defaults.
func NewAuthPipelineContext(service *AuthService, w radius.ResponseWriter, r *radius.Request) *AuthPipelineContext {
	return &AuthPipelineContext{
		Context:                context.Background(),
		Service:                service,
		Writer:                 w,
		Request:                r,
		VendorRequest:          &VendorRequest{},
		VendorRequestForPlugin: &vendorparsers.VendorRequest{},
	}
}

// Stop halts further stage execution.
func (ctx *AuthPipelineContext) Stop() {
	ctx.stop = true
}

// IsStopped reports whether execution has been halted.
func (ctx *AuthPipelineContext) IsStopped() bool {
	return ctx.stop
}
