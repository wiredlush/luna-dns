//go:build web

package web

func init() {
	StartFunc = startFromFlags
}
