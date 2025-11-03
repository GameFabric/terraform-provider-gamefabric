package validators

import (
	"bytes"
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/gamefabric/gf-apicore/api/validation"
	"github.com/gamefabric/gf-apicore/runtime"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/tfutils"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// StoreValidator represents a GameFabric store that can validate runtime.Objects.
type StoreValidator interface {
	Validate(obj runtime.Object) validation.Errors
}

// Model represents a model that can be converted to its underlying runtime.Object.
type Model[T runtime.Object] interface {
	ToObject() T
}

// GameFabricValidatorRequest represents a request to validate a GameFabric configuration value.
type GameFabricValidatorRequest struct {
	Config      tfsdk.Config
	ConfigValue attr.Value
	Path        path.Path
	PathExpr    string
}

// GameFabricValidator defines the interface for validating GameFabric configuration values.
type GameFabricValidator interface {
	Validate(ctx context.Context, req GameFabricValidatorRequest) diag.Diagnostics
}

type gamefabricStoreValidator[T runtime.Object, M Model[T]] struct {
	val          StoreValidator
	pathExprOnce sync.Once
	pathExprs    map[string]path.Path
}

// NewGameFabricValidator creates a new gamefabricStoreValidator that wraps the given StoreValidator.
func NewGameFabricValidator[T runtime.Object, M Model[T]](fn func() StoreValidator) GameFabricValidator {
	return &gamefabricStoreValidator[T, M]{val: fn()}
}

func (v *gamefabricStoreValidator[T, M]) Validate(ctx context.Context, req GameFabricValidatorRequest) diag.Diagnostics {
	v.pathExprOnce.Do(func() {
		// The first time we run, we need to collect all path expressions from the schema.
		tfutils.WalkResourceSchema(req.Config.Schema.(rschema.Schema), func(attr rschema.Attribute, p path.Path) {
			vals := tfutils.Validators(attr)
			for _, val := range vals {
				if gfVal, ok := val.(gamefabricAttributeValidator); ok {
					if v.pathExprs == nil {
						v.pathExprs = make(map[string]path.Path)
					}
					v.pathExprs[gfVal.pathExpr] = p
				}
			}
		})
	})

	// If the value is not known, delay the validation until it is known.
	if !req.Config.Raw.IsFullyKnown() {
		return nil
	}

	var model M
	diags := req.Config.Get(ctx, &model)
	if diags.HasError() {
		return diags
	}

	p := completePathExpr(req.PathExpr, req.Path.Steps())
	for _, err := range v.val.Validate(model.ToObject()) {
		errMsg := err.Error()
		idx := strings.Index(errMsg, p)
		if idx == -1 || idx > len(errMsg)/2 {
			// Slight hack. It is possible for the path to appear in another error message,
			// but in general it should appear near the start of the message if it is relevant.
			// So if it is not found or is found too far in, we skip it.
			continue
		}
		errMsg = errMsg[idx+len(p)+1:]

		// Replace path expressions with actual paths including variable values for better diagnostics.
		// This checks all related path expressions to reduce the number of substitutions attempted.
		for pathExpr, tfPath := range v.pathExprs {
			if !relatedPathExprs(req.PathExpr, pathExpr) {
				continue
			}

			pe := completePathExpr(pathExpr, req.Path.Steps())
			if pe == p || !strings.Contains(errMsg, pe) {
				continue
			}
			ps := pathStringWithVars(tfPath.Steps(), req.Path.Steps())
			errMsg = strings.ReplaceAll(errMsg, pe, ps)
		}

		diags.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			errMsg,
			req.ConfigValue.String(),
		))
	}
	return diags
}

// This type of validator must satisfy all types.
var (
	_ validator.Bool    = gamefabricAttributeValidator{}
	_ validator.Float64 = gamefabricAttributeValidator{}
	_ validator.Float32 = gamefabricAttributeValidator{}
	_ validator.Int32   = gamefabricAttributeValidator{}
	_ validator.Int64   = gamefabricAttributeValidator{}
	_ validator.List    = gamefabricAttributeValidator{}
	_ validator.Map     = gamefabricAttributeValidator{}
	_ validator.Number  = gamefabricAttributeValidator{}
	_ validator.Object  = gamefabricAttributeValidator{}
	_ validator.Set     = gamefabricAttributeValidator{}
	_ validator.String  = gamefabricAttributeValidator{}
	_ validator.Dynamic = gamefabricAttributeValidator{}
)

type gamefabricAttributeValidator struct {
	val      GameFabricValidator
	pathExpr string
}

// PathExpr returns the GameFabric path expression.
func (v gamefabricAttributeValidator) PathExpr() string {
	return v.pathExpr
}

