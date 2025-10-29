package services

//go:generate ../../../../../go/src/github.com/benoitkugler/gomacro/cmd/gomacro services.go typescript/types:../../../registro-web/src/clients/services/logic/types.ts

const EndpointServices = "services"

type Service uint8

const (
	_ Service = iota
	TransfertFicheSanitaire
)
