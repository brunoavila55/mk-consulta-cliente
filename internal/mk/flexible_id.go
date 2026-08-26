package mk

import "encoding/json"

// FlexibleID carrega um identificador vindo da API da MK que pode ser
// serializado como string ou como número, dependendo do endpoint. O valor é
// devolvido no mesmo formato em que chegou, para não quebrar contrato com
// quem consome a resposta desta API (ex.: chatbot esperando o mesmo shape).
type FlexibleID struct {
	raw      string
	isString bool
	empty    bool
}

func EmptyID() FlexibleID {
	return FlexibleID{empty: true}
}

func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*f = FlexibleID{empty: true}
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*f = FlexibleID{raw: s, isString: true}
		return nil
	}
	*f = FlexibleID{raw: string(data)}
	return nil
}

func (f FlexibleID) MarshalJSON() ([]byte, error) {
	if f.empty {
		return []byte(`""`), nil
	}
	if f.isString {
		return json.Marshal(f.raw)
	}
	return []byte(f.raw), nil
}

// String devolve o valor bruto, usado para montar as URLs de consulta.
func (f FlexibleID) String() string {
	return f.raw
}
