package nower

import "time"

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=nower Nower
type Nower interface {
	Now() time.Time
}
