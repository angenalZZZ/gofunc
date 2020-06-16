package f

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const maxURLRuneCount = 2083
const minURLRuneCount = 3
const RF3339WithoutZone = "2006-01-02T15:04:05"

var (
	fieldsRequiredByDefault bool
	nilPtrAllowedByRequired = false
	notNumberRegexp         = regexp.MustCompile("[^0-9]+")
	whiteSpacesAndMinus     = regexp.MustCompile(`[\s-]+`)
	paramsRegexp            = regexp.MustCompile(`\(.*\)$`)
)

// ValidateMap use validation map for fields.
// result will be equal to `false` if there are any errors.
// m is the validation map in the form
// map[string]interface{}{"name":"required,alpha","address":map[string]interface{}{"line1":"required,alphanum"}}
func ValidateMap(s map[string]interface{}, m map[string]interface{}) (bool, error) {
	if s == nil {
		return true, nil
	}
	result := true
	var err error
	var errs Errors
	var index int
	val := reflect.ValueOf(s)
	for key, value := range s {
		presentResult := true
		validator, ok := m[key]
		if !ok {
			presentResult = false
			var err error
			err = fmt.Errorf("all map keys has to be present in the validation map; got %s", key)
			err = PrependPathToErrors(err, key)
			errs = append(errs, err)
		}
		valueField := reflect.ValueOf(value)
		mapResult := true
		typeResult := true
		structResult := true
		resultField := true
		switch subValidator := validator.(type) {
		case map[string]interface{}:
			var err error
			if v, ok := value.(map[string]interface{}); !ok {
				mapResult = false
				err = fmt.Errorf("map validator has to be for the map type only; got %s", valueField.Type().String())
				err = PrependPathToErrors(err, key)
				errs = append(errs, err)
			} else {
				mapResult, err = ValidateMap(v, subValidator)
				if err != nil {
					mapResult = false
					err = PrependPathToErrors(err, key)
					errs = append(errs, err)
				}
			}
		case string:
			if (valueField.Kind() == reflect.Struct ||
				(valueField.Kind() == reflect.Ptr && valueField.Elem().Kind() == reflect.Struct)) &&
				subValidator != "-" {
				var err error
				structResult, err = ValidateStruct(valueField.Interface())
				if err != nil {
					err = PrependPathToErrors(err, key)
					errs = append(errs, err)
				}
			}
			resultField, err = typeCheck(valueField, reflect.StructField{
				Name:      key,
				PkgPath:   "",
				Type:      val.Type(),
				Tag:       reflect.StructTag(fmt.Sprintf("%s:%q", RxTagName, subValidator)),
				Offset:    0,
				Index:     []int{index},
				Anonymous: false,
			}, val, nil)
			if err != nil {
				errs = append(errs, err)
			}
		case nil:
			// already handlerd when checked before
		default:
			typeResult = false
			err = fmt.Errorf("map validator has to be either map[string]interface{} or string; got %s", valueField.Type().String())
			err = PrependPathToErrors(err, key)
			errs = append(errs, err)
		}
		result = result && presentResult && typeResult && resultField && structResult && mapResult
		index++
	}
	// check required keys
	requiredResult := true
	for key, value := range m {
		if schema, ok := value.(string); ok {
			tags := parseTagIntoMap(schema)
			if required, ok := tags["required"]; ok {
				if _, ok := s[key]; !ok {
					requiredResult = false
					if required.customErrorMessage != "" {
						err = Error{key, fmt.Errorf(required.customErrorMessage), true, "required", []string{}}
					} else {
						err = Error{key, fmt.Errorf("required field missing"), false, "required", []string{}}
					}
					errs = append(errs, err)
				}
			}
		}
	}

	if len(errs) > 0 {
		err = errs
	}
	return result && requiredResult, err
}

