package utils

func ClearMap[M ~map[K]V, K comparable, V any](m M) {
    for k := range m {
        delete(m, k)
    }
}