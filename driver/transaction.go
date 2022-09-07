package driver

type biTableTransaction struct {
	conn *Conn
}

func (b biTableTransaction) Commit() error {
	return nil
}

func (b biTableTransaction) Rollback() error {
	return nil
}
