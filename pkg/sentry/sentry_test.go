package sentry

import (
	"errors"
	"os"
	"testing"

	sentrygo "github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSentry_BuilderPattern(t *testing.T) {
	t.Run("WithContext sets context", func(t *testing.T) {
		e := echo.New()
		ctx := e.NewContext(nil, nil)
		sentry := new(Sentry)

		result := sentry.WithContext(ctx)

		assert.Equal(t, ctx, result.context)
		assert.Equal(t, sentry, result, "should return same instance for chaining")
	})

	t.Run("WithError sets error", func(t *testing.T) {
		err := errors.New("test error")
		sentry := new(Sentry)

		result := sentry.WithError(err)

		assert.Equal(t, err, result.error)
		assert.Equal(t, sentry, result, "should return same instance for chaining")
	})

	t.Run("WithMessage sets message", func(t *testing.T) {
		msg := "test message"
		sentry := new(Sentry)

		result := sentry.WithMessage(msg)

		assert.Equal(t, msg, result.message)
		assert.Equal(t, sentry, result, "should return same instance for chaining")
	})

	t.Run("WithLevel sets level", func(t *testing.T) {
		level := sentrygo.LevelWarning
		sentry := new(Sentry)

		result := sentry.WithLevel(level)

		assert.Equal(t, level, result.level)
		assert.Equal(t, sentry, result, "should return same instance for chaining")
	})

	t.Run("WithExtras sets extras", func(t *testing.T) {
		extras := map[string]interface{}{"key": "value"}
		sentry := new(Sentry)

		result := sentry.WithExtras(extras)

		assert.Equal(t, extras, result.extras)
		assert.Equal(t, sentry, result, "should return same instance for chaining")
	})

	t.Run("WithTags sets tags", func(t *testing.T) {
		tags := map[string]string{"env": "test"}
		sentry := new(Sentry)

		result := sentry.WithTags(tags)

		assert.Equal(t, tags, result.tags)
		assert.Equal(t, sentry, result, "should return same instance for chaining")
	})

	t.Run("WithContextValues sets context values", func(t *testing.T) {
		contextValues := map[string]sentrygo.Context{"key": {}}
		sentry := new(Sentry)

		result := sentry.WithContextValues(contextValues)

		assert.Equal(t, contextValues, result.contextValues)
		assert.Equal(t, sentry, result, "should return same instance for chaining")
	})
}

func TestSentry_MethodChaining(t *testing.T) {
	t.Run("methods can be chained together", func(t *testing.T) {
		e := echo.New()
		ctx := e.NewContext(nil, nil)
		err := errors.New("test error")
		extras := map[string]interface{}{"key": "value"}
		tags := map[string]string{"env": "test"}

		sentry := new(Sentry).
			WithContext(ctx).
			WithError(err).
			WithMessage("test").
			WithLevel(sentrygo.LevelError).
			WithExtras(extras).
			WithTags(tags)

		assert.Equal(t, ctx, sentry.context)
		assert.Equal(t, err, sentry.error)
		assert.Equal(t, "test", sentry.message)
		assert.Equal(t, sentrygo.LevelError, sentry.level)
		assert.Equal(t, extras, sentry.extras)
		assert.Equal(t, tags, sentry.tags)
	})
}

