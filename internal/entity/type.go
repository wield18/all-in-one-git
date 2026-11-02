package entity

type CanResultMany interface {
	ToMapForTransmitMany() map[string]interface{}
}
