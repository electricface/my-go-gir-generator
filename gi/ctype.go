package gi

import (
	"errors"
	"strings"
)

type CType struct {
	Name       string
	IsConst    bool
	IsUnsigned bool
	NumStar    int
}

func ParseCType(ctype string) (*CType, error) {
	numStar := strings.Count(ctype, "*")
	ctype = strings.TrimRight(ctype, "*")

	var (
		isConst    bool
		isUnsigned bool
	)

	fields := strings.Fields(ctype)
	if len(fields) == 0 {
		return nil, errors.New("fields empty")
	}

	for _, f := range fields[:len(fields)-1] {
		if f == "const" {
			isConst = true
		} else if f == "unsigned" {
			isUnsigned = true
		} else {
			return nil, errors.New("unknown type mod " + f)
		}
	}

	name := fields[len(fields)-1]

	return &CType{
		Name:       name,
		NumStar:    numStar,
		IsConst:    isConst,
		IsUnsigned: isUnsigned,
	}, nil
}

func (ct *CType) CgoNotation() string {
	ret := strings.Repeat("*", ct.NumStar)
	name := ct.Name
	if ct.IsUnsigned {
		if ct.Name == "int" {
			name = "uint"
		} else {
			panic("unsupported unsigned type " + ct.Name)
		}
	}

	ret = ret + "C." + name
	return ret
}

func (ct *CType) Elem() *CType {
	numStar := ct.NumStar - 1
	if numStar < 0 {
		panic("assert failed numStr >= 0")
	}

	return &CType{
		Name:       ct.Name,
		NumStar:    numStar,
		IsConst:    ct.IsConst,
		IsUnsigned: ct.IsUnsigned,
	}
}
