package types 

type WSMessage struct{
	Type string 
	Data []byte
}
type Login struct{
	ClientID int
	Username string 
}
