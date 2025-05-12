package cyber

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Validator 验证器接口
type Validator interface {
	// Validate 验证数据
	Validate(data interface{}) error
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors 验证错误列表
type ValidationErrors []ValidationError

// Error 实现error接口
func (ve ValidationErrors) Error() string {
	var msgs []string
	for _, err := range ve {
		msgs = append(msgs, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(msgs, "; ")
}

// DefaultValidator 默认验证器
type DefaultValidator struct{}

// Validate 验证数据
func (v *DefaultValidator) Validate(data interface{}) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors.New("validation only works on structs")
	}

	typ := val.Type()
	var errs ValidationErrors

	// 遍历结构体字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)

		// 递归验证嵌套结构体
		if field.Kind() == reflect.Struct {
			if err := v.Validate(field.Interface()); err != nil {
				if validErrs, ok := err.(ValidationErrors); ok {
					errs = append(errs, validErrs...)
				} else {
					errs = append(errs, ValidationError{
						Field:   typeField.Name,
						Message: err.Error(),
					})
				}
			}
			continue
		}

		// 获取validation标签
		validTag := typeField.Tag.Get("valid")
		if validTag == "" {
			continue
		}

		// 解析验证规则
		rules := strings.Split(validTag, ",")
		for _, rule := range rules {
			parts := strings.Split(rule, "=")
			var ruleType, ruleValue string

			ruleType = parts[0]
			if len(parts) > 1 {
				ruleValue = parts[1]
			}

			// 应用验证规则
			var err error
			switch ruleType {
			case "required":
				err = validateRequired(field)
			case "min":
				err = validateMin(field, ruleValue)
			case "max":
				err = validateMax(field, ruleValue)
			case "email":
				err = validateEmail(field)
			case "pattern":
				err = validatePattern(field, ruleValue)
			}

			if err != nil {
				errs = append(errs, ValidationError{
					Field:   typeField.Name,
					Message: err.Error(),
				})
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// validateRequired 验证必填
func validateRequired(val reflect.Value) error {
	if isEmptyValue(val) {
		return errors.New("field is required")
	}
	return nil
}

// validateMin 验证最小值
func validateMin(val reflect.Value, minStr string) error {
	min, err := strconv.Atoi(minStr)
	if err != nil {
		return fmt.Errorf("invalid min value: %s", minStr)
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int() < int64(min) {
			return fmt.Errorf("value must be at least %d", min)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val.Uint() < uint64(min) {
			return fmt.Errorf("value must be at least %d", min)
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() < float64(min) {
			return fmt.Errorf("value must be at least %d", min)
		}
	case reflect.String:
		if len(val.String()) < min {
			return fmt.Errorf("length must be at least %d", min)
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if val.Len() < min {
			return fmt.Errorf("length must be at least %d", min)
		}
	default:
		return fmt.Errorf("min rule not applicable to type %s", val.Type())
	}

	return nil
}

// validateMax 验证最大值
func validateMax(val reflect.Value, maxStr string) error {
	max, err := strconv.Atoi(maxStr)
	if err != nil {
		return fmt.Errorf("invalid max value: %s", maxStr)
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int() > int64(max) {
			return fmt.Errorf("value must be at most %d", max)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val.Uint() > uint64(max) {
			return fmt.Errorf("value must be at most %d", max)
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() > float64(max) {
			return fmt.Errorf("value must be at most %d", max)
		}
	case reflect.String:
		if len(val.String()) > max {
			return fmt.Errorf("length must be at most %d", max)
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if val.Len() > max {
			return fmt.Errorf("length must be at most %d", max)
		}
	default:
		return fmt.Errorf("max rule not applicable to type %s", val.Type())
	}

	return nil
}

// validateEmail 验证邮箱格式
func validateEmail(val reflect.Value) error {
	if val.Kind() != reflect.String {
		return fmt.Errorf("email validation only applies to string type")
	}

	email := val.String()
	if email == "" {
		return nil // 允许空值，使用required规则来验证必填
	}

	// 邮箱正则表达式
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(pattern, email)
	if err != nil {
		return err
	}

	if !matched {
		return errors.New("invalid email format")
	}

	return nil
}

// validatePattern 验证正则表达式
func validatePattern(val reflect.Value, pattern string) error {
	if val.Kind() != reflect.String {
		return fmt.Errorf("pattern validation only applies to string type")
	}

	str := val.String()
	if str == "" {
		return nil // 允许空值，使用required规则来验证必填
	}

	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return fmt.Errorf("invalid pattern: %s", pattern)
	}

	if !matched {
		return fmt.Errorf("value does not match pattern: %s", pattern)
	}

	return nil
}

// isEmptyValue 判断值是否为空
func isEmptyValue(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.String:
		return val.Len() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return val.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return val.IsNil()
	}
	return false
}

// Bind 绑定请求数据并验证
func (c *Context) Bind(obj interface{}) error {
	// 根据Content-Type解析请求数据
	contentType := c.Request.Header.Get("Content-Type")

	// 处理JSON数据
	if strings.Contains(contentType, "application/json") {
		decoder := json.NewDecoder(c.Request.Body)
		if err := decoder.Decode(obj); err != nil {
			return err
		}
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// 处理表单数据
		if err := c.Request.ParseForm(); err != nil {
			return err
		}
		// 将表单数据绑定到结构体
		// 这里需要实现表单数据到结构体的映射
		// TODO: 实现表单数据绑定
	} else if strings.Contains(contentType, "multipart/form-data") {
		// 处理multipart表单数据
		err := c.Request.ParseMultipartForm(32 << 20) // 32MB
		if err != nil {
			return err
		}
		// 将表单数据绑定到结构体
		// TODO: 实现multipart表单数据绑定
	}

	// 验证数据
	validator := &DefaultValidator{}
	return validator.Validate(obj)
}
