package mock

//go:generate go tool mockgen -destination=repositories.go -package=mock fluffy-giggle/core/internal/domain/port RunRepository,StepRepository,BenchmarkRepository,CaseCatalog
//go:generate go tool mockgen -destination=pipeline.go -package=mock fluffy-giggle/core/internal/domain/port Model,Detector,RunQueue,EventBus
