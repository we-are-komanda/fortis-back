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

	// Если поле имеет специальное имя (например, "Inn"), проверяем его отдельно
	// Это нужно для кастомных сообщений об ошибках для специфичных полей
	switch fe.Field() {
	case "Inn":
		// Если ошибка связана с проверкой контрольной суммы ИНН
		if fe.Tag() == "inn_check" || fe.Tag() == "inn_check_by_subject" {
			return "invalid inn: must be 10 or 12 digits with valid checksum"
		}
		// Для других ошибок валидации поля Inn (например, "len", "required")
		// возвращаем общее сообщение о длине ИНН
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

// validateUUID4 проверяет, что значение является UUID версии 4
// Поддерживает типы uuid.UUID и string
func validateUUID4(fl validator.FieldLevel) bool {
	field := fl.Field()

	// Проверяем, является ли значение типом uuid.UUID
	if field.Type() == reflect.TypeOf(uuid.UUID{}) {
		uuidValue := field.Interface().(uuid.UUID)
		// Проверяем, что UUID не является нулевым значением
		if uuidValue == uuid.Nil {
			return false
		}
		// Проверяем версию UUID (должна быть 4)
		return uuidValue.Version() == 4
	}

	// Если это строка, пытаемся распарсить и проверить
	if field.Kind() == reflect.String {
		uuidStr := field.String()
		if uuidStr == "" {
			return false
		}
		uuidValue, err := uuid.Parse(uuidStr)
		if err != nil {
			return false
		}
		// Проверяем версию UUID (должна быть 4)
		return uuidValue.Version() == 4
	}

	// Для других типов возвращаем false
	return false
}

func RegisterCustomValidations(validate *validator.Validate) {
	// Перезаписываем встроенный валидатор uuid4 для поддержки типа uuid.UUID
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

	// Проверка длины
	if len(inn) != 10 && len(inn) != 12 {
		return false
	}

	// Проверка что только цифры
	for _, r := range inn {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	// Вспомогательная функция
	calcCheckDigit := func(nums []int, weights []int) int {
		sum := 0
		for i, w := range weights {
			sum += nums[i] * w
		}
		return (sum % 11) % 10
	}

	// Преобразуем в массив чисел
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

// validateINNByBusinessSubject проверяет ИНН с учетом типа BusinessSubject
// Для Организации должна быть длина 10, для СЗ и ИП - 12
func validateINNByBusinessSubject(fl validator.FieldLevel) bool {
	inn := fl.Field().String()

	// Получаем родительскую структуру для доступа к BusinessSubject
	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	businessSubjectField := parent.FieldByName("BusinessSubject")
	if !businessSubjectField.IsValid() {
		return false
	}

	businessSubject := businessSubjectField.String()

	// Проверяем длину в зависимости от типа
	var expectedLength int
	if businessSubject == "Организация" {
		expectedLength = 10
	} else if businessSubject == "Самозанятый" || businessSubject == "ИП" {
		expectedLength = 12
	} else {
		// Если BusinessSubject не определен, используем стандартную проверку
		return validateINN(fl)
	}

	if len(inn) != expectedLength {
		return false
	}

	// Проверяем контрольную сумму
	return validateINN(fl)
}

// validateOrganizationNameRequired проверяет, что organizationName обязателен только для Организации
func validateOrganizationNameRequired(fl validator.FieldLevel) bool {
	orgName := fl.Field().String()

	// Получаем родительскую структуру для доступа к BusinessSubject
	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	businessSubjectField := parent.FieldByName("BusinessSubject")
	if !businessSubjectField.IsValid() {
		return false
	}

	businessSubject := businessSubjectField.String()

	// Для Организации поле обязательно
	if businessSubject == "Организация" {
		return len(orgName) > 0 && len(orgName) <= 200
	}

	// Для СЗ и ИП поле необязательно (игнорируем)
	return true
}

// validateTaxFieldsRequired проверяет, что TaxPayerType, TaxRate, HasVAT обязательны для ИП и Организаций
func validateTaxFieldsRequired(fl validator.FieldLevel) bool {
	fieldValue := fl.Field().String()

	// Получаем родительскую структуру для доступа к BusinessSubject
	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	businessSubjectField := parent.FieldByName("BusinessSubject")
	if !businessSubjectField.IsValid() {
		return false
	}

	businessSubject := businessSubjectField.String()

	// Для Самозанятого эти поля необязательны
	if businessSubject == "Самозанятый" {
		return true
	}

	// Для ИП и Организаций поля обязательны
	if businessSubject == "ИП" || businessSubject == "Организация" {
		return len(fieldValue) > 0
	}

	return true
}

// validateSelfEmployedTax проверяет, что для Самозанятого TaxPayerType должен быть "НПД"
func validateSelfEmployedTax(fl validator.FieldLevel) bool {
	taxPayerType := fl.Field().String()

	// Получаем родительскую структуру для доступа к BusinessSubject
	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return false
	}

	businessSubjectField := parent.FieldByName("BusinessSubject")
	if !businessSubjectField.IsValid() {
		return false
	}

	businessSubject := businessSubjectField.String()

	// Проверка только для Самозанятого
	if businessSubject == "Самозанятый" {
		return taxPayerType == "НПД"
	}

	// Для других типов проверка не требуется
	return true
}

// validateBIK проверяет БИК (9 цифр)
func validateBIK(fl validator.FieldLevel) bool {
	bik := fl.Field().String()

	// Проверка длины
	if len(bik) != 9 {
		return false
	}

	// Проверка что только цифры
	for _, r := range bik {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

// validateKPP проверяет КПП (9 цифр)
func validateKPP(fl validator.FieldLevel) bool {
	kpp := fl.Field().String()

	// Проверка длины
	if len(kpp) != 9 {
		return false
	}

	// Проверка что только цифры
	for _, r := range kpp {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

// validateAccount проверяет расчетный счет (20 цифр) с контрольной суммой
func validateAccount(fl validator.FieldLevel) bool {
	account := fl.Field().String()

	// Проверка длины
	if len(account) != 20 {
		return false
	}

	// Проверка что только цифры
	for _, r := range account {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	// Получаем БИК из родительской структуры
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

	// Вычисляем контрольную сумму
	return validateAccountChecksum(bik, account)
}

// validateCorrespondentAccount проверяет корреспондентский счет (20 цифр) с контрольной суммой
func validateCorrespondentAccount(fl validator.FieldLevel) bool {
	corrAccount := fl.Field().String()

	// Проверка длины
	if len(corrAccount) != 20 {
		return false
	}

	// Проверка что только цифры
	for _, r := range corrAccount {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	// Получаем БИК из родительской структуры
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

	// Вычисляем контрольную сумму для корреспондентского счета
	return validateCorrespondentAccountChecksum(bik, corrAccount)
}

// validateAccountChecksum проверяет контрольную сумму расчетного счета
// Алгоритм: берем цифры 5-7 БИК (индексы 4-6), конкатенируем с номером счета
// Используем веса [7, 1, 3] повторяющиеся
// Проверяем, что сумма делится на 10 без остатка
func validateAccountChecksum(bik, account string) bool {
	// Берем цифры 5-7 БИК (индексы 4-6)
	bikPart := bik[4:7]

	// Создаем строку для проверки: 3 цифры БИК + 20 цифр счета = 23 цифры
	checkString := bikPart + account

	// Веса для расчета: [7, 1, 3] повторяющиеся
	weights := []int{7, 1, 3}

	// Вычисляем сумму произведений для всех цифр
	sum := 0
	for i := 0; i < len(checkString); i++ {
		digit := int(checkString[i] - '0')
		weight := weights[i%len(weights)]
		sum += digit * weight
	}

	// Проверяем, что сумма делится на 10 без остатка
	return sum%10 == 0
}

// validateCorrespondentAccountChecksum проверяет контрольную сумму корреспондентского счета
// Алгоритм: берем последние 3 цифры БИК (индексы 6-8), конкатенируем с номером счета
// Используем веса [7, 3, 1] повторяющиеся
// Проверяем, что сумма делится на 10 без остатка
func validateCorrespondentAccountChecksum(bik, corrAccount string) bool {
	// Берем последние 3 цифры БИК (индексы 6-8)
	bikPart := bik[6:9]

	// Создаем строку для проверки: 3 цифры БИК + 20 цифр счета = 23 цифры
	checkString := bikPart + corrAccount

	// Веса для расчета: [7, 3, 1] повторяющиеся
	weights := []int{7, 3, 1}

	// Вычисляем сумму произведений для всех цифр
	sum := 0
	for i := 0; i < len(checkString); i++ {
		digit := int(checkString[i] - '0')
		weight := weights[i%len(weights)]
		sum += digit * weight
	}

	// Проверяем, что сумма делится на 10 без остатка
	return sum%10 == 0
}

