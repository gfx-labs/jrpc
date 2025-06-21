package jmux

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"gfx.cafe/open/jrpc/pkg/jsonrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock response writer for testing
type mockResponseWriter struct {
	result      any
	err         error
	extraFields jsonrpc.ExtraFields
}

func (m *mockResponseWriter) Send(result any, err error) error {
	m.result = result
	m.err = err
	return nil
}

func (m *mockResponseWriter) Notify(event string, data any) error {
	return nil
}

func (m *mockResponseWriter) ExtraFields() jsonrpc.ExtraFields {
	if m.extraFields == nil {
		m.extraFields = make(jsonrpc.ExtraFields)
	}
	return m.extraFields
}

func TestNewMux(t *testing.T) {
	mx := NewMux()
	assert.NotNil(t, mx)
	assert.NotNil(t, mx.tree)
	assert.NotNil(t, mx.pool)
	assert.Nil(t, mx.handler)
	assert.Empty(t, mx.middlewares)
}

func TestBasicRouting(t *testing.T) {
	tests := []struct {
		name           string
		routes         map[string]jsonrpc.HandlerFunc
		requestMethod  string
		expectedResult any
		expectedError  error
	}{
		{
			name: "simple route",
			routes: map[string]jsonrpc.HandlerFunc{
				"/test": func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
					w.Send("test response", nil)
				},
			},
			requestMethod:  "/test",
			expectedResult: "test response",
		},
		{
			name: "nested route",
			routes: map[string]jsonrpc.HandlerFunc{
				"/api/v1/users": func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
					w.Send("users list", nil)
				},
			},
			requestMethod:  "/api/v1/users",
			expectedResult: "users list",
		},
		{
			name: "route not found",
			routes: map[string]jsonrpc.HandlerFunc{
				"/exists": func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
					w.Send("exists", nil)
				},
			},
			requestMethod: "/notfound",
			expectedError: jsonrpc.NewMethodNotFoundError("/notfound"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mx := NewMux()
			
			// Register routes
			for pattern, handler := range tt.routes {
				mx.HandleFunc(pattern, handler)
			}

			// Create request and response writer
			w := &mockResponseWriter{}
			r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), tt.requestMethod, nil)

			// Serve the request
			mx.ServeRPC(w, r)

			// Check results
			if tt.expectedError != nil {
				assert.Error(t, w.err)
				assert.Equal(t, tt.expectedError.Error(), w.err.Error())
			} else {
				assert.NoError(t, w.err)
				assert.Equal(t, tt.expectedResult, w.result)
			}
		})
	}
}

func TestRouteParameters(t *testing.T) {
	tests := []struct {
		name           string
		pattern        string
		requestMethod  string
		expectedParams map[string]string
		shouldMatch    bool
	}{
		{
			name:          "single parameter",
			pattern:       "/users/{id}",
			requestMethod: "/users/123",
			expectedParams: map[string]string{
				"id": "123",
			},
			shouldMatch: true,
		},
		{
			name:          "multiple parameters",
			pattern:       "/users/{userId}/posts/{postId}",
			requestMethod: "/users/456/posts/789",
			expectedParams: map[string]string{
				"userId": "456",
				"postId": "789",
			},
			shouldMatch: true,
		},
		{
			name:          "regexp parameter",
			pattern:       "/users/{id:[0-9]+}",
			requestMethod: "/users/123",
			expectedParams: map[string]string{
				"id": "123",
			},
			shouldMatch: true,
		},
		{
			name:          "regexp parameter no match",
			pattern:       "/users/{id:[0-9]+}",
			requestMethod: "/users/abc",
			shouldMatch:   false,
		},
		{
			name:          "catch all parameter",
			pattern:       "/api/*",
			requestMethod: "/api/v1/users/123",
			expectedParams: map[string]string{
				"*": "v1/users/123",
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mx := NewMux()
			var capturedParams map[string]string

			mx.HandleFunc(tt.pattern, func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
				capturedParams = make(map[string]string)
				rctx := RouteContext(r.Context())
				if rctx != nil {
					for i, key := range rctx.MethodParams.Keys {
						if i < len(rctx.MethodParams.Values) {
							capturedParams[key] = rctx.MethodParams.Values[i]
						}
					}
				}
				w.Send("ok", nil)
			})

			w := &mockResponseWriter{}
			r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), tt.requestMethod, nil)

			mx.ServeRPC(w, r)

			if tt.shouldMatch {
				assert.NoError(t, w.err)
				assert.Equal(t, tt.expectedParams, capturedParams)
			} else {
				assert.Error(t, w.err)
			}
		})
	}
}