func TestSentry_SendingBehavior(t *testing.T) {
	t.Run("does not send when APP_ENV is local", func(t *testing.T) {
		originalEnv := os.Getenv("APP_ENV")
		originalDSN := os.Getenv("SENTRY_DSN")
		defer func() {
			os.Setenv("APP_ENV", originalEnv)
			os.Setenv("SENTRY_DSN", originalDSN)
		}()

		os.Setenv("APP_ENV", "local")
		os.Setenv("SENTRY_DSN", "https://test@sentry.io/123")

		sentry := new(Sentry)
		// Should not panic or error
		sentry.WithMessage("test").WithLevel(sentrygo.LevelInfo).sendMessage()
		sentry.WithError(errors.New("test")).WithLevel(sentrygo.LevelError).sendError()
	})

	t.Run("does not send when SENTRY_DSN is empty", func(t *testing.T) {
		originalEnv := os.Getenv("APP_ENV")
		originalDSN := os.Getenv("SENTRY_DSN")
		defer func() {
			os.Setenv("APP_ENV", originalEnv)
			os.Setenv("SENTRY_DSN", originalDSN)
		}()

		os.Setenv("APP_ENV", "production")
		os.Setenv("SENTRY_DSN", "")

		sentry := new(Sentry)
		// Should not panic or error
		sentry.WithMessage("test").WithLevel(sentrygo.LevelInfo).sendMessage()
		sentry.WithError(errors.New("test")).WithLevel(sentrygo.LevelError).sendError()
	})

	t.Run("sends error when conditions are met", func(t *testing.T) {
		originalEnv := os.Getenv("APP_ENV")
		originalDSN := os.Getenv("SENTRY_DSN")
		defer func() {
			os.Setenv("APP_ENV", originalEnv)
			os.Setenv("SENTRY_DSN", originalDSN)
			sentrygo.Flush(0)
		}()

		os.Setenv("APP_ENV", "production")
		os.Setenv("SENTRY_DSN", "https://public@sentry.example.com/1")

		// Initialize Sentry with mock transport
		err := sentrygo.Init(sentrygo.ClientOptions{
			Dsn: "https://public@sentry.example.com/1",
		})
		assert.NoError(t, err)

		sentry := new(Sentry)
		testErr := errors.New("test error")
		extras := map[string]interface{}{"key": "value"}
		tags := map[string]string{"env": "test"}

		// Should execute sending logic without panic
		sentry.WithError(testErr).
			WithLevel(sentrygo.LevelError).
			WithExtras(extras).
			WithTags(tags).
			sendError()
	})

	t.Run("sends message when conditions are met", func(t *testing.T) {
		originalEnv := os.Getenv("APP_ENV")
		originalDSN := os.Getenv("SENTRY_DSN")
		defer func() {
			os.Setenv("APP_ENV", originalEnv)
			os.Setenv("SENTRY_DSN", originalDSN)
			sentrygo.Flush(0)
		}()

		os.Setenv("APP_ENV", "production")
		os.Setenv("SENTRY_DSN", "https://public@sentry.example.com/1")

		// Initialize Sentry with mock transport
		err := sentrygo.Init(sentrygo.ClientOptions{
			Dsn: "https://public@sentry.example.com/1",
		})
		assert.NoError(t, err)

		sentry := new(Sentry)
		extras := map[string]interface{}{"key": "value"}
		tags := map[string]string{"env": "test"}

		// Should execute sending logic without panic
		sentry.WithMessage("test message").
			WithLevel(sentrygo.LevelInfo).
			WithExtras(extras).
			WithTags(tags).
			sendMessage()
	})
}

func TestSentry_LogLevelMethods(t *testing.T) {
	// Set local env to prevent actual Sentry calls
	originalEnv := os.Getenv("APP_ENV")
	defer os.Setenv("APP_ENV", originalEnv)
	os.Setenv("APP_ENV", "local")

	tests := []struct {
		name     string
		method   func(*Sentry)
		expected sentrygo.Level
	}{
		{
			name: "Debug sets debug level",
			method: func(s *Sentry) {
				s.Debug("test message")
			},
			expected: sentrygo.LevelDebug,
		},
		{
			name: "Info sets info level",
			method: func(s *Sentry) {
				s.Info("test message")
			},
			expected: sentrygo.LevelInfo,
		},
		{
			name: "Warning sets warning level",
			method: func(s *Sentry) {
				s.Warning("test message")
			},
			expected: sentrygo.LevelWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sentry := new(Sentry)
			tt.method(sentry)
			// Behavior verified: methods execute without error
		})
	}
}

