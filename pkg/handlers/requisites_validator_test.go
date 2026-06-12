package handlers

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestBIKValidation(t *testing.T) {
	validate := validator.New()
	RegisterCustomValidations(validate)

	tests := []struct {
		name  string
		bik   string
		valid bool
	}{
		{"valid BIK", "044525225", true},
		{"valid BIK 2", "045004525", true},
		{"invalid - too short", "04452522", false},
		{"invalid - too long", "0445252255", false},
		{"invalid - contains letters", "04452522A", false},
		{"invalid - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Bik string `validate:"bik_validation"`
			}
			err := validate.Struct(TestStruct{Bik: tt.bik})
			if (err == nil) != tt.valid {
				t.Errorf("BIK validation for %s: got %v, want %v", tt.bik, err == nil, tt.valid)
			}
		})
	}
}

func TestKPPValidation(t *testing.T) {
	validate := validator.New()
	RegisterCustomValidations(validate)

	tests := []struct {
		name  string
		kpp   string
		valid bool
	}{
		{"valid KPP", "770501001", true},
		{"valid KPP all zeros", "000000000", true},
		{"invalid - too short", "77050100", false},
		{"invalid - too long", "7705010011", false},
		{"invalid - contains letters", "77050100A", false},
		{"invalid - empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Kpp string `validate:"kpp_validation"`
			}
			err := validate.Struct(TestStruct{Kpp: tt.kpp})
			if (err == nil) != tt.valid {
				t.Errorf("KPP validation for %s: got %v, want %v", tt.kpp, err == nil, tt.valid)
			}
		})
	}
}

func TestAccountValidation(t *testing.T) {
	validate := validator.New()
	RegisterCustomValidations(validate)

	tests := []struct {
		name    string
		bik     string
		account string
		valid   bool
	}{
		{"valid account", "044525225", "40817810099910004312", true},
		{"invalid - wrong length", "044525225", "4081781009991000431", false},
		{"invalid - contains letters", "044525225", "4081781009991000431A", false},
		{"invalid - empty", "044525225", "", false},
		{"invalid - BIK too short", "04452522", "40817810099910004312", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Bik           string `validate:"bik_validation"`
				AccountNumber string `validate:"account_validation"`
			}
			testStruct := TestStruct{
				Bik:           tt.bik,
				AccountNumber: tt.account,
			}
			err := validate.Struct(testStruct)
			if (err == nil) != tt.valid {
				t.Errorf("Account validation for BIK %s, Account %s: got %v, want %v", tt.bik, tt.account, err == nil, tt.valid)
			}
		})
	}
}

func TestCorrespondentAccountValidation(t *testing.T) {
	validate := validator.New()
	RegisterCustomValidations(validate)

	tests := []struct {
		name                 string
		bik                  string
		correspondentAccount string
		valid                bool
	}{
		{"valid correspondent account", "044525225", "30101810400000000225", true},
		{"invalid - wrong length", "044525225", "3010181040000000022", false},
		{"invalid - contains letters", "044525225", "3010181040000000022A", false},
		{"invalid - empty", "044525225", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Bik                  string `validate:"bik_validation"`
				CorrespondentAccount string `validate:"correspondent_account_validation"`
			}
			testStruct := TestStruct{
				Bik:                  tt.bik,
				CorrespondentAccount: tt.correspondentAccount,
			}
			err := validate.Struct(testStruct)
			if (err == nil) != tt.valid {
				t.Errorf("Correspondent account validation for BIK %s, Account %s: got %v, want %v", tt.bik, tt.correspondentAccount, err == nil, tt.valid)
			}
		})
	}
}