func TestMethodParam(t *testing.T) {
	mx := NewMux()
	
	var capturedId string
	mx.HandleFunc("/users/{id}", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		capturedId = MethodParam(r, "id")
		w.Send(fmt.Sprintf("user_%s", capturedId), nil)
	})

	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/users/42", nil)

	mx.ServeRPC(w, r)

	assert.NoError(t, w.err)
	assert.Equal(t, "42", capturedId)
	assert.Equal(t, "user_42", w.result)
}

func TestMiddleware(t *testing.T) {
	mx := NewMux()
	
	var middlewareOrder []string
	
	// Global middleware
	mx.Use(func(next jsonrpc.Handler) jsonrpc.Handler {
		return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			middlewareOrder = append(middlewareOrder, "global1")
			next.ServeRPC(w, r)
		})
	})
	
	mx.Use(func(next jsonrpc.Handler) jsonrpc.Handler {
		return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			middlewareOrder = append(middlewareOrder, "global2")
			next.ServeRPC(w, r)
		})
	})
	
	// Route with inline middleware
	mx.With(func(next jsonrpc.Handler) jsonrpc.Handler {
		return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			middlewareOrder = append(middlewareOrder, "inline")
			next.ServeRPC(w, r)
		})
	}).HandleFunc("/test", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		middlewareOrder = append(middlewareOrder, "handler")
		w.Send("ok", nil)
	})

	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/test", nil)

	mx.ServeRPC(w, r)

	assert.NoError(t, w.err)
	assert.Equal(t, []string{"global1", "global2", "inline", "handler"}, middlewareOrder)
}

func TestGroup(t *testing.T) {
	mx := NewMux()
	
	var middlewareOrder []string
	
	mx.Use(func(next jsonrpc.Handler) jsonrpc.Handler {
		return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			middlewareOrder = append(middlewareOrder, "root")
			next.ServeRPC(w, r)
		})
	})
	
	mx.Group(func(r Router) {
		r.Use(func(next jsonrpc.Handler) jsonrpc.Handler {
			return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
				middlewareOrder = append(middlewareOrder, "group")
				next.ServeRPC(w, r)
			})
		})
		
		r.HandleFunc("/grouped", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			middlewareOrder = append(middlewareOrder, "handler")
			w.Send("grouped", nil)
		})
	})

	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/grouped", nil)

	mx.ServeRPC(w, r)

	assert.NoError(t, w.err)
	assert.Equal(t, []string{"root", "group", "handler"}, middlewareOrder)
}

func TestMount(t *testing.T) {
	mx := NewMux()
	
	// Create a sub-router
	sub := NewRouter()
	sub.HandleFunc("/users", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send("users from sub", nil)
	})
	sub.HandleFunc("/users/{id}", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		id := MethodParam(r, "id")
		w.Send(fmt.Sprintf("user %s from sub", id), nil)
	})
	
	// Mount the sub-router
	mx.Mount("/api/v1", sub)
	
	tests := []struct {
		method         string
		expectedResult any
	}{
		{
			method:         "/api/v1/users",
			expectedResult: "users from sub",
		},
		{
			method:         "/api/v1/users/123",
			expectedResult: "user 123 from sub",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			w := &mockResponseWriter{}
			r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), tt.method, nil)
			
			mx.ServeRPC(w, r)
			
			assert.NoError(t, w.err)
			assert.Equal(t, tt.expectedResult, w.result)
		})
	}
}

func TestRoute(t *testing.T) {
	mx := NewMux()
	
	mx.Route("/api", func(r Router) {
		r.HandleFunc("/users", func(w jsonrpc.ResponseWriter, req *jsonrpc.Request) {
			w.Send("users", nil)
		})
		
		r.Route("/v1", func(r Router) {
			r.HandleFunc("/posts", func(w jsonrpc.ResponseWriter, req *jsonrpc.Request) {
				w.Send("v1 posts", nil)
			})
		})
	})
	
	tests := []struct {
		method         string
		expectedResult any
		shouldError    bool
	}{
		{
			method:         "/api/users",
			expectedResult: "users",
		},
		{
			method:         "/api/v1/posts",
			expectedResult: "v1 posts",
		},
		{
			method:      "/api/v2/posts",
			shouldError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			w := &mockResponseWriter{}
			r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), tt.method, nil)
			
			mx.ServeRPC(w, r)
			
			if tt.shouldError {
				assert.Error(t, w.err)
			} else {
				assert.NoError(t, w.err)
				assert.Equal(t, tt.expectedResult, w.result)
			}
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	mx := NewMux()
	
	// Set custom not found handler
	mx.NotFound(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send(nil, errors.New("custom not found"))
	})
	
	mx.HandleFunc("/exists", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send("exists", nil)
	})
	
	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/doesnotexist", nil)
	
	mx.ServeRPC(w, r)
	
	assert.Error(t, w.err)
	assert.Equal(t, "custom not found", w.err.Error())
}

