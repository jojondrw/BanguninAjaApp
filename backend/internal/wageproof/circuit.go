package wageproof

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

const (
	MaxWorkers   = 16
	treeDepth    = 4
	workDaysBase = 25
	maxDays      = 31
)

type WageCircuit struct {
	Root        frontend.Variable `gnark:",public"`
	MinimumWage frontend.Variable `gnark:",public"`
	Count       frontend.Variable `gnark:",public"`

	Worker [MaxWorkers]frontend.Variable
	Pay    [MaxWorkers]frontend.Variable
	Days   [MaxWorkers]frontend.Variable
	Salt   [MaxWorkers]frontend.Variable
}

func (c *WageCircuit) Define(api frontend.API) error {
	api.AssertIsLessOrEqual(c.Count, MaxWorkers)
	level := make([]frontend.Variable, MaxWorkers)
	for index := 0; index < MaxWorkers; index++ {
		active := api.IsZero(api.Add(api.Cmp(index, c.Count), 1))

		hasher, err := mimc.NewMiMC(api)
		if err != nil {
			return err
		}
		hasher.Write(c.Worker[index], c.Pay[index], c.Days[index], c.Salt[index])
		level[index] = api.Select(active, hasher.Sum(), 0)

		days := api.Select(active, c.Days[index], 1)
		api.AssertIsLessOrEqual(1, days)
		api.AssertIsLessOrEqual(days, maxDays)
		required := api.Select(active, api.Mul(c.MinimumWage, c.Days[index]), 0)
		earned := api.Select(active, api.Mul(c.Pay[index], workDaysBase), 0)
		api.AssertIsLessOrEqual(required, earned)
	}
	for depth := 0; depth < treeDepth; depth++ {
		next := make([]frontend.Variable, len(level)/2)
		for index := range next {
			hasher, err := mimc.NewMiMC(api)
			if err != nil {
				return err
			}
			hasher.Write(level[2*index], level[2*index+1])
			next[index] = hasher.Sum()
		}
		level = next
	}
	api.AssertIsEqual(level[0], c.Root)
	return nil
}
