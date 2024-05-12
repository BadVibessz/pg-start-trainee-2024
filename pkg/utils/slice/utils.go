package slice

func Map[T any, V any](s []T, f func(T) V) []V {
	res := make([]V, len(s))

	for i, e := range s {
		res[i] = f(e)
	}

	return res
}
