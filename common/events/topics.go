package events

type Topic string
type RoutingKey string

const (
	Product   Topic = "product"
	Inventory Topic = "inventory"
)

// routing key format <topic>.<action>
const (
	ProductCreated RoutingKey = "product.created"
)