func TestMatch(t *testing.T) {
	mx := NewMux()
	
	mx.HandleFunc("/users", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {})
	mx.HandleFunc("/users/{id}", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {})
	mx.HandleFunc("/posts/*", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {})
	
	tests := []struct {
		path        string
		shouldMatch bool
	}{
		{"/users", true},
		{"/users/123", true},
		{"/posts/anything/goes/here", true},
		{"/comments", false},
		{"/user", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rctx := NewRouteContext()
			matched := mx.Match(rctx, tt.path)
			assert.Equal(t, tt.shouldMatch, matched)
		})
	}
}

func TestRoutes(t *testing.T) {
	mx := NewMux()
	
	mx.HandleFunc("/users", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {})
	mx.HandleFunc("/users/{id}", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {})
	mx.HandleFunc("/posts", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {})
	
	routes := mx.Routes()
	assert.Len(t, routes, 3)
	
	// Check that all patterns are present
	patterns := make(map[string]bool)
	for _, route := range routes {
		patterns[route.Pattern] = true
	}
	
	assert.True(t, patterns["/users"])
	assert.True(t, patterns["/users/{id}"])
	assert.True(t, patterns["/posts"])
}

// Test panic scenarios
func TestPanics(t *testing.T) {
	t.Run("panic on middleware after routes", func(t *testing.T) {
		mx := NewMux()
		mx.HandleFunc("/test", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {})
		
		assert.Panics(t, func() {
			mx.Use(func(next jsonrpc.Handler) jsonrpc.Handler {
				return next
			})
		})
	})
	
	t.Run("panic on nil route function", func(t *testing.T) {
		mx := NewMux()
		assert.Panics(t, func() {
			mx.Route("/test", nil)
		})
	})
	
	t.Run("panic on nil mount handler", func(t *testing.T) {
		mx := NewMux()
		assert.Panics(t, func() {
			mx.Mount("/test", nil)
		})
	})
	
	t.Run("panic on duplicate mount pattern", func(t *testing.T) {
		mx := NewMux()
		sub1 := NewRouter()
		sub2 := NewRouter()
		
		mx.Mount("/api", sub1)
		assert.Panics(t, func() {
			mx.Mount("/api", sub2)
		})
	})
}

// TestService for struct registration tests
type TestService struct{}

func (s *TestService) Add(a, b int) int {
	return a + b
}

func (s *TestService) Subtract(a, b int) int {
	return a - b
}

// Test struct registration
func TestRegisterStruct(t *testing.T) {
	
	mx := NewMux()
	service := &TestService{}
	
	err := mx.RegisterStruct("math", service)
	require.NoError(t, err)
	
	// Test that methods are accessible (formatName converts first char to lowercase)
	tests := []struct {
		method string
		exists bool
	}{
		{"/math/add", true},
		{"/math/subtract", true},
		{"/math/multiply", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			rctx := NewRouteContext()
			matched := mx.Match(rctx, tt.method)
			assert.Equal(t, tt.exists, matched)
		})
	}
}

// Test function registration
func TestRegisterFunc(t *testing.T) {
	mx := NewMux()
	
	addFunc := func(a, b int) int {
		return a + b
	}
	
	err := mx.RegisterFunc("add", addFunc)
	require.NoError(t, err)
	
	// Check that the function is registered
	rctx := NewRouteContext()
	matched := mx.Match(rctx, "/add")
	assert.True(t, matched)
}

// Test context propagation
func TestContextPropagation(t *testing.T) {
	mx := NewMux()
	
	type ctxKey string
	testKey := ctxKey("test")
	testValue := "test-value"
	
	mx.HandleFunc("/test", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		val := r.Context().Value(testKey)
		assert.Equal(t, testValue, val)
		w.Send("ok", nil)
	})
	
	ctx := context.WithValue(context.Background(), testKey, testValue)
	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(ctx, jsonrpc.NewStringIDPtr("1"), "/test", nil)
	
	mx.ServeRPC(w, r)
	
	assert.NoError(t, w.err)
}

// Test empty pattern handling
func TestEmptyPattern(t *testing.T) {
	mx := NewMux()
	
	assert.Panics(t, func() {
		mx.HandleFunc("", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {})
	})
}

