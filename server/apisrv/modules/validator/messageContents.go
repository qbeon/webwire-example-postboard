package validator

// MessageContents implements the Validator interface
func (vld *validator) MessageContents(contents string) error {
	length := len(contents)
	if uint32(length) < vld.MinMessageLength {
		return Errorf(
			"invalid message contents: too short (%d / %d)",
			length,
			vld.MinMessageLength,
		)
	} else if uint32(length) > vld.MaxMessageLength {
		return Errorf(
			"invalid message contents: too long (%d / %d)",
			length,
			vld.MaxMessageLength,
		)
	}
	return nil
}