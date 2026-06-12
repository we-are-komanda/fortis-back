package handlers

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestValidateINN(t *testing.T) {
	validate := validator.New()
	RegisterCustomValidations(validate)

	tests := []struct {
		name    string
		inn     string
		wantErr bool
	}{
		{
			name:    "valid 10-digit INN",
			inn:     "7707083893",
			wantErr: false,
		},
		{
			name:    "valid 12-digit INN",
			inn:     "500100732259",
			wantErr: false,
		},
		{
			name:    "invalid INN - wrong length",
			inn:     "12345",
			wantErr: true,
		},
		{
			name:    "invalid INN - wrong checksum",
			inn:     "7707083890",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Inn string `validate:"inn_check"`
			}

			test := TestStruct{Inn: tt.inn}
			err := validate.Struct(test)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateINN(%s) error = %v, wantErr %v", tt.inn, err, tt.wantErr)
			}
		})
	}
}

func TestValidateINNByBusinessSubject(t *testing.T) {
	validate := validator.New()
	RegisterCustomValidations(validate)

	tests := []struct {
		name            string
		businessSubject string
		inn             string
		wantErr         bool
	}{
		{
			name:            "valid 10-digit INN for Organization",
			businessSubject: "Организация",
			inn:             "7707083893",
			wantErr:         false,
		},
		{
			name:            "invalid 12-digit INN for Organization",
			businessSubject: "Организация",
			inn:             "500100732259",
			wantErr:         true,
		},
		{
			name:            "valid 12-digit INN for IP",
			businessSubject: "ИП",
			inn:             "500100732259",
			wantErr:         false,
		},
		{
			name:            "invalid 10-digit INN for IP",
			businessSubject: "ИП",
			inn:             "7707083893",
			wantErr:         true,
		},
		{
			name:            "valid 12-digit INN for SelfEmployed",
			businessSubject: "Самозанятый",
			inn:             "500100732259",
			wantErr:         false,
		},
		{
			name:            "invalid 10-digit INN for SelfEmployed",
			businessSubject: "Самозанятый",
			inn:             "7707083893",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				BusinessSubject string `validate:"required"`
				Inn             string `validate:"inn_check_by_subject"`
			}

			test := TestStruct{
				BusinessSubject: tt.businessSubject,
				Inn:             tt.inn,
			}
			err := validate.Struct(test)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateINNByBusinessSubject(BusinessSubject=%s, Inn=%s) error = %v, wantErr %v", tt.businessSubject, tt.inn, err, tt.wantErr)
			}
		})
	}
}

func TestValidateOrganizationNameRequired(t *testing.T) {
	validate := validator.New()
	RegisterCustomValidations(validate)

	tests := []struct {
		name            string
		businessSubject string
		orgName         string
		wantErr         bool
	}{
		{
			name:            "valid org name for Organization",
			businessSubject: "Организация",
			orgName:         "ООО Ромашка",
			wantErr:         false,
		},
		{
			name:            "invalid empty org name for Organization",
			businessSubject: "Организация",
			orgName:         "",
			wantErr:         true,
		},
		{
			name:            "valid empty org name for IP",
			businessSubject: "ИП",
			orgName:         "",
			wantErr:         false,
		},
		{
			name:            "valid empty org name for SelfEmployed",
			businessSubject: "Самозанятый",
			orgName:         "",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				BusinessSubject  string `validate:"required"`
				OrganizationName string `validate:"org_name_required"`
			}

			test := TestStruct{
				BusinessSubject:  tt.businessSubject,
				OrganizationName: tt.orgName,
			}
			err := validate.Struct(test)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOrganizationNameRequired(BusinessSubject=%s, OrgName=%s) error = %v, wantErr %v", tt.businessSubject, tt.orgName, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTaxFieldsRequired(t *testing.T) {
	validate := validator.New()
	RegisterCustomValidations(validate)

	tests := []struct {
		name            string
		businessSubject string
		taxPayerType    string
		taxRate         string
		hasVAT          string
		wantErrTaxType  bool
		wantErrTaxRate  bool
		wantErrHasVAT   bool
	}{
		{
			name:            "valid tax fields for Organization",
			businessSubject: "Организация",
			taxPayerType:    "ОСНО",
			taxRate:         "6%",
			hasVAT:          "Есть",
			wantErrTaxType:  false,
			wantErrTaxRate:  false,
			wantErrHasVAT:   false,
		},
		{
			name:            "invalid empty tax fields for Organization",
			businessSubject: "Организация",
			taxPayerType:    "",
			taxRate:         "",
			hasVAT:          "",
			wantErrTaxType:  true,
			wantErrTaxRate:  true,
			wantErrHasVAT:   true,
		},
		{
			name:            "valid empty tax fields for SelfEmployed",
			businessSubject: "Самозанятый",
			taxPayerType:    "",
			taxRate:         "",
			hasVAT:          "",
			wantErrTaxType:  false,
			wantErrTaxRate:  false,
			wantErrHasVAT:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				BusinessSubject string `validate:"required"`
				TaxPayerType    string `validate:"tax_fields_required"`
				TaxRate         string `validate:"tax_fields_required"`
				HasVAT          string `validate:"tax_fields_required"`
			}

			test := TestStruct{
				BusinessSubject: tt.businessSubject,
				TaxPayerType:    tt.taxPayerType,
				TaxRate:         tt.taxRate,
				HasVAT:          tt.hasVAT,
			}
			err := validate.Struct(test)

			hasTaxTypeErr := err != nil && containsFieldError(err, "TaxPayerType")
			hasTaxRateErr := err != nil && containsFieldError(err, "TaxRate")
			hasVATErr := err != nil && containsFieldError(err, "HasVAT")

			if hasTaxTypeErr != tt.wantErrTaxType {
				t.Errorf("TaxPayerType error = %v, wantErr %v", hasTaxTypeErr, tt.wantErrTaxType)
			}
			if hasTaxRateErr != tt.wantErrTaxRate {
				t.Errorf("TaxRate error = %v, wantErr %v", hasTaxRateErr, tt.wantErrTaxRate)
			}
			if hasVATErr != tt.wantErrHasVAT {
				t.Errorf("HasVAT error = %v, wantErr %v", hasVATErr, tt.wantErrHasVAT)
			}
		})
	}
}

func TestValidateSelfEmployedTax(t *testing.T) {
	validate := validator.New()
	RegisterCustomValidations(validate)

	tests := []struct {
		name            string
		businessSubject string
		taxPayerType    string
		wantErr         bool
	}{
		{
			name:            "valid НПД for SelfEmployed",
			businessSubject: "Самозанятый",
			taxPayerType:    "НПД",
			wantErr:         false,
		},
		{
			name:            "invalid ОСНО for SelfEmployed",
			businessSubject: "Самозанятый",
			taxPayerType:    "ОСНО",
			wantErr:         true,
		},
		{
			name:            "valid ОСНО for Organization",
			businessSubject: "Организация",
			taxPayerType:    "ОСНО",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				BusinessSubject string `validate:"required"`
				TaxPayerType    string `validate:"self_employed_tax_check"`
			}

			test := TestStruct{
				BusinessSubject: tt.businessSubject,
				TaxPayerType:    tt.taxPayerType,
			}
			err := validate.Struct(test)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSelfEmployedTax(BusinessSubject=%s, TaxPayerType=%s) error = %v, wantErr %v", tt.businessSubject, tt.taxPayerType, err, tt.wantErr)
			}
		})
	}
}

func containsFieldError(err error, fieldName string) bool {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		for _, validationErr := range validationErrs {
			if validationErr.Field() == fieldName {
				return true
			}
		}
	}
	return false
}
