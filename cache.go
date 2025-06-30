package main

type IStore[KT any, VT any] interface {
	Get(key KT) (VT, error)
	Set(key KT, value VT) error
}
