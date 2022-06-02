package operator

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type Key interface{}

type entry struct {
	key             Key
	value           interface{}
	accessFrequency int
}

type CacheOperator interface {
	Add(key Key, value interface{}, len int)
	Get(key Key) (value interface{}, ok bool)
	Remove(key Key)
	RemoveBasedOnPolicy()
	Len() int
	Clear()
	ContainsKey(key Key) bool
	NextVictim() (key string)
}
