package zk

import (
	"expvar"
	"runtime"
)

func currentGoVersion() interface{} {
	return runtime.Version()
}

func getNumCPUs() interface{} {
	return runtime.NumCPU()
}

func getGoOS() interface{} {
	return runtime.GOOS
}

func getNumGoroutins() interface{} {
	return runtime.NumGoroutine()
}

func getNumCgoCall() interface{} {
	return runtime.NumCgoCall()
}

func getNumConnections() interface{} {
	return activeSessions.Count()
}

func init() {
	// http.HandleFunc("/debug/vars", GetMetric)
	expvar.Publish("version", expvar.Func(currentGoVersion))
	expvar.Publish("cpus", expvar.Func(getNumCPUs))
	expvar.Publish("os", expvar.Func(getGoOS))
	expvar.Publish("cgo", expvar.Func(getNumCgoCall))
	expvar.Publish("goroutine", expvar.Func(getNumGoroutins))
	expvar.Publish("connections", expvar.Func(getNumConnections))
}
