package main

import (
	"fmt"
)

type Person struct {
	Name string
	Age  int
}

func (p Person) PrintName() {
	fmt.Println("Name:", p.Name)
}

func (p Person) SetAge2(age int) {
	p.Age = age
}
func (p *Person) SetAge(age int) {
	p.Age = age
}

type Singer struct {
	Person // extends Person by embedding it
	works  []string
}

type My struct {
	num int
}

func (self My) AddOne() {
	self.num++
}
func (self *My) AddTwo() {
	self.num += 2
}
func Test() {
	my1 := My{1} //值接收者
	my1.AddOne()
	// 1 不改变num的值
	fmt.Println(my1.num)

	my2 := &My{1} // 指针接收者
	my2.AddTwo()
	// 3改变num的值
	fmt.Println(my2.num)
}

func main() {
	//Test()
	gaga := Singer{Person: Person{"Gaga", 30}}
	gaga.PrintName() // Name: Gaga
	gaga.Name = "Lady Gaga"
	gaga.SetAge2(1)
	fmt.Println(gaga.Age)
	//gaga.SetAge(31)
	//fmt.Println(&gaga)
	(&gaga).SetAge2(31)
	(&gaga).PrintName() // Name: Lady Gaga
	fmt.Println(gaga.Person.Age)
	fmt.Println(gaga.Age) // 31

	//person := Person{"1", 31}
	//person.PrintName()
	//person.SetAge2(1)
	//fmt.Println(person.Age)

	//var gaga = Singer{}
	//gaga.PrintName()
	//gaga.SetAge(1)
	//fmt.Println(gaga)

	//t := reflect.TypeOf(Singer{}) // the Singer type
	//fmt.Println(t, "has", t.NumField(), "fields:")
	//for i := 0; i < t.NumField(); i++ {
	//	fmt.Print(" field#", i, ": ", t.Field(i).Name, "\n")
	//}
	//fmt.Println(t, "has", t.NumMethod(), "methods:")
	//for i := 0; i < t.NumMethod(); i++ {
	//	fmt.Print(" method#", i, ": ", t.Method(i).Name, "\n")
	//}
	//
	//pt := reflect.TypeOf(&Singer{}) // the *Singer type
	//fmt.Println(pt, "has", pt.NumMethod(), "methods:")
	//for i := 0; i < pt.NumMethod(); i++ {
	//	fmt.Print(" method#", i, ": ", pt.Method(i).Name, "\n")
	//}
}
