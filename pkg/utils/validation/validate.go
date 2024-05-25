package utils

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var validate *validator.Validate

var translator ut.Translator

func init() {
	validate = validator.New()
	translator, _ = ut.New(en.New(), en.New()).GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, translator)
	translateOverride(translator)

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func Check(val any) error {
	if err := validate.Struct(val); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			msg := errs[0].Translate(translator)
			return errors.New(msg)
		}
	}
	return nil
}

func translateOverride(trans ut.Translator) {
	requiredTag := "required"
	validate.RegisterTranslation(requiredTag, trans, func(ut ut.Translator) error {
		return ut.Add(requiredTag, "{0} must have a value!", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(requiredTag, fe.Field())

		return t
	})

	colorTag := "hexcolor|rgb|rgba"
	validate.RegisterTranslation(colorTag, trans, func(ut ut.Translator) error {
		return ut.Add(colorTag, "{0} must be a valid color", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(colorTag, fe.Field())

		return t
	})
}