// ValidateStruct use tags for fields.
// result will be equal to `false` if there are any errors.
// to-do currently there is no guarantee that errors will be returned in predictable order (tests may to fail)
func ValidateStruct(s interface{}) (bool, error) {
	if s == nil {
		return true, nil
	}
	result := true
	var err error
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	// we only accept structs
	if val.Kind() != reflect.Struct {
		return false, fmt.Errorf("function only accepts structs; got %s", val.Kind())
	}
	var errs Errors
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		if typeField.PkgPath != "" {
			continue // Private field
		}
		structResult := true
		if valueField.Kind() == reflect.Interface {
			valueField = valueField.Elem()
		}
		if (valueField.Kind() == reflect.Struct ||
			(valueField.Kind() == reflect.Ptr && valueField.Elem().Kind() == reflect.Struct)) &&
			typeField.Tag.Get(RxTagName) != "-" {
			var err error
			structResult, err = ValidateStruct(valueField.Interface())
			if err != nil {
				err = PrependPathToErrors(err, typeField.Name)
				errs = append(errs, err)
			}
		}
		resultField, err2 := typeCheck(valueField, typeField, val, nil)
		if err2 != nil {

			// Replace structure name with JSON name if there is a tag on the variable
			jsonTag := toJSONName(typeField.Tag.Get("json"))
			if jsonTag != "" {
				switch jsonError := err2.(type) {
				case Error:
					jsonError.Name = jsonTag
					err2 = jsonError
				case Errors:
					for i2, err3 := range jsonError {
						switch customErr := err3.(type) {
						case Error:
							customErr.Name = jsonTag
							jsonError[i2] = customErr
						}
					}

					err2 = jsonError
				}
			}

			errs = append(errs, err2)
		}
		result = result && resultField && structResult
	}
	if len(errs) > 0 {
		err = errs
	}
	return result, err
}

// parseTagIntoMap parses a struct tag `valid:required~Some error message,length(2|3)` into map[string]string{"required": "Some error message", "length(2|3)": ""}
func parseTagIntoMap(tag string) tagOptionsMap {
	optionsMap := make(tagOptionsMap)
	options := strings.Split(tag, ",")

	for i, option := range options {
		option = strings.TrimSpace(option)

		validationOptions := strings.Split(option, "~")
		if !isValidTag(validationOptions[0]) {
			continue
		}
		if len(validationOptions) == 2 {
			optionsMap[validationOptions[0]] = tagOption{validationOptions[0], validationOptions[1], i}
		} else {
			optionsMap[validationOptions[0]] = tagOption{validationOptions[0], "", i}
		}
	}
	return optionsMap
}

func checkRequired(v reflect.Value, t reflect.StructField, options tagOptionsMap) (bool, error) {
	if nilPtrAllowedByRequired {
		k := v.Kind()
		if (k == reflect.Ptr || k == reflect.Interface) && v.IsNil() {
			return true, nil
		}
	}

	if requiredOption, isRequired := options["required"]; isRequired {
		if len(requiredOption.customErrorMessage) > 0 {
			return false, Error{t.Name, fmt.Errorf(requiredOption.customErrorMessage), true, "required", []string{}}
		}
		return false, Error{t.Name, fmt.Errorf("non zero value required"), false, "required", []string{}}
	} else if _, isOptional := options["optional"]; fieldsRequiredByDefault && !isOptional {
		return false, Error{t.Name, fmt.Errorf("Missing required field"), false, "required", []string{}}
	}
	// not required and empty is valid
	return true, nil
}

