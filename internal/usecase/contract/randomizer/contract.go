package randomizer

//go:generate mockgen -source=contract.go -destination=mocks/contract_mock.go -package=randomizer Randomizer
type Randomizer interface {
	Shuffle(n int, swap func(i, j int))
}
