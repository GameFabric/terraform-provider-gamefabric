package tfutils

import (
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Validators returns all validators defined on the given attribute.
// The interfaces can be found in Terraforms fwxschema/attribute_validators.go.
// Reproduced here as the fwxschema package is internal.
//
//nolint:cyclop // Splitting this function up would add unnecessary complexity.
func Validators(attr rschema.Attribute) []any {
	var validators []any
	switch v := attr.(type) {
	case interface{ BoolValidators() []validator.Bool }:
		for _, val := range v.BoolValidators() {
			validators = append(validators, val)
		}
	case interface{ Float32Validators() []validator.Float32 }:
		for _, val := range v.Float32Validators() {
			validators = append(validators, val)
		}
	case interface{ Float64Validators() []validator.Float64 }:
		for _, val := range v.Float64Validators() {
			validators = append(validators, val)
		}
	case interface{ Int32Validators() []validator.Int32 }:
		for _, val := range v.Int32Validators() {
			validators = append(validators, val)
		}
	case interface{ Int64Validators() []validator.Int64 }:
		for _, val := range v.Int64Validators() {
			validators = append(validators, val)
		}
	case interface{ ListValidators() []validator.List }:
		for _, val := range v.ListValidators() {
			validators = append(validators, val)
		}
	case interface{ MapValidators() []validator.Map }:
		for _, val := range v.MapValidators() {
			validators = append(validators, val)
		}
	case interface{ NumberValidators() []validator.Number }:
		for _, val := range v.NumberValidators() {
			validators = append(validators, val)
		}
	case interface{ ObjectValidators() []validator.Object }:
		for _, val := range v.ObjectValidators() {
			validators = append(validators, val)
		}
	case interface{ SetValidators() []validator.Set }:
		for _, val := range v.SetValidators() {
			validators = append(validators, val)
		}
	case interface{ StringValidators() []validator.String }:
		for _, val := range v.StringValidators() {
			validators = append(validators, val)
		}
	case interface{ DynamicValidators() []validator.Dynamic }:
		for _, val := range v.DynamicValidators() {
			validators = append(validators, val)
		}
	}
	return validators
}
