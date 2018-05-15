package main

import "gopkg.in/mgo.v2"
import "gopkg.in/mgo.v2/bson"
import "log"
import "fmt"

type Person struct {
	Name  string
	Phone string
}

// Insert Person entry to the Mongo Collection
func Insert(c *mgo.Collection, p *Person) {
	err := c.Insert(p)
	if err != nil {
		log.Fatal(err)
	}
}

// SelectOneByName selects one Person by his/her name from given mgo Collection
func SelectOneByName(c *mgo.Collection, name string) Person {
	result := Person{}
	err := c.Find(bson.M{"name": name}).One(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func main() {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	c := session.DB("test").C("users")

	Insert(c, &Person{"Alice", "555"})

	result := SelectOneByName(c, "Alice")
	fmt.Println("Phone: ", result.Phone)
}
