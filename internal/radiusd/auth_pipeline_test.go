package radiusd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStage is a simple test implementation of AuthPipelineStage
type mockStage struct {
	name     string
	executed bool
	err      error
}

func newMockStage(name string) *mockStage {
	return &mockStage{name: name}
}

func newMockStageWithError(name string, err error) *mockStage {
	return &mockStage{name: name, err: err}
}

func (s *mockStage) Name() string {
	return s.name
}

func (s *mockStage) Execute(ctx *AuthPipelineContext) error {
	s.executed = true
	return s.err
}

func TestNewAuthPipeline(t *testing.T) {
	p := NewAuthPipeline()
	assert.NotNil(t, p)
	assert.Empty(t, p.Stages())
}

func TestAuthPipeline_Use(t *testing.T) {
	p := NewAuthPipeline()

	stage1 := newMockStage("stage1")
	stage2 := newMockStage("stage2")
	stage3 := newMockStage("stage3")

	p.Use(stage1).Use(stage2).Use(stage3)

	stages := p.Stages()
	require.Len(t, stages, 3)
	assert.Equal(t, "stage1", stages[0].Name())
	assert.Equal(t, "stage2", stages[1].Name())
	assert.Equal(t, "stage3", stages[2].Name())
}

func TestAuthPipeline_InsertBefore(t *testing.T) {
	t.Run("insert before existing stage", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("first")).Use(newMockStage("third"))

		err := p.InsertBefore("third", newMockStage("second"))
		assert.NoError(t, err)

		stages := p.Stages()
		require.Len(t, stages, 3)
		assert.Equal(t, "first", stages[0].Name())
		assert.Equal(t, "second", stages[1].Name())
		assert.Equal(t, "third", stages[2].Name())
	})

	t.Run("insert before first stage", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("second")).Use(newMockStage("third"))

		err := p.InsertBefore("second", newMockStage("first"))
		assert.NoError(t, err)

		stages := p.Stages()
		require.Len(t, stages, 3)
		assert.Equal(t, "first", stages[0].Name())
		assert.Equal(t, "second", stages[1].Name())
		assert.Equal(t, "third", stages[2].Name())
	})

	t.Run("error when target not found", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("first"))

		err := p.InsertBefore("nonexistent", newMockStage("new"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nonexistent")
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestAuthPipeline_InsertAfter(t *testing.T) {
	t.Run("insert after existing stage", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("first")).Use(newMockStage("third"))

		err := p.InsertAfter("first", newMockStage("second"))
		assert.NoError(t, err)

		stages := p.Stages()
		require.Len(t, stages, 3)
		assert.Equal(t, "first", stages[0].Name())
		assert.Equal(t, "second", stages[1].Name())
		assert.Equal(t, "third", stages[2].Name())
	})

	t.Run("insert after last stage", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("first")).Use(newMockStage("second"))

		err := p.InsertAfter("second", newMockStage("third"))
		assert.NoError(t, err)

		stages := p.Stages()
		require.Len(t, stages, 3)
		assert.Equal(t, "first", stages[0].Name())
		assert.Equal(t, "second", stages[1].Name())
		assert.Equal(t, "third", stages[2].Name())
	})

	t.Run("error when target not found", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("first"))

		err := p.InsertAfter("nonexistent", newMockStage("new"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nonexistent")
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestAuthPipeline_Replace(t *testing.T) {
	t.Run("replace existing stage", func(t *testing.T) {
		p := NewAuthPipeline()
		original := newMockStage("target")
		replacement := newMockStage("target") // Same name, different instance
		p.Use(newMockStage("first")).Use(original).Use(newMockStage("third"))

		err := p.Replace("target", replacement)
		assert.NoError(t, err)

		stages := p.Stages()
		require.Len(t, stages, 3)
		assert.Equal(t, "first", stages[0].Name())
		assert.Equal(t, "target", stages[1].Name())
		assert.Equal(t, "third", stages[2].Name())
		// Verify it's the replacement, not original
		assert.Same(t, replacement, stages[1])
	})

	t.Run("error when target not found", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("first"))

		err := p.Replace("nonexistent", newMockStage("replacement"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nonexistent")
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestAuthPipeline_Remove(t *testing.T) {
	t.Run("remove middle stage", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("first")).Use(newMockStage("second")).Use(newMockStage("third"))

		err := p.Remove("second")
		assert.NoError(t, err)

		stages := p.Stages()
		require.Len(t, stages, 2)
		assert.Equal(t, "first", stages[0].Name())
		assert.Equal(t, "third", stages[1].Name())
	})

	t.Run("remove first stage", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("first")).Use(newMockStage("second"))

		err := p.Remove("first")
		assert.NoError(t, err)

		stages := p.Stages()
		require.Len(t, stages, 1)
		assert.Equal(t, "second", stages[0].Name())
	})

	t.Run("remove last stage", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("first")).Use(newMockStage("second"))

		err := p.Remove("second")
		assert.NoError(t, err)

		stages := p.Stages()
		require.Len(t, stages, 1)
		assert.Equal(t, "first", stages[0].Name())
	})

	t.Run("error when target not found", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("first"))

		err := p.Remove("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nonexistent")
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestAuthPipeline_Stages(t *testing.T) {
	t.Run("returns copy of stages", func(t *testing.T) {
		p := NewAuthPipeline()
		p.Use(newMockStage("stage1")).Use(newMockStage("stage2"))

		stages1 := p.Stages()
		stages2 := p.Stages()

		// Should be equal content
		assert.Equal(t, len(stages1), len(stages2))

		// Modifying the returned slice should not affect the original
		stages1[0] = newMockStage("modified")
		assert.NotEqual(t, stages1[0].Name(), p.Stages()[0].Name())
	})

	t.Run("empty pipeline returns empty slice", func(t *testing.T) {
		p := NewAuthPipeline()
		stages := p.Stages()
		assert.Empty(t, stages)
		assert.NotNil(t, stages)
	})
}

