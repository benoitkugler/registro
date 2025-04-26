package services

const EndpointServices = "services"

type Service uint8

const (
	_ Service = iota
	TransfertFicheSanitaire
)