func typeCheck(v reflect.Value, t reflect.StructField, o reflect.Value, options tagOptionsMap) (isValid bool, resultErr error) {
	if !v.IsValid() {
		return false, nil
	}

	tag := t.Tag.Get(RxTagName)

	// Check if the field should be ignored
	switch tag {
	case "":
		if v.Kind() != reflect.Slice && v.Kind() != reflect.Map {
			if !fieldsRequiredByDefault {
				return true, nil
			}
			return false, Error{t.Name, fmt.Errorf("All fields are required to at least have one validation defined"), false, "required", []string{}}
		}
	case "-":
		return true, nil
	}

	isRootType := false
	if options == nil {
		isRootType = true
		options = parseTagIntoMap(tag)
	}

	if !isFieldSet(v) {
		// an empty value is not validated, check only required
		isValid, resultErr = checkRequired(v, t, options)
		for key := range options {
			delete(options, key)
		}
		return isValid, resultErr
	}

	var customTypeErrors Errors
	optionsOrder := options.orderedKeys()
	for _, validatorName := range optionsOrder {
		validatorStruct := options[validatorName]
		if validatefunc, ok := CustomTypeTagMap.Get(validatorName); ok {
			delete(options, validatorName)

			if result := validatefunc(v.Interface(), o.Interface()); !result {
				if len(validatorStruct.customErrorMessage) > 0 {
					customTypeErrors = append(customTypeErrors, Error{Name: t.Name, Err: TruncatingErrorf(validatorStruct.customErrorMessage, fmt.Sprint(v), validatorName), CustomErrorMessageExists: true, Validator: stripParams(validatorName)})
					continue
				}
				customTypeErrors = append(customTypeErrors, Error{Name: t.Name, Err: fmt.Errorf("%s does not validate as %s", fmt.Sprint(v), validatorName), CustomErrorMessageExists: false, Validator: stripParams(validatorName)})
			}
		}
	}

	if customTypeErrors != nil && len(customTypeErrors.Errors()) > 0 {
		return false, customTypeErrors
	}

	if isRootType {
		// Ensure that we've checked the value by all specified validators before report that the value is valid
		defer func() {
			delete(options, "optional")
			delete(options, "required")

			if isValid && resultErr == nil && len(options) != 0 {
				optionsOrder := options.orderedKeys()
				for _, validator := range optionsOrder {
					isValid = false
					resultErr = Error{t.Name, fmt.Errorf(
						"the following validator is invalid or can't be applied to the field: %q", validator), false, stripParams(validator), []string{}}
					return
				}
			}
		}()
	}

	for _, validatorSpec := range optionsOrder {
		validatorStruct := options[validatorSpec]
		var negate bool
		validator := validatorSpec
		customMsgExists := len(validatorStruct.customErrorMessage) > 0

		// Check whether the tag looks like '!something' or 'something'
		if validator[0] == '!' {
			validator = validator[1:]
			negate = true
		}

		// Check for interface param validators
		for key, value := range InterfaceParamTagRegexMap {
			ps := value.FindStringSubmatch(validator)
			if len(ps) == 0 {
				continue
			}

			validatefunc, ok := InterfaceParamTagMap[key]
			if !ok {
				continue
			}

			delete(options, validatorSpec)

			field := fmt.Sprint(v)
			if result := validatefunc(v.Interface(), ps[1:]...); (!result && !negate) || (result && negate) {
				if customMsgExists {
					return false, Error{t.Name, TruncatingErrorf(validatorStruct.customErrorMessage, field, validator), customMsgExists, stripParams(validatorSpec), []string{}}
				}
				if negate {
					return false, Error{t.Name, fmt.Errorf("%s does validate as %s", field, validator), customMsgExists, stripParams(validatorSpec), []string{}}
				}
				return false, Error{t.Name, fmt.Errorf("%s does not validate as %s", field, validator), customMsgExists, stripParams(validatorSpec), []string{}}
			}
		}
	}

	switch v.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.String:
		// for each tag option check the map of validator functions
		for _, validatorSpec := range optionsOrder {
			validatorStruct := options[validatorSpec]
			var negate bool
			validator := validatorSpec
			customMsgExists := len(validatorStruct.customErrorMessage) > 0

			// Check whether the tag looks like '!something' or 'something'
			if validator[0] == '!' {
				validator = validator[1:]
				negate = true
			}

			// Check for param validators
			for key, value := range ParamTagRegexMap {
				ps := value.FindStringSubmatch(validator)
				if len(ps) == 0 {
					continue
				}

				validatefunc, ok := ParamTagMap[key]
				if !ok {
					continue
				}

				delete(options, validatorSpec)

				switch v.Kind() {
				case reflect.String,
					reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
					reflect.Float32, reflect.Float64:

					field := fmt.Sprint(v) // make value into string, then validate with regex
					if result := validatefunc(field, ps[1:]...); (!result && !negate) || (result && negate) {
						if customMsgExists {
							return false, Error{t.Name, TruncatingErrorf(validatorStruct.customErrorMessage, field, validator), customMsgExists, stripParams(validatorSpec), []string{}}
						}
						if negate {
							return false, Error{t.Name, fmt.Errorf("%s does validate as %s", field, validator), customMsgExists, stripParams(validatorSpec), []string{}}
						}
						return false, Error{t.Name, fmt.Errorf("%s does not validate as %s", field, validator), customMsgExists, stripParams(validatorSpec), []string{}}
					}
				default:
					// type not yet supported, fail
					return false, Error{t.Name, fmt.Errorf("validator %s doesn't support kind %s", validator, v.Kind()), false, stripParams(validatorSpec), []string{}}
				}
			}

			if validatefunc, ok := TagMap[validator]; ok {
				delete(options, validatorSpec)

				switch v.Kind() {
				case reflect.String,
					reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
					reflect.Float32, reflect.Float64:
					field := fmt.Sprint(v) // make value into string, then validate with regex
					if result := validatefunc(field); !result && !negate || result && negate {
						if customMsgExists {
							return false, Error{t.Name, TruncatingErrorf(validatorStruct.customErrorMessage, field, validator), customMsgExists, stripParams(validatorSpec), []string{}}
						}
						if negate {
							return false, Error{t.Name, fmt.Errorf("%s does validate as %s", field, validator), customMsgExists, stripParams(validatorSpec), []string{}}
						}
						return false, Error{t.Name, fmt.Errorf("%s does not validate as %s", field, validator), customMsgExists, stripParams(validatorSpec), []string{}}
					}
				default:
					//Not Yet Supported Types (Fail here!)
					err := fmt.Errorf("validator %s doesn't support kind %s for value %v", validator, v.Kind(), v)
					return false, Error{t.Name, err, false, stripParams(validatorSpec), []string{}}
				}
			}
		}
		return true, nil
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return false, &UnsupportedTypeError{v.Type()}
		}
		var sv stringValues
		sv = v.MapKeys()
		sort.Sort(sv)
		result := true
		for i, k := range sv {
			var resultItem bool
			var err error
			if v.MapIndex(k).Kind() != reflect.Struct {
				resultItem, err = typeCheck(v.MapIndex(k), t, o, options)
				if err != nil {
					return false, err
				}
			} else {
				resultItem, err = ValidateStruct(v.MapIndex(k).Interface())
				if err != nil {
					err = PrependPathToErrors(err, t.Name+"."+sv[i].Interface().(string))
					return false, err
				}
			}
			result = result && resultItem
		}
		return result, nil
	case reflect.Slice, reflect.Array:
		result := true
		for i := 0; i < v.Len(); i++ {
			var resultItem bool
			var err error
			if v.Index(i).Kind() != reflect.Struct {
				resultItem, err = typeCheck(v.Index(i), t, o, options)
				if err != nil {
					return false, err
				}
			} else {
				resultItem, err = ValidateStruct(v.Index(i).Interface())
				if err != nil {
					err = PrependPathToErrors(err, t.Name+"."+strconv.Itoa(i))
					return false, err
				}
			}
			result = result && resultItem
		}
		return result, nil
	case reflect.Interface:
		// If the value is an interface then encode its element
		if v.IsNil() {
			return true, nil
		}
		return ValidateStruct(v.Interface())
	case reflect.Ptr:
		// If the value is a pointer then check its element
		if v.IsNil() {
			return true, nil
		}
		return typeCheck(v.Elem(), t, o, options)
	case reflect.Struct:
		return true, nil
	default:
		return false, &UnsupportedTypeError{v.Type()}
	}
}

