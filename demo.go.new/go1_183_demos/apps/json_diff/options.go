package jsondiff

type options struct {
	ignores     map[string]struct{}
	sliceOrders map[string]struct{}
}

type optsFunc func(*options)

func WithIgnores(paths []string) optsFunc {
	return func(opts *options) {
		for _, path := range paths {
			opts.ignores[path] = struct{}{}
		}
	}
}

func WithSliceOrders(paths []string) optsFunc {
	return func(opts *options) {
		for _, path := range paths {
			opts.sliceOrders[path] = struct{}{}
		}
	}
}
