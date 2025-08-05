package structs

import (
	"encoding/json"
)

// MarshalMap omits zero values when marshal to json.
type MarshalMap map[string]any

func (m *MarshalMap) MarshalJSON() ([]byte, error) {
	mapWithOmit := make(map[string]any, len(*m))
	for k, v := range *m {
		switch val := v.(type) {
		case int:
			if val > 0 {
				mapWithOmit[k] = v
			}
		case string:
			if len(val) > 0 {
				mapWithOmit[k] = v
			}
		case bool:
			mapWithOmit[k] = v
		default:
		}
	}

	return json.Marshal(mapWithOmit)
}

func (m *MarshalMap) UnmarshalJSON(data []byte) error {
	if *m == nil {
		*m = make(map[string]any)
	}
	if err := json.Unmarshal(data, (*map[string]any)(m)); err != nil {
		return err
	}

	mapWithOmit := make(map[string]any, len(*m))
	for k, v := range *m {
		switch val := v.(type) {
		case float64:
			if val > 0 {
				mapWithOmit[k] = int(val)
			}
		case string:
			if len(val) > 0 {
				mapWithOmit[k] = val
			}
		case bool:
			mapWithOmit[k] = v
		default:
		}
	}

	*m = mapWithOmit
	return nil
}
