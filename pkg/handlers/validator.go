package handlers

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func ValidateApplicant(valueObject interface{}) []Error {
	validate := validator.New()
	RegisterCustomValidations(validate)

	err := validate.Struct(valueObject)
	if err == nil {
		return nil
	}

	var errs []Error
	for _, fieldErr := range err.(validator.ValidationErrors) {
		errs = append(errs, Error{
			Name:  strings.ToLower(fieldErr.Field()),
			Value: buildErrorMessage(fieldErr),
		})
	}
	return errs
}

func buildErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("field %s is required", strings.ToLower(fe.Field()))
	case "uuid4":
		return fmt.Sprintf("field %s must be UUIDv4", strings.ToLower(fe.Field()))
	case "email":
		return fmt.Sprintf("field %s must contain a valid email", strings.ToLower(fe.Field()))
	case "url", "url_or_domain":
		return fmt.Sprintf("field %s must be a valid URL or domain", strings.ToLower(fe.Field()))
	case "datetime":
		return fmt.Sprintf("field %s must be date in YYYY-MM-DD format", strings.ToLower(fe.Field()))
	case "gt":
		return fmt.Sprintf("field %s must be more %s", strings.ToLower(fe.Field()), strings.ToLower(fe.Param()))
	case "e164":
		return fmt.Sprintf("field %s must be like +7...", strings.ToLower(fe.Field()))
	case "e164_plus7":
		return fmt.Sprintf("field %s must be like +7...", strings.ToLower(fe.Field()))
	}

	switch fe.Field() {
	case "Inn":
		if fe.Tag() == "inn_check" || fe.Tag() == "inn_check_by_subject" {
			return "invalid inn: must be 10 or 12 digits with valid checksum"
		}
		return "the inn length must be 10 or 12 characters."
	case "OrganizationName":
		if fe.Tag() == "org_name_required" {
			return "field organizationName is required for Organization"
		}
	case "TaxPayerType", "TaxRate", "HasVAT":
		if fe.Tag() == "tax_fields_required" {
			return fmt.Sprintf("field %s is required for IP and Organization", strings.ToLower(fe.Field()))
		}
		if fe.Tag() == "self_employed_tax_check" {
			return "taxPayerType must be НПД for Самозанятый"
		}
	case "Bik":
		if fe.Tag() == "bik_validation" {
			return "invalid bik: must be 9 digits"
		}
	case "Kpp":
		if fe.Tag() == "kpp_validation" {
			return "invalid kpp: must be 9 digits"
		}
	case "AccountNumber":
		if fe.Tag() == "account_validation" {
			return "invalid account number: must be 20 digits with valid checksum"
		}
	case "CorrespondentAccount":
		if fe.Tag() == "correspondent_account_validation" {
			return "invalid correspondent account: must be 20 digits with valid checksum"
		}
	}

	return fmt.Sprintf("field %s is invalid: %s", strings.ToLower(fe.Field()), strings.ToLower(fe.Tag()))
}

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}(:[0-9]{1,5})?(/.*)?$`)

func validateUUID4(fl validator.FieldLevel) bool {
	field := fl.Field()

	if field.Type() == reflect.TypeOf(uuid.UUID{}) {
		uuidValue := field.Interface().(uuid.UUID)
		if uuidValue == uuid.Nil {
			return false
		}
		return uuidValue.Version() == 4
	}

	if field.Kind() == reflect.String {
		uuidStr := field.String()
		if uuidStr == "" {
			return false
		}
		uuidValue, err := uuid.Parse(uuidStr)
		if err != nil {
			return false
		}
		return uuidValue.Version() == 4
	}

	return false
}

func RegisterCustomValidations(validate *validator.Validate) {
	validate.RegisterValidation("uuid4", validateUUID4)
	validate.RegisterValidation("url_or_domain", validateURLOrDomain)
	validate.RegisterValidation("e164_plus7", IsE164WithPlus7)
	validate.RegisterValidation("inn_check", validateINN)
	validate.RegisterValidation("inn_check_by_subject", validateINNByBusinessSubject)
	validate.RegisterValidation("org_name_required", validateOrganizationNameRequired)
	validate.RegisterValidation("tax_fields_required", validateTaxFieldsRequired)
	validate.RegisterValidation("self_employed_tax_check", validateSelfEmployedTax)
	validate.RegisterValidation("bik_validation", validateBIK)
	validate.RegisterValidation("kpp_validation", validateKPP)
	validate.RegisterValidation("account_validation", validateAccount)
	validate.RegisterValidation("correspondent_account_validation", validateCorrespondentAccount)
}

func validateURLOrDomain(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return false
	}

	parsedURL, err := url.Parse(value)
	if err != nil {
		return false
	}

	if parsedURL.Scheme != "" {
		return parsedURL.Scheme == "http" || parsedURL.Scheme == "https" && parsedURL.Host != ""
	}

	return domainRegex.MatchString(value)
}

var e164Plus7Regex = regexp.MustCompile(`^\+7\d{10,14}$`)

func IsE164WithPlus7(fl validator.FieldLevel) bool {
	fieldValue := fl.Field().String()
	return e164Plus7Regex.MatchString(fieldValue)
}

func validateINN(fl validator.FieldLevel) bool {
	inn := fl.Field().String()

	if len(inn) != 10 && len(inn) != 12 {
		return false
	}

	for _, r := range inn {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	calcCheckDigit := func(nums []int, weights []int) int {
		sum := 0
		for i, w := range weights {
			sum += nums[i] * w
		}
		return (sum % 11) % 10
	}

	nums := make([]int, len(inn))
	for i := range inn {
		nums[i] = int(inn[i] - '0')
	}

	if len(inn) == 10 {
		weights := []int{2, 4, 10, 3, 5, 9, 4, 6, 8}
		return calcCheckDigit(nums, weights) == nums[9]
	}

	if len(inn) == 12 {
		weights11 := []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
		weights12 := []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}
		return calcCheckDigit(nums, weights11) == nums[10] &&
			calcCheckDigit(nums, weights12) == nums[11]
	}

	return false
}

func validateINNByBusinessSubject(fl validator.FieldLevel) bool {
	inn := fl.Field().String()

	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	businessSubjectField := parent.FieldByName("BusinessSubject")
	if !businessSubjectField.IsValid() {
		return false
	}

	businessSubject := businessSubjectField.String()

	var expectedLength int
	if businessSubject == "Организация" {
		expectedLength = 10
	} else if businessSubject == "Самозанятый" || businessSubject == "ИП" {
		expectedLength = 12
	} else {
		return validateINN(fl)
	}

	if len(inn) != expectedLength {
		return false
	}

	return validateINN(fl)
}

func validateOrganizationNameRequired(fl validator.FieldLevel) bool {
	orgName := fl.Field().String()

	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	businessSubjectField := parent.FieldByName("BusinessSubject")
	if !businessSubjectField.IsValid() {
		return false
	}

	businessSubject := businessSubjectField.String()

	if businessSubject == "Организация" {
		return len(orgName) > 0 && len(orgName) <= 200
	}

	return true
}

func validateTaxFieldsRequired(fl validator.FieldLevel) bool {
	fieldValue := fl.Field().String()

	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	businessSubjectField := parent.FieldByName("BusinessSubject")
	if !businessSubjectField.IsValid() {
		return false
	}

	businessSubject := businessSubjectField.String()

	if businessSubject == "Самозанятый" {
		return true
	}

	if businessSubject == "ИП" || businessSubject == "Организация" {
		return len(fieldValue) > 0
	}

	return true
}

func validateSelfEmployedTax(fl validator.FieldLevel) bool {
	taxPayerType := fl.Field().String()

	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	businessSubjectField := parent.FieldByName("BusinessSubject")
	if !businessSubjectField.IsValid() {
		return false
	}

	businessSubject := businessSubjectField.String()

	if businessSubject == "Самозанятый" {
		return taxPayerType == "НПД"
	}

	return true
}

func validateBIK(fl validator.FieldLevel) bool {
	bik := fl.Field().String()

	if len(bik) != 9 {
		return false
	}

	for _, r := range bik {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func validateKPP(fl validator.FieldLevel) bool {
	kpp := fl.Field().String()

	if len(kpp) != 9 {
		return false
	}

	for _, r := range kpp {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

func validateAccount(fl validator.FieldLevel) bool {
	account := fl.Field().String()

	if len(account) != 20 {
		return false
	}

	for _, r := range account {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	bikField := parent.FieldByName("Bik")
	if !bikField.IsValid() {
		return false
	}

	bik := bikField.String()
	if len(bik) != 9 {
		return false
	}

	return validateAccountChecksum(bik, account)
}

func validateCorrespondentAccount(fl validator.FieldLevel) bool {
	corrAccount := fl.Field().String()

	if len(corrAccount) != 20 {
		return false
	}

	for _, r := range corrAccount {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	bikField := parent.FieldByName("Bik")
	if !bikField.IsValid() {
		return false
	}

	bik := bikField.String()
	if len(bik) != 9 {
		return false
	}

	return validateCorrespondentAccountChecksum(bik, corrAccount)
}

func validateAccountChecksum(bik, account string) bool {
	bikPart := bik[4:7]

	checkString := bikPart + account

	weights := []int{7, 1, 3}

	sum := 0
	for i := 0; i < len(checkString); i++ {
		digit := int(checkString[i] - '0')
		weight := weights[i%len(weights)]
		sum += digit * weight
	}

	return sum%10 == 0
}

func validateCorrespondentAccountChecksum(bik, corrAccount string) bool {
	bikPart := bik[6:9]

	checkString := bikPart + corrAccount

	weights := []int{7, 3, 1}

	sum := 0
	for i := 0; i < len(checkString); i++ {
		digit := int(checkString[i] - '0')
		weight := weights[i%len(weights)]
		sum += digit * weight
	}

	return sum%10 == 0
}