func TestPaymentRequisitesValidation(t *testing.T) {
	validate := validator.New()
	RegisterCustomValidations(validate)

	type PaymentRequisites struct {
		Recipient            string `validate:"required,min=3,max=160"`
		BankName             string `validate:"required,min=3,max=160"`
		AccountNumber        string `validate:"required,len=20,account_validation"`
		CorrespondentAccount string `validate:"required,len=20,correspondent_account_validation"`
		Bik                  string `validate:"required,len=9,bik_validation"`
		Kpp                  string `validate:"required,len=9,kpp_validation"`
	}

	tests := []struct {
		name        string
		requisites  PaymentRequisites
		valid       bool
		description string
	}{
		{
			name: "valid - real data from production issue",
			requisites: PaymentRequisites{
				Recipient:            "ООО Ромашка",
				BankName:             "ПАО Сбербанк",
				AccountNumber:        "40817810099910004312",
				CorrespondentAccount: "30101810400000000225",
				Bik:                  "044525225",
				Kpp:                  "770501001",
			},
			valid:       true,
			description: "Реальные данные из production, которые ранее не проходили валидацию",
		},
		{
			name: "invalid - wrong account number checksum",
			requisites: PaymentRequisites{
				Recipient:            "ООО Ромашка",
				BankName:             "ПАО Сбербанк",
				AccountNumber:        "40817810099910004313",
				CorrespondentAccount: "30101810400000000225",
				Bik:                  "044525225",
				Kpp:                  "770501001",
			},
			valid:       false,
			description: "Расчетный счет с неправильной контрольной суммой",
		},
		{
			name: "invalid - wrong correspondent account checksum",
			requisites: PaymentRequisites{
				Recipient:            "ООО Ромашка",
				BankName:             "ПАО Сбербанк",
				AccountNumber:        "40817810099910004312",
				CorrespondentAccount: "30101810400000000226",
				Bik:                  "044525225",
				Kpp:                  "770501001",
			},
			valid:       false,
			description: "Корреспондентский счет с неправильной контрольной суммой",
		},
		{
			name: "invalid - wrong BIK",
			requisites: PaymentRequisites{
				Recipient:            "ООО Ромашка",
				BankName:             "ПАО Сбербанк",
				AccountNumber:        "40817810099910004312",
				CorrespondentAccount: "30101810400000000225",
				Bik:                  "044525226",
				Kpp:                  "770501001",
			},
			valid:       false,
			description: "БИК не соответствует контрольным суммам счетов",
		},
		{
			name: "invalid - missing required fields",
			requisites: PaymentRequisites{
				Recipient:            "",
				BankName:             "ПАО Сбербанк",
				AccountNumber:        "40817810099910004312",
				CorrespondentAccount: "30101810400000000225",
				Bik:                  "044525225",
				Kpp:                  "770501001",
			},
			valid:       false,
			description: "Отсутствует обязательное поле получателя",
		},
		{
			name: "invalid - account number wrong length",
			requisites: PaymentRequisites{
				Recipient:            "ООО Ромашка",
				BankName:             "ПАО Сбербанк",
				AccountNumber:        "4081781009991000431",
				CorrespondentAccount: "30101810400000000225",
				Bik:                  "044525225",
				Kpp:                  "770501001",
			},
			valid:       false,
			description: "Расчетный счет неправильной длины",
		},
		{
			name: "invalid - correspondent account wrong length",
			requisites: PaymentRequisites{
				Recipient:            "ООО Ромашка",
				BankName:             "ПАО Сбербанк",
				AccountNumber:        "40817810099910004312",
				CorrespondentAccount: "3010181040000000022",
				Bik:                  "044525225",
				Kpp:                  "770501001",
			},
			valid:       false,
			description: "Корреспондентский счет неправильной длины",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.requisites)
			isValid := err == nil

			if isValid != tt.valid {
				if err != nil {
					t.Errorf("Validation failed for %s: got errors %v, want valid=%v. Description: %s",
						tt.name, err, tt.valid, tt.description)
				} else {
					t.Errorf("Validation passed for %s: expected it to fail. Description: %s",
						tt.name, tt.description)
				}
			} else {
				if tt.valid {
					t.Logf("✓ Successfully validated requisites: BIK=%s, Account=%s, CorrAccount=%s",
						tt.requisites.Bik, tt.requisites.AccountNumber, tt.requisites.CorrespondentAccount)
				}
			}
		})
	}
}
