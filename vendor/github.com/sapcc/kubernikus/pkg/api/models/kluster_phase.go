// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/validate"
)

// KlusterPhase kluster phase
// swagger:model KlusterPhase
type KlusterPhase string

const (

	// KlusterPhasePending captures enum value "Pending"
	KlusterPhasePending KlusterPhase = "Pending"

	// KlusterPhaseCreating captures enum value "Creating"
	KlusterPhaseCreating KlusterPhase = "Creating"

	// KlusterPhaseRunning captures enum value "Running"
	KlusterPhaseRunning KlusterPhase = "Running"

	// KlusterPhaseTerminating captures enum value "Terminating"
	KlusterPhaseTerminating KlusterPhase = "Terminating"
)

// for schema
var klusterPhaseEnum []interface{}

func init() {
	var res []KlusterPhase
	if err := json.Unmarshal([]byte(`["Pending","Creating","Running","Terminating"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		klusterPhaseEnum = append(klusterPhaseEnum, v)
	}
}

func (m KlusterPhase) validateKlusterPhaseEnum(path, location string, value KlusterPhase) error {
	if err := validate.Enum(path, location, value, klusterPhaseEnum); err != nil {
		return err
	}
	return nil
}

// Validate validates this kluster phase
func (m KlusterPhase) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateKlusterPhaseEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}