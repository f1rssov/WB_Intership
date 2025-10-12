package main

import "fmt"

type Human struct {
	Body string
	Old  int
}

func (h *Human) ShowBody() {
	fmt.Println(h.Body)
}

func (h *Human) ShowOld() {
	fmt.Println(h.Old)
}

type Action struct {
	Human
	num int
}

type ExampleInterface interface {
	ShowBody()
	ShowOld()
}

func main() {
	example := &Action{
		Human: Human{
			Body: "Hello World!",
			Old:  20,
		},
		num: 10,
	}
	example.ShowBody()
	example.ShowOld()
	example.Human.ShowBody()
	example.Human.ShowOld()

	var inter ExampleInterface
	inter = example
	inter.ShowBody()
	inter.ShowOld()

	fmt.Println(example.Body, example.Old)
}
