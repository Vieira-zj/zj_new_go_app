package demos

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	newerrors "github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func TestErrorsIs(t *testing.T) {
	err1 := fmt.Errorf("custom error")
	err2 := fmt.Errorf("custom error")
	t.Log("is err string equal:", err1.Error() == err2.Error())
	t.Log("is err equal:", errors.Is(err2, err1))

	err2 = fmt.Errorf("wrapped: %w", err1)
	t.Log("is wrapped err equal:", errors.Is(err2, err1)) // wrapped err as 1st arg

	valueOf := reflect.ValueOf(err1)
	if valueOf.Kind() == reflect.Ptr {
		t.Log("err is a pointer")
		valueOf = valueOf.Elem()
	}
	t.Log("type:", valueOf.Kind(), valueOf.Type().Name())
}

func TestWrappedError(t *testing.T) {
	rawErr := fmt.Errorf("raw error")
	wErr := fmt.Errorf("wrapped error => %w", rawErr)
	t.Log("error:", wErr)
	t.Log("is raw error:", errors.Is(wErr, rawErr))

	newErr := fmt.Errorf("new wrapped error => %w", wErr)
	t.Log("error:", newErr)
	t.Log("is raw error:", errors.Is(newErr, rawErr))

	err := errors.Unwrap(newErr)
	t.Log("unwrap error:", err)
	err = errors.Unwrap(err)
	t.Log("unwrap error:", err)
	t.Log("is raw error:", errors.Is(err, rawErr))

	newErr = fmt.Errorf("new error with msg => %v", wErr)
	t.Log("error:", newErr)
	t.Log("is raw error:", errors.Is(newErr, rawErr)) // false
}

func TestVerifyErrorType(t *testing.T) {
	err := &os.PathError{
		Path: "/tmp/test",
		Op:   "access",
		Err:  errors.New("path not exist"),
	}
	wErr := fmt.Errorf("wrapped error: %w", err)
	t.Log("error:", wErr)

	if _, ok := wErr.(*os.PathError); ok {
		t.Log("wrapped error is os.PathError")
	}

	var p *os.PathError
	if errors.As(wErr, &p) {
		t.Log("wrapped error as os.PathError")
	}
}

type CustomError struct {
	content string
}

func (e *CustomError) Error() string {
	return e.content
}

func TestCustomError(t *testing.T) {
	err := &CustomError{
		content: "system exception",
	}
	wErr := fmt.Errorf("wrapped error: %w", err)
	t.Log("error:", wErr)

	var tErr *CustomError
	t.Log("error as:", errors.As(wErr, &tErr))
}

func TestWrappedErrorWithMsg(t *testing.T) {
	rootErr := fmt.Errorf("this is a test root error")
	wrappedErr := newerrors.Wrap(rootErr, "wrapped error")
	newErr := newerrors.WithMessage(wrappedErr, "add message 1")
	newErr = newerrors.WithMessage(newErr, "add message 2")
	fmt.Println(newErr.Error())
	fmt.Println("errors.Is(newErr, rootErr):", errors.Is(newErr, rootErr))

	err := newerrors.Unwrap(newErr)
	fmt.Println("unwrapped:", err.Error())

	err = newerrors.Cause(newErr)
	t.Log("errors.Is(err, rootErr):", errors.Is(err, rootErr))
	fmt.Println("cause:", err.Error())
}

// use ErrGroup instead of WaitGroup for async func which returns error.

func TestErrGroup(t *testing.T) {
	g := new(errgroup.Group)
	urls := []string{
		"http://www.notexisturl.com/",
		"http://www.golang.org/",
		"http://www.google.com/",
	}
	for _, url := range urls {
		url := url
		g.Go(func() error {
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
				fmt.Println("pass for: " + url)
			}
			return err
		})
	}

	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
	t.Log("successfully fetched all URLs.")
}

// ErrGroup with context

type Result string
type Search func(ctx context.Context, query string) (Result, error)

func fakeSearch(kind string) Search {
	return func(_ context.Context, query string) (Result, error) {
		if kind == "video" {
			return "", fmt.Errorf("not support " + kind)
		}
		return Result(fmt.Sprintf("%s result for %q", kind, query)), nil
	}
}

var (
	Web   = fakeSearch("web")
	Image = fakeSearch("image")
	Video = fakeSearch("video")
)

func TestErrGroupWithContext(t *testing.T) {
	Google := func(ctx context.Context, query string) ([]Result, error) {
		g, ctx := errgroup.WithContext(ctx)

		searches := []Search{Web, Image, Video}
		results := make([]Result, len(searches))
		for i, search := range searches {
			i, search := i, search
			g.Go(func() error {
				result, err := search(ctx, query)
				if err == nil {
					results[i] = result
				}
				return err
			})
		}

		if err := g.Wait(); err != nil {
			return nil, err
		}
		return results, nil
	}

	results, err := Google(context.Background(), "golang")
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		t.Log(result)
	}
}

// pipeline by ErrGroup

type result struct {
	path string
	sum  [md5.Size]byte
}

func MD5All(ctx context.Context, root string) (map[string][md5.Size]byte, error) {
	g, ctx := errgroup.WithContext(ctx)
	paths := make(chan string)

	g.Go(func() error {
		defer close(paths)
		return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			select {
			case paths <- path:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	})

	c := make(chan result)
	const numDigesters = 5
	for i := 0; i < numDigesters; i++ {
		g.Go(func() error {
			for path := range paths {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				select {
				case c <- result{path, md5.Sum(data)}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	go func() {
		g.Wait()
		fmt.Println("close chan")
		close(c)
	}()

	m := make(map[string][md5.Size]byte)
	for r := range c {
		m[r.path] = r.sum
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return m, nil
}

func TestPipelineByErrGroup(t *testing.T) {
	path := filepath.Join(os.Getenv("HOME"), "Downloads/tmps")
	m, err := MD5All(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}

	for path, sum := range m {
		fmt.Printf("%s:\t%x\n", path, sum)
	}
	fmt.Printf("total: %d\n", len(m))
}