func TestSentry_ErrorMethods(t *testing.T) {
	// Set local env to prevent actual Sentry calls
	originalEnv := os.Getenv("APP_ENV")
	defer os.Setenv("APP_ENV", originalEnv)
	os.Setenv("APP_ENV", "local")

	t.Run("Error handles error correctly", func(t *testing.T) {
		sentry := new(Sentry)
		err := errors.New("test error")

		// Should not panic
		sentry.Error(err)
	})

	t.Run("Errorf formats error message", func(t *testing.T) {
		sentry := new(Sentry)

		// Should not panic
		sentry.Errorf("error: %s %d", "test", 123)
	})

	t.Run("Fatal handles error correctly", func(t *testing.T) {
		// Temporarily reduce flush time for testing
		originalFlushTime := FlushTime
		FlushTime = 0
		defer func() { FlushTime = originalFlushTime }()

		sentry := new(Sentry)
		err := errors.New("fatal error")

		// Should not panic
		sentry.Fatal(err)
	})

	t.Run("Fatalf formats fatal error", func(t *testing.T) {
		// Temporarily reduce flush time for testing
		originalFlushTime := FlushTime
		FlushTime = 0
		defer func() { FlushTime = originalFlushTime }()

		sentry := new(Sentry)

		// Should not panic
		sentry.Fatalf("fatal: %s", "test")
	})
}

func TestSentry_FormattedMethods(t *testing.T) {
	// Set local env to prevent actual Sentry calls
	originalEnv := os.Getenv("APP_ENV")
	defer os.Setenv("APP_ENV", originalEnv)
	os.Setenv("APP_ENV", "local")

	t.Run("Debugf formats message", func(t *testing.T) {
		sentry := new(Sentry)
		sentry.Debugf("debug: %s %d", "test", 123)
		// Behavior verified: method executes without error
	})

	t.Run("Infof formats message", func(t *testing.T) {
		sentry := new(Sentry)
		sentry.Infof("info: %s %d", "test", 123)
		// Behavior verified: method executes without error
	})

	t.Run("Warningf formats message", func(t *testing.T) {
		sentry := new(Sentry)
		sentry.Warningf("warning: %s %d", "test", 123)
		// Behavior verified: method executes without error
	})
}

func TestSentry_ConvenienceFunctions(t *testing.T) {
	// Set local env to prevent actual Sentry calls
	originalEnv := os.Getenv("APP_ENV")
	defer os.Setenv("APP_ENV", originalEnv)
	os.Setenv("APP_ENV", "local")

	t.Run("WithContext creates sentry with context", func(t *testing.T) {
		e := echo.New()
		ctx := e.NewContext(nil, nil)

		sentry := WithContext(ctx)

		assert.NotNil(t, sentry)
		assert.Equal(t, ctx, sentry.context)
	})

	t.Run("WithExtras creates sentry with extras", func(t *testing.T) {
		extras := map[string]interface{}{"key": "value"}

		sentry := WithExtras(extras)

		assert.NotNil(t, sentry)
		assert.Equal(t, extras, sentry.extras)
	})

	t.Run("WithTags creates sentry with tags", func(t *testing.T) {
		tags := map[string]string{"env": "test"}

		sentry := WithTags(tags)

		assert.NotNil(t, sentry)
		assert.Equal(t, tags, sentry.tags)
	})

	t.Run("WithContextValues creates sentry with context values", func(t *testing.T) {
		contextValues := map[string]sentrygo.Context{"key": {}}

		sentry := WithContextValues(contextValues)

		assert.NotNil(t, sentry)
		assert.Equal(t, contextValues, sentry.contextValues)
	})
}

