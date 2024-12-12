#Events

## Incoming

RouteCreated

- id
- distance
- directions
- - lat
- - lng

##Proccess and return

FreightCalculated

- route_id
- amount

---

Delivery Started

- route_id

### effect

DriverMoved

- route_id
- lat
- lng
