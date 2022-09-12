package utils

import "github.com/cosmos/cosmos-sdk/codec"

// ValidatedProtoMarshaler is a ProtoMarshaler that can also be validated
type ValidatedProtoMarshaler interface {
	codec.ProtoMarshaler
	ValidateBasic() error
}

type validatedProtoMarshaler struct {
	codec.ProtoMarshaler
	validate func() error
}

func (v validatedProtoMarshaler) ValidateBasic() error {
	return v.validate()
}

// WithValidation adds a ValidateBasic function to an existing ProtoMarshaler
func WithValidation(value codec.ProtoMarshaler, validate func() error) ValidatedProtoMarshaler {
	return validatedProtoMarshaler{
		ProtoMarshaler: value,
		validate:       validate,
	}
}

// NoValidation wraps a ProtoMarshaler so it can be used by the store without actually adding any validation
func NoValidation(value codec.ProtoMarshaler) ValidatedProtoMarshaler {
	return validatedProtoMarshaler{
		ProtoMarshaler: value,
		validate:       func() error { return nil },
	}
}