func (v gamefabricAttributeValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v gamefabricAttributeValidator) MarkdownDescription(_ context.Context) string {
	return "Validates the attribute using GameFabric configuration validation."
}

func (v gamefabricAttributeValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

func (v gamefabricAttributeValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

func (v gamefabricAttributeValidator) ValidateFloat32(ctx context.Context, req validator.Float32Request, resp *validator.Float32Response) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

func (v gamefabricAttributeValidator) ValidateInt32(ctx context.Context, req validator.Int32Request, resp *validator.Int32Response) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

func (v gamefabricAttributeValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

func (v gamefabricAttributeValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

func (v gamefabricAttributeValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

func (v gamefabricAttributeValidator) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

func (v gamefabricAttributeValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

func (v gamefabricAttributeValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

func (v gamefabricAttributeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

func (v gamefabricAttributeValidator) ValidateDynamic(ctx context.Context, req validator.DynamicRequest, resp *validator.DynamicResponse) {
	validateReq := GameFabricValidatorRequest{
		Config:      req.Config,
		ConfigValue: req.ConfigValue,
		Path:        req.Path,
		PathExpr:    v.pathExpr,
	}
	resp.Diagnostics.Append(v.val.Validate(ctx, validateReq)...)
}

// GFFieldBool creates a GameFabric attribute validator for boolean attributes.
func GFFieldBool(val GameFabricValidator, pathExpr string) validator.Bool {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

// GFFieldFloat64 creates a GameFabric attribute validator for float64 attributes.
func GFFieldFloat64(val GameFabricValidator, pathExpr string) validator.Float64 {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

// GFFieldFloat32 creates a GameFabric attribute validator for float32 attributes.
func GFFieldFloat32(val GameFabricValidator, pathExpr string) validator.Float32 {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

// GFFieldInt32 creates a GameFabric attribute validator for int32 attributes.
func GFFieldInt32(val GameFabricValidator, pathExpr string) validator.Int32 {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

// GFFieldInt64 creates a GameFabric attribute validator for int64 attributes.
func GFFieldInt64(val GameFabricValidator, pathExpr string) validator.Int64 {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

// GFFieldList creates a GameFabric attribute validator for list attributes.
func GFFieldList(val GameFabricValidator, pathExpr string) validator.List {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

// GFFieldMap creates a GameFabric attribute validator for map attributes.
func GFFieldMap(val GameFabricValidator, pathExpr string) validator.Map {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

// GFFieldNumber creates a GameFabric attribute validator for number attributes.
func GFFieldNumber(val GameFabricValidator, pathExpr string) validator.Number {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

// GFFieldObject creates a GameFabric attribute validator for object attributes.
func GFFieldObject(val GameFabricValidator, pathExpr string) validator.Object {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

// GFFieldSet creates a GameFabric attribute validator for set attributes.
func GFFieldSet(val GameFabricValidator, pathExpr string) validator.Set {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

// GFFieldString creates a GameFabric attribute validator for string attributes.
func GFFieldString(val GameFabricValidator, pathExpr string) validator.String {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

// GFFieldDynamic creates a GameFabric attribute validator for dynamic attributes.
func GFFieldDynamic(val GameFabricValidator, pathExpr string) validator.Dynamic {
	return gamefabricAttributeValidator{
		val:      val,
		pathExpr: pathExpr,
	}
}

func completePathExpr(pathExpr string, steps path.PathSteps) string {
	var stepVals []string
	for _, step := range steps {
		switch v := step.(type) {
		case path.PathStepElementKeyInt:
			stepVals = append(stepVals, strconv.Itoa(int(v)))
		case path.PathStepElementKeyString:
			stepVals = append(stepVals, string(v))
		}
	}

	n := strings.Count(pathExpr, "?")
	if n > len(stepVals) || n == 0 {
		return pathExpr
	}

	b := ([]byte)(pathExpr)
	for _, val := range stepVals {
		b = bytes.Replace(b, []byte("?"), []byte(val), 1)
	}
	return string(b)
}

func relatedPathExprs(target, expr string) bool {
	if idx := strings.LastIndex(expr, "?"); idx >= 0 {
		return strings.HasPrefix(target, expr[:idx])
	}

	// If there are no variables, it is very difficult to determine relation.
	// We should test them all.
	return true
}

func pathStringWithVars(dest, src path.PathSteps) string {
	var p path.PathSteps = make([]path.PathStep, 0, len(dest))
	for i, step := range dest {
		if _, ok := step.(path.PathStepAttributeName); ok || len(src) <= i {
			p.Append(step)
			continue
		}
		p.Append(src[i])
	}
	return p.String()
}
