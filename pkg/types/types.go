package types

func StringValueFromPointer(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