func TestAuthPipeline_Execute(t *testing.T) {
	t.Run("executes all stages in order", func(t *testing.T) {
		p := NewAuthPipeline()
		stage1 := newMockStage("stage1")
		stage2 := newMockStage("stage2")
		stage3 := newMockStage("stage3")
		p.Use(stage1).Use(stage2).Use(stage3)

		ctx := &AuthPipelineContext{}
		err := p.Execute(ctx)

		assert.NoError(t, err)
		assert.True(t, stage1.executed)
		assert.True(t, stage2.executed)
		assert.True(t, stage3.executed)
	})

	t.Run("stops execution when ctx.Stop() called", func(t *testing.T) {
		p := NewAuthPipeline()
		stage1 := newMockStage("stage1")
		stageStop := &mockStage{
			name: "stopper",
		}
		// Override Execute to call Stop
		stage2 := &stoppingStage{name: "stopper"}
		stage3 := newMockStage("stage3")
		p.Use(stage1).Use(stage2).Use(stage3)

		ctx := &AuthPipelineContext{}
		err := p.Execute(ctx)

		assert.NoError(t, err)
		assert.True(t, stage1.executed)
		assert.True(t, stage2.executed)
		assert.False(t, stage3.executed)
		_ = stageStop // silence unused warning
	})

	t.Run("returns error when stage fails", func(t *testing.T) {
		p := NewAuthPipeline()
		stage1 := newMockStage("stage1")
		stage2 := newMockStageWithError("failing", errors.New("stage error"))
		stage3 := newMockStage("stage3")
		p.Use(stage1).Use(stage2).Use(stage3)

		ctx := &AuthPipelineContext{}
		err := p.Execute(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failing")
		assert.Contains(t, err.Error(), "stage error")
		assert.True(t, stage1.executed)
		assert.True(t, stage2.executed)
		assert.False(t, stage3.executed) // Should not execute after error
	})

	t.Run("empty pipeline executes without error", func(t *testing.T) {
		p := NewAuthPipeline()
		ctx := &AuthPipelineContext{}

		err := p.Execute(ctx)
		assert.NoError(t, err)
	})
}

// stoppingStage is a stage that stops pipeline execution
type stoppingStage struct {
	name     string
	executed bool
}

func (s *stoppingStage) Name() string {
	return s.name
}

func (s *stoppingStage) Execute(ctx *AuthPipelineContext) error {
	s.executed = true
	ctx.Stop()
	return nil
}

func TestStageFunc(t *testing.T) {
	t.Run("newStage creates functional stage", func(t *testing.T) {
		executed := false
		stage := newStage("test-stage", func(ctx *AuthPipelineContext) error {
			executed = true
			return nil
		})

		assert.Equal(t, "test-stage", stage.Name())

		ctx := &AuthPipelineContext{}
		err := stage.Execute(ctx)
		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("newStage returns error from function", func(t *testing.T) {
		expectedErr := errors.New("function error")
		stage := newStage("error-stage", func(ctx *AuthPipelineContext) error {
			return expectedErr
		})

		ctx := &AuthPipelineContext{}
		err := stage.Execute(ctx)
		assert.Equal(t, expectedErr, err)
	})
}

func TestAuthPipelineContext(t *testing.T) {
	t.Run("NewAuthPipelineContext creates context with defaults", func(t *testing.T) {
		ctx := NewAuthPipelineContext(nil, nil, nil)
		assert.NotNil(t, ctx)
		assert.NotNil(t, ctx.Context)
		assert.NotNil(t, ctx.VendorRequest)
		assert.NotNil(t, ctx.VendorRequestForPlugin)
		assert.False(t, ctx.IsStopped())
	})

	t.Run("Stop and IsStopped work correctly", func(t *testing.T) {
		ctx := &AuthPipelineContext{}
		assert.False(t, ctx.IsStopped())

		ctx.Stop()
		assert.True(t, ctx.IsStopped())
	})
}

func TestAuthPipeline_IndexOf(t *testing.T) {
	p := NewAuthPipeline()
	p.Use(newMockStage("first")).Use(newMockStage("second")).Use(newMockStage("third"))

	// Test via InsertBefore/InsertAfter/Replace/Remove which internally use indexOf
	// indexOf is private, so we test it indirectly

	// Insert before first should work
	err := p.InsertBefore("first", newMockStage("zeroth"))
	assert.NoError(t, err)
	assert.Equal(t, "zeroth", p.Stages()[0].Name())

	// Insert after third should work
	err = p.InsertAfter("third", newMockStage("fourth"))
	assert.NoError(t, err)
	stages := p.Stages()
	assert.Equal(t, "fourth", stages[len(stages)-1].Name())
}

func TestAuthPipeline_ConcurrentAccess(t *testing.T) {
	p := NewAuthPipeline()
	p.Use(newMockStage("initial"))

	done := make(chan bool, 100)

	// Concurrent reads
	for i := 0; i < 50; i++ {
		go func() {
			_ = p.Stages()
			done <- true
		}()
	}

	// Concurrent writes
	for i := 0; i < 50; i++ {
		go func(n int) {
			_ = p.Use(newMockStage("concurrent"))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Should not panic and have stages added
	stages := p.Stages()
	assert.Greater(t, len(stages), 1)
}
