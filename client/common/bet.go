package common

import (
	"strings"
)

const DELIMITER = "|"

type Bet struct {
	agencia string
	nombre string
	apellido string
	documento string
	nacimiento string
	numero string
}

func NewBet(agencia string, nombre string, apellido string, documento string, nacimiento string, numero string) *Bet {
	return &Bet{
		agencia: agencia,
		nombre: nombre,
		apellido: apellido,
		documento: documento,
		nacimiento: nacimiento,
		numero: numero,
	}
}

func (b *Bet) serialize() []byte {
	dataFields := []string{b.agencia, b.nombre, b.apellido, b.documento, b.nacimiento, b.numero}
	data := strings.Join(dataFields, DELIMITER)
	
	dataBytes := []byte(data)
	return dataBytes
}