package models

type DeliveryChannel string // @name DeliveryChannel

func (o DeliveryChannel) String() string { return string(o) }

func (o DeliveryChannel) Valid() bool {
	_, ok := validDeliveryChannels[o]
	return ok
}

const (
	EmailDeliveryChannel    DeliveryChannel = "email"
	TelegramDeliveryChannel DeliveryChannel = "telegram"
)

var validDeliveryChannels = map[DeliveryChannel]struct{}{
	EmailDeliveryChannel:    {},
	TelegramDeliveryChannel: {},
}
