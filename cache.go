package main

type ArrayCache struct {
	Elements     []*string
	CurrentIndex int
	MaxSize      int
}

func InitArrayCache(maxSize int) *ArrayCache {
	return &ArrayCache{
		Elements:     make([]*string, maxSize),
		CurrentIndex: 0,
		MaxSize:      maxSize,
	}
}

func (c *ArrayCache) AddElement(element *string) {
	c.Elements[c.CurrentIndex] = element
	c.CurrentIndex = (c.CurrentIndex + 1) % c.MaxSize
}
