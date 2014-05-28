package gore

// Transaction implements MULTI/EXEC/WATCH protocol of redis
type Transaction struct {
	watchedKeys [][]byte
	commands    []*Command
}