// Test pattern normalization
func TestPatternNormalization(t *testing.T) {
	mx := NewMux()
	
	mx.HandleFunc("users", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send("normalized", nil)
	})
	
	// Should work with leading slash
	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/users", nil)
	mx.ServeRPC(w, r)
	
	assert.NoError(t, w.err)
	assert.Equal(t, "normalized", w.result)
}

// Test middleware on sub-routers
func TestSubRouterMiddleware(t *testing.T) {
	mx := NewMux()
	var order []string
	
	mx.Use(func(next jsonrpc.Handler) jsonrpc.Handler {
		return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			order = append(order, "root")
			next.ServeRPC(w, r)
		})
	})
	
	sub := NewRouter()
	sub.Use(func(next jsonrpc.Handler) jsonrpc.Handler {
		return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
			order = append(order, "sub")
			next.ServeRPC(w, r)
		})
	})
	
	sub.HandleFunc("/test", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		order = append(order, "handler")
		w.Send("ok", nil)
	})
	
	mx.Mount("/sub", sub)
	
	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/sub/test", nil)
	mx.ServeRPC(w, r)
	
	assert.NoError(t, w.err)
	assert.Equal(t, []string{"root", "sub", "handler"}, order)
}

// Test RouteContext helper functions
func TestRouteContextHelpers(t *testing.T) {
	mx := NewMux()
	
	mx.HandleFunc("/users/{id}/posts/{postId}", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		// Test MethodParam function
		userId := MethodParam(r, "id")
		postId := MethodParam(r, "postId")
		
		// Test MethodParamFromCtx function
		userIdFromCtx := MethodParamFromCtx(r.Context(), "id")
		postIdFromCtx := MethodParamFromCtx(r.Context(), "postId")
		
		assert.Equal(t, userId, userIdFromCtx)
		assert.Equal(t, postId, postIdFromCtx)
		
		// Test RoutePattern
		rctx := RouteContext(r.Context())
		pattern := rctx.RoutePattern()
		assert.NotEmpty(t, pattern)
		
		w.Send(fmt.Sprintf("user=%s,post=%s", userId, postId), nil)
	})
	
	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/users/123/posts/456", nil)
	mx.ServeRPC(w, r)
	
	assert.NoError(t, w.err)
	assert.Equal(t, "user=123,post=456", w.result)
}

// Test nested groups
func TestNestedGroups(t *testing.T) {
	mx := NewMux()
	var middlewares []string
	
	mx.Group(func(r Router) {
		r.Use(func(next jsonrpc.Handler) jsonrpc.Handler {
			return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, req *jsonrpc.Request) {
				middlewares = append(middlewares, "group1")
				next.ServeRPC(w, req)
			})
		})
		
		r.Group(func(r Router) {
			r.Use(func(next jsonrpc.Handler) jsonrpc.Handler {
				return jsonrpc.HandlerFunc(func(w jsonrpc.ResponseWriter, req *jsonrpc.Request) {
					middlewares = append(middlewares, "group2")
					next.ServeRPC(w, req)
				})
			})
			
			r.HandleFunc("/nested", func(w jsonrpc.ResponseWriter, req *jsonrpc.Request) {
				middlewares = append(middlewares, "handler")
				w.Send("nested", nil)
			})
		})
	})
	
	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/nested", nil)
	mx.ServeRPC(w, r)
	
	assert.NoError(t, w.err)
	assert.Equal(t, []string{"group1", "group2", "handler"}, middlewares)
}

// Test method not allowed handler
func TestMethodNotAllowed(t *testing.T) {
	mx := NewMux()
	
	// Set custom method not allowed handler
	var customHandlerCalled bool
	mx.MethodNotAllowed(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		customHandlerCalled = true
		w.Send(nil, errors.New("custom method not allowed"))
	})
	
	// In jmux, methodNotAllowed is triggered when a route pattern matches but has no handler
	// This is a bit tricky to trigger directly, but we can test the handler is set correctly
	
	// First, let's verify the custom handler is set
	assert.NotNil(t, mx.MethodNotAllowedHandler())
	
	// Test default method not allowed handler
	mx2 := NewMux()
	defaultHandler := mx2.MethodNotAllowedHandler()
	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/test", nil)
	defaultHandler.ServeRPC(w, r)
	assert.Error(t, w.err)
	assert.Equal(t, "forbidden", w.err.Error())
	
	// Test custom method not allowed handler
	customHandler := mx.MethodNotAllowedHandler()
	w2 := &mockResponseWriter{}
	customHandler.ServeRPC(w2, r)
	assert.Error(t, w2.err)
	assert.Equal(t, "custom method not allowed", w2.err.Error())
	assert.True(t, customHandlerCalled)
}

