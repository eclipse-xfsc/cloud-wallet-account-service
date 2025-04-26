package common

type ProtocolType string

const BasePath = "/api/accounts"

const (
	Nats ProtocolType = "nats"
)

type DatabaseType string

const (
	Postgres DatabaseType = "postgres"
)

type DataKey string

const (
	UserKey DataKey = "user"
	TTLKey  DataKey = "ttl"
)

const ModeDownload = "download"
const ModeUpload = "upload"

const (
	Consent             RecordEventType = "consent"
	Pairing             RecordEventType = "pairing"
	Issued              RecordEventType = "issued"
	Presented           RecordEventType = "presented"
	Revoked             RecordEventType = "revoked"
	PresentationRequest RecordEventType = "presentationRequest"
	DeviceConnection    RecordEventType = "device.connection"
)

const EventTypeOfferingAcceptance = "retrieval.offering.acceptance"

func RecordEventTypes() []RecordEventType {
	return []RecordEventType{Consent, Pairing, Issued, Presented, Revoked, PresentationRequest, DeviceConnection}
}
