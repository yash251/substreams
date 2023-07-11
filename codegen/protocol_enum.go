// Code generated by go-enum DO NOT EDIT.
// Version: 0.5.6
// Revision: 97611fddaa414f53713597918c5e954646cb8623
// Build Date: 2023-03-26T21:38:06Z
// Built By: goreleaser

package codegen

import (
	"fmt"
	"strings"
)

const (
	// ProtocolEthereum is a Protocol of type Ethereum.
	ProtocolEthereum Protocol = iota
	// ProtocolOther is a Protocol of type Other.
	ProtocolOther
)

var ErrInvalidProtocol = fmt.Errorf("not a valid Protocol, try [%s]", strings.Join(_ProtocolNames, ", "))

const _ProtocolName = "EthereumOther"

var _ProtocolNames = []string{
	_ProtocolName[0:8],
	_ProtocolName[8:13],
}

// ProtocolNames returns a list of possible string values of Protocol.
func ProtocolNames() []string {
	tmp := make([]string, len(_ProtocolNames))
	copy(tmp, _ProtocolNames)
	return tmp
}

var _ProtocolMap = map[Protocol]string{
	ProtocolEthereum: _ProtocolName[0:8],
	ProtocolOther:    _ProtocolName[8:13],
}

// String implements the Stringer interface.
func (x Protocol) String() string {
	if str, ok := _ProtocolMap[x]; ok {
		return str
	}
	return fmt.Sprintf("Protocol(%d)", x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x Protocol) IsValid() bool {
	_, ok := _ProtocolMap[x]
	return ok
}

var _ProtocolValue = map[string]Protocol{
	_ProtocolName[0:8]:                   ProtocolEthereum,
	strings.ToLower(_ProtocolName[0:8]):  ProtocolEthereum,
	_ProtocolName[8:13]:                  ProtocolOther,
	strings.ToLower(_ProtocolName[8:13]): ProtocolOther,
}

// ParseProtocol attempts to convert a string to a Protocol.
func ParseProtocol(name string) (Protocol, error) {
	if x, ok := _ProtocolValue[name]; ok {
		return x, nil
	}
	// Case insensitive parse, do a separate lookup to prevent unnecessary cost of lowercasing a string if we don't need to.
	if x, ok := _ProtocolValue[strings.ToLower(name)]; ok {
		return x, nil
	}
	return Protocol(0), fmt.Errorf("%s is %w", name, ErrInvalidProtocol)
}

// MarshalText implements the text marshaller method.
func (x Protocol) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *Protocol) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseProtocol(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
