package gore

import (
	"fmt"
)

func ExampleCommand() {
	conn, err := Dial("localhost:6379", 0)
	if err != nil {
		return
	}
        defer conn.Close()
	// Set key - value
	rep, _ := NewCommand("SET", "kirisame", "marisa").Run(conn)
	fmt.Println(rep.IsOk())

	//Get key - value
	rep, _ = NewCommand("GET", "kirisame").Run(conn)
	s, _ := rep.String()
	fmt.Println(s)
	//Output: true
	// marisa
}

func ExampleReply() {
	conn, err := Dial("localhost:6379", 0)
        if err != nil {
		return
	}
        defer conn.Close()

	// Set integer value
	rep, _ := NewCommand("SET", "int", 123456789).Run(conn)
	fmt.Println(rep.IsOk())

	// Get integer value
	rep, _ = NewCommand("GET", "int").Run(conn)
	x, _ := rep.Int()
	fmt.Println(x)

	//Output: true
	// 123456789
}

func ExampleReply_convert() {
	conn, err := Dial("localhost:6379", 0)
	if err != nil {
		return
	}
        defer conn.Close()

	rep, _ := NewCommand("ZRANGE", "test", 0, -1).Run(conn)
	s := []string{}
	_ = rep.Slice(&s)
	fmt.Println(s)
}

func ExampleReply_pair() {
	conn, err := Dial("localhost:6379", 0)
        if err != nil {
                return
        }
        defer conn.Close()

        rep, _ := NewCommand("HGETALL", "test").Run(conn)
        s := []*Pair{}
        _ = rep.Slice(&s)
	fmt.Println(s)
}