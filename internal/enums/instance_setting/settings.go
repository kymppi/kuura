package instance_setting

import (
	"errors"
	"strconv"
)

type InstanceSetting int64

const (
	InternalServiceId InstanceSetting = iota
)

func (s InstanceSetting) String() string {
	switch s {
	case InternalServiceId:
		return "INTERNAL_SERVICE_ID"
	default:
		return "UNKNOWN"
	}
}

func (s InstanceSetting) IsValid() bool {
	return s == InternalServiceId
}

func From(v interface{}) (InstanceSetting, error) {
	switch val := v.(type) {
	case string:
		switch val {
		case "INTERNAL_SERVICE_ID":
			return InternalServiceId, nil
		default:
			// Try to parse the string as an int64 and convert to InstanceSetting
			num, err := strconv.ParseInt(val, 10, 64)
			if err == nil {
				return InstanceSetting(num), nil
			}
			return InternalServiceId, errors.New("invalid settings key")
		}
	case int:
		return InstanceSetting(val), nil
	case int64:
		return InstanceSetting(val), nil
	case float64:
		return InstanceSetting(int64(val)), nil
	default:
		return InternalServiceId, errors.New("invalid settings key")
	}
}

func AllPossibleStatuses() []InstanceSetting {
	return []InstanceSetting{
		InternalServiceId,
	}
}