// Test fetching non-existent route parameters
func TestNonExistentRouteParams(t *testing.T) {
	mx := NewMux()
	
	mx.HandleFunc("/users/{id}", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		// Try to fetch existing param
		id := MethodParam(r, "id")
		assert.Equal(t, "123", id)
		
		// Try to fetch non-existent params
		nonExistent := MethodParam(r, "nonexistent")
		assert.Equal(t, "", nonExistent)
		
		// Also test with context
		nonExistentFromCtx := MethodParamFromCtx(r.Context(), "doesnotexist")
		assert.Equal(t, "", nonExistentFromCtx)
		
		w.Send("ok", nil)
	})
	
	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/users/123", nil)
	mx.ServeRPC(w, r)
	
	assert.NoError(t, w.err)
}

// Test RouteContext when it doesn't exist
func TestRouteContextNotExists(t *testing.T) {
	// Create a request without going through the mux
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/test", nil)
	
	// These should return empty values when no route context exists
	param := MethodParam(r, "any")
	assert.Equal(t, "", param)
	
	paramFromCtx := MethodParamFromCtx(r.Context(), "any")
	assert.Equal(t, "", paramFromCtx)
	
	rctx := RouteContext(r.Context())
	assert.Nil(t, rctx)
}

// Test multiple parameters with same name (last one wins)
func TestDuplicateParamNames(t *testing.T) {
	mx := NewMux()
	
	// Mount a sub-router that might have conflicting param names
	sub := NewRouter()
	sub.HandleFunc("/{id}/details", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		// The last 'id' in the path should win
		id := MethodParam(r, "id")
		w.Send(fmt.Sprintf("id=%s", id), nil)
	})
	
	mx.Mount("/users/{id}", sub)
	
	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/users/user123/456/details", nil)
	mx.ServeRPC(w, r)
	
	assert.NoError(t, w.err)
	// Should get the last id value (456)
	assert.Equal(t, "id=456", w.result)
}

// Test method not allowed with sub-routers
func TestMethodNotAllowedWithSubRouter(t *testing.T) {
	mx := NewMux()
	
	mx.MethodNotAllowed(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send(nil, errors.New("main method not allowed"))
	})
	
	sub := NewRouter()
	// Sub-router should inherit the parent's method not allowed handler
	mx.Mount("/api", sub)
	
	// Verify sub inherits the handler
	assert.NotNil(t, sub.MethodNotAllowedHandler())
	
	// Create a new sub-router to test setting method not allowed after mounting
	sub2 := NewRouter()
	sub2.MethodNotAllowed(func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send(nil, errors.New("sub method not allowed"))
	})
	mx.Mount("/api2", sub2)
	
	// Test that handlers are properly set
	assert.NotNil(t, mx.MethodNotAllowedHandler())
	assert.NotNil(t, sub2.MethodNotAllowedHandler())
	
	// Verify sub2 has its own handler
	w := &mockResponseWriter{}
	r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), "/test", nil)
	sub2.MethodNotAllowedHandler().ServeRPC(w, r)
	assert.Error(t, w.err)
	assert.Equal(t, "sub method not allowed", w.err.Error())
}

// Test edge cases for route parameters
func TestRouteParamEdgeCases(t *testing.T) {
	mx := NewMux()
	
	// Test empty parameter value
	mx.HandleFunc("/test/{param}", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		param := MethodParam(r, "param")
		w.Send(fmt.Sprintf("param='%s'", param), nil)
	})
	
	// Test with empty segment (consecutive slashes)
	mx.HandleFunc("/double//slash", func(w jsonrpc.ResponseWriter, r *jsonrpc.Request) {
		w.Send("double slash", nil)
	})
	
	tests := []struct {
		name           string
		path           string
		expectedResult any
		shouldError    bool
	}{
		{
			name:           "normal param",
			path:           "/test/value",
			expectedResult: "param='value'",
		},
		{
			name:           "param with special chars",
			path:           "/test/hello%20world",
			expectedResult: "param='hello%20world'",
		},
		{
			name:           "param with slashes encoded",
			path:           "/test/a%2Fb%2Fc",
			expectedResult: "param='a%2Fb%2Fc'",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &mockResponseWriter{}
			r := jsonrpc.NewRawRequest(context.Background(), jsonrpc.NewStringIDPtr("1"), tt.path, nil)
			mx.ServeRPC(w, r)
			
			if tt.shouldError {
				assert.Error(t, w.err)
			} else {
				assert.NoError(t, w.err)
				assert.Equal(t, tt.expectedResult, w.result)
			}
		})
	}
}