func stripParams(validatorString string) string {
	return paramsRegexp.ReplaceAllString(validatorString, "")
}

// isFieldSet returns false for nil pointers, interfaces, maps, and slices. For all other values, it returns true.
func isFieldSet(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Interface, reflect.Ptr:
		return !v.IsNil()
	}

	return true
}

// ErrorByField returns error for specified field of the struct
// validated by ValidateStruct or empty string if there are no errors
// or this field doesn't exists or doesn't have any errors.
func ErrorByField(e error, field string) string {
	if e == nil {
		return ""
	}
	return ErrorsByField(e)[field]
}

// ErrorsByField returns map of errors of the struct validated
// by ValidateStruct or empty map if there are no errors.
func ErrorsByField(e error) map[string]string {
	m := make(map[string]string)
	if e == nil {
		return m
	}
	// prototype for ValidateStruct

	switch e.(type) {
	case Error:
		m[e.(Error).Name] = e.(Error).Err.Error()
	case Errors:
		for _, item := range e.(Errors).Errors() {
			n := ErrorsByField(item)
			for k, v := range n {
				m[k] = v
			}
		}
	}

	return m
}

// Error returns string equivalent for reflect.Type
func (e *UnsupportedTypeError) Error() string {
	return "validator: unsupported type: " + e.Type.String()
}

func (sv stringValues) Len() int           { return len(sv) }
func (sv stringValues) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }
func (sv stringValues) Less(i, j int) bool { return sv.get(i) < sv.get(j) }
func (sv stringValues) get(i int) string   { return sv[i].String() }

// SetFieldsRequiredByDefault causes validation to fail when struct fields
// do not include validations or are not explicitly marked as exempt (using `valid:"-"` or `valid:"email,optional"`).
// This struct definition will fail govalidator.ValidateStruct() (and the field values do not matter):
//     type exampleStruct struct {
//         Name  string ``
//         Email string `valid:"email"`
// This, however, will only fail when Email is empty or an invalid email address:
//     type exampleStruct2 struct {
//         Name  string `valid:"-"`
//         Email string `valid:"email"`
// Lastly, this will only fail when Email is an invalid email address but not when it's empty:
//     type exampleStruct2 struct {
//         Name  string `valid:"-"`
//         Email string `valid:"email,optional"`
func SetFieldsRequiredByDefault(value bool) {
	fieldsRequiredByDefault = value
}

// SetNilPtrAllowedByRequired causes validation to pass for nil ptrs when a field is set to required.
// The validation will still reject ptr fields in their zero value state. Example with this enabled:
//     type exampleStruct struct {
//         Name  *string `valid:"required"`
// With `Name` set to "", this will be considered invalid input and will cause a validation error.
// With `Name` set to nil, this will be considered valid by validation.
// By default this is disabled.
func SetNilPtrAllowedByRequired(value bool) {
	nilPtrAllowedByRequired = value
}