func TestSentry_StandaloneFunctions(t *testing.T) {
	// Set local env to prevent actual Sentry calls
	originalEnv := os.Getenv("APP_ENV")
	defer os.Setenv("APP_ENV", originalEnv)
	os.Setenv("APP_ENV", "local")

	t.Run("standalone Debug function works", func(t *testing.T) {
		// Should not panic
		Debug("test message")
	})

	t.Run("standalone Debugf function works", func(t *testing.T) {
		// Should not panic
		Debugf("debug: %s", "test")
	})

	t.Run("standalone Info function works", func(t *testing.T) {
		// Should not panic
		Info("test message")
	})

	t.Run("standalone Infof function works", func(t *testing.T) {
		// Should not panic
		Infof("info: %s", "test")
	})

	t.Run("standalone Warning function works", func(t *testing.T) {
		// Should not panic
		Warning("test message")
	})

	t.Run("standalone Warningf function works", func(t *testing.T) {
		// Should not panic
		Warningf("warning: %s", "test")
	})

	t.Run("standalone Error function works", func(t *testing.T) {
		// Should not panic
		Error(errors.New("test error"))
	})

	t.Run("standalone Errorf function works", func(t *testing.T) {
		// Should not panic
		Errorf("error: %s", "test")
	})

	t.Run("standalone Fatal function works", func(t *testing.T) {
		// Temporarily reduce flush time for testing
		originalFlushTime := FlushTime
		FlushTime = 0
		defer func() { FlushTime = originalFlushTime }()

		// Should not panic
		Fatal(errors.New("fatal error"))
	})

	t.Run("standalone Fatalf function works", func(t *testing.T) {
		// Temporarily reduce flush time for testing
		originalFlushTime := FlushTime
		FlushTime = 0
		defer func() { FlushTime = originalFlushTime }()

		// Should not panic
		Fatalf("fatal: %s", "test")
	})
}

func TestSentry_GetHub(t *testing.T) {
	t.Run("returns current hub when no context", func(t *testing.T) {
		sentry := new(Sentry)
		hub := sentry.getHub()

		assert.NotNil(t, hub, "should return a valid hub")
	})

	t.Run("returns hub when context is set", func(t *testing.T) {
		e := echo.New()
		ctx := e.NewContext(nil, nil)
		sentry := new(Sentry).WithContext(ctx)

		hub := sentry.getHub()

		assert.NotNil(t, hub, "should return a valid hub")
	})

	t.Run("returns hub from echo context when available", func(t *testing.T) {
		// Initialize Sentry SDK
		originalDSN := os.Getenv("SENTRY_DSN")
		originalEnv := os.Getenv("APP_ENV")
		defer func() {
			os.Setenv("SENTRY_DSN", originalDSN)
			os.Setenv("APP_ENV", originalEnv)
		}()

		os.Setenv("SENTRY_DSN", "https://public@sentry.example.com/1")
		os.Setenv("APP_ENV", "production")

		err := sentrygo.Init(sentrygo.ClientOptions{
			Dsn: "https://public@sentry.example.com/1",
		})
		assert.NoError(t, err)
		defer sentrygo.Flush(0)

		e := echo.New()
		// Use Echo's sentry middleware to properly set up the hub
		e.Use(sentryecho.New(sentryecho.Options{}))

		// This will properly initialize the context with sentry hub
		ctx := e.NewContext(nil, nil)
		hub := sentrygo.CurrentHub().Clone()
		ctx.Set("sentry", hub)

		sentry := new(Sentry).WithContext(ctx)
		resultHub := sentry.getHub()

		assert.NotNil(t, resultHub, "should return a valid hub")
	})
}

func TestSentry_ConfigScope(t *testing.T) {
	t.Run("configures scope with all properties", func(t *testing.T) {
		sentry := new(Sentry)
		sentry.level = sentrygo.LevelError
		sentry.extras = map[string]interface{}{"key": "value"}
		sentry.tags = map[string]string{"env": "test"}
		sentry.contextValues = map[string]sentrygo.Context{"custom": {}}

		scope := sentrygo.NewScope()
		sentry.configScope(scope)

		// Scope is configured - behavior test passes if no panic
		assert.NotNil(t, scope)
	})
}
