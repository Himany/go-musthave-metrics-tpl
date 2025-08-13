package utils

func SetIntIfUnset(envSet map[string]bool, envKey string, cfgValue *int, flagValue int) {
	if envSet[envKey] {
		return
	}
	*cfgValue = flagValue
}

func SetStringIfUnset(envSet map[string]bool, envKey string, cfgValue *string, flagValue string) {
	if envSet[envKey] {
		return
	}
	*cfgValue = flagValue
}

func SetBoolIfUnset(envSet map[string]bool, envKey string, cfgValue *bool, flagValue bool) {
	if envSet[envKey] {
		return
	}
	*cfgValue = flagValue
}
