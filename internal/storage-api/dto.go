// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2021 ARDUINO SA (http://www.arduino.cc/)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package storageapi

import (
	"time"

	"github.com/arduino/arduino-cli/table"
	"github.com/gofrs/uuid"
)

type (
	ImportCustomTemplateResponse struct {
		Message    string `json:"message"`
		Name       string `json:"name"`
		TemplateId string `json:"template_id"`
	}
	TemplateEntry struct {
		TemplateId string `json:"template_id"`
		Name       string `json:"name"`
		CreatedAt  string `json:"created_at"`
		UpdatedAt  string `json:"updated_at"`
	}
	TemplatesListResponse struct {
		Templates []TemplateEntry `json:"templates"`
	}
)

func (r *TemplatesListResponse) Data() interface{} {
	return r.Templates
}

func (r *TemplatesListResponse) String() string {
	if len(r.Templates) == 0 {
		return ""
	}
	t := table.New()
	t.SetHeader("Template ID", "Name", "Created At", "Updated At")

	// Now print the table
	for _, tem := range r.Templates {
		line := []any{tem.TemplateId, tem.Name, formatHumanReadableTs(tem.CreatedAt), formatHumanReadableTs(tem.UpdatedAt)}
		t.AddRow(line...)
	}

	return t.Render()
}

func formatHumanReadableTs(ts string) string {
	if ts == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ts
	}
	return parsed.Format(time.RFC3339)
}

type TemplateDescribeResponse struct {
	// CompatibleBoards List of devices compatible with the template
	CompatibleBoards *[]string `json:"compatible_boards,omitempty"`

	// CreatedAt Template creation date/time
	CreatedAt *time.Time `json:"created_at,omitempty"`

	// DashboardTemplates List of dashboard templates
	DashboardTemplates *[]DashboardTemplate `json:"dashboard_templates,omitempty"`

	// DeletedAt Template soft deletion date/time
	DeletedAt *time.Time `json:"deleted_at,omitempty"`

	// Description Description of the template
	Description string `json:"description"`

	// ImageLinks Optional list of images to be included in the template
	ImageLinks *[]string `json:"image_links,omitempty"`

	// Name Template name
	Name string `json:"name" validate:"required,max=128"`

	// OrganizationId Organization identifier
	OrganizationId *uuid.UUID `json:"organization_id,omitempty"`
	TemplateId     uuid.UUID  `json:"template_id" validate:"required,uuid"`

	// ThingTemplates List of thing templates
	ThingTemplates *[]ThingTemplate `json:"thing_templates,omitempty"`

	// TriggerTemplates List of trigger templates
	TriggerTemplates *[]TriggerTemplate `json:"trigger_templates,omitempty"`

	// UpdatedAt Template update date/time
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	// UserId User identifier
	UserId *uuid.UUID `json:"user_id,omitempty" validate:"uuid"`
}

type ThingTemplate struct {
	// Id Template ID
	Id string `json:"id"`

	// Name Name
	Name string `json:"name"`

	// Tags Tags
	Tags *[]Tag `json:"tags,omitempty"`

	// Variables Thing variables
	Variables *[]Variable `json:"variables,omitempty"`

	// WebhookUri Webhook URI
	WebhookUri *string `json:"webhook_uri,omitempty"`
}

type Variable struct {
	// Id Id
	Id *string `json:"id,omitempty"`

	// Name Name
	Name *string `json:"name,omitempty"`

	// Permission Permission
	Permission *string `json:"permission,omitempty"`

	// Type Type
	Type *string `json:"type,omitempty"`

	// UpdateParameter Update parameter
	UpdateParameter *float32 `json:"update_parameter,omitempty"`

	// UpdateStrategy Update strategy
	UpdateStrategy *string `json:"update_strategy,omitempty"`

	// VariableName Variable name
	VariableName *string `json:"variable_name,omitempty"`
}

type Tag struct {
	// Key Key
	Key string `json:"key"`

	// Value Value
	Value string `json:"value"`
}

type DashboardTemplate struct {
	// Id Template ID
	Id string `json:"id"`

	// Name Name
	Name string `json:"name"`

	// Widgets Widgets
	Widgets *[]Widget `json:"widgets,omitempty"`
}

type Widget struct {
	// Height height
	Height *float32 `json:"height,omitempty"`

	// HeightMobile height mobile
	HeightMobile *float32 `json:"height_mobile,omitempty"`

	// Name Name
	Name string `json:"name"`

	// Options Widget options
	Options *map[string]interface{} `json:"options,omitempty"`

	// Type Widget type
	Type string `json:"type"`

	// Variables Widget variables
	Variables *[]WidgetVariable `json:"variables,omitempty"`

	// Width Width
	Width *float32 `json:"width,omitempty"`

	// WidthMobile Width mobile
	WidthMobile *float32 `json:"width_mobile,omitempty"`

	// X x
	X *float32 `json:"x,omitempty"`

	// XMobile x mobile
	XMobile *float32 `json:"x_mobile,omitempty"`

	// Y y
	Y *float32 `json:"y,omitempty"`

	// YMobile y mobile
	YMobile *float32 `json:"y_mobile,omitempty"`
}

type WidgetVariable struct {
	// Name Name
	Name string `json:"name"`

	// Permission Permission
	Permission *string `json:"permission,omitempty"`

	// ThingId Thing ID
	ThingId *string `json:"thing_id,omitempty"`

	// Type Widget type
	Type string `json:"type"`

	// VariableId Variable ID
	VariableId *string `json:"variable_id,omitempty"`
}

type TriggerTemplate struct {
	// Actions Actions
	Actions *[]TriggerAction `json:"actions,omitempty"`

	// Active Active
	Active *bool `json:"active,omitempty"`

	// Description Description
	Description *string `json:"description,omitempty"`

	// Id Template ID
	Id             string          `json:"id"`
	LinkedDevice   *LinkedDevice   `json:"linked_device,omitempty"`
	LinkedProperty *LinkedProperty `json:"linked_property,omitempty"`

	// Name Name
	Name string `json:"name"`
}

type TriggerAction struct {
	// Description Description
	Description *string      `json:"description,omitempty"`
	Email       *EmailAction `json:"email,omitempty"`

	// Kind Kind
	Kind *string `json:"kind,omitempty"`

	// Name Name
	Name             string                  `json:"name"`
	PushNotification *PushNotificationAction `json:"push_notification,omitempty"`
}

type LinkedDevice struct {
	// DeviceConnectedDelay Device connected delay
	DeviceConnectedDelay *float32 `json:"device_connected_delay,omitempty"`

	// DeviceDisconnectedDelay Device disconnected delay
	DeviceDisconnectedDelay *float32 `json:"device_disconnected_delay,omitempty"`

	// ThingId Thing ID
	ThingId string `json:"thing_id"`
}

// LinkedProperty defines model for LinkedProperty.
type LinkedProperty struct {
	// PropertyId Property ID
	PropertyId string `json:"property_id"`

	// ThingId Thing ID
	ThingId string `json:"thing_id"`
}

// NotificationBody defines model for NotificationBody.
type NotificationBody struct {
	// Expression Expression
	Expression string `json:"expression"`

	// Variables Variables
	Variables []BodyVariable `json:"variables"`
}

// PushNotificationAction defines model for PushNotificationAction.
type PushNotificationAction struct {
	Body     NotificationBody `json:"body"`
	Delivery Delivery         `json:"delivery"`
	Title    Subject          `json:"title"`
}

// Subject defines model for Subject.
type Subject struct {
	// Expression expression
	Expression string `json:"expression"`
}

type BodyVariable struct {
	// Attribute attribute
	Attribute string `json:"attribute"`

	// Entity entity
	Entity string `json:"entity"`

	// Placeholder placeholder
	Placeholder string `json:"placeholder"`

	// PropertyId property_id
	PropertyId *string `json:"property_id,omitempty"`

	// ThingId thing_id
	ThingId *string `json:"thing_id,omitempty"`
}

type Delivery struct {
	// Bcc BCC
	Bcc *[]DeliveryTo `json:"bcc,omitempty"`

	// Cc CC
	Cc *[]DeliveryTo `json:"cc,omitempty"`

	// To To
	To []DeliveryTo `json:"to"`
}

// DeliveryTo defines model for DeliveryTo.
type DeliveryTo struct {
	// Email email
	Email *string `json:"email,omitempty"`

	// Username username
	Username *string `json:"username,omitempty"`
}

// EmailAction defines model for EmailAction.
type EmailAction struct {
	Body     NotificationBody `json:"body"`
	Delivery Delivery         `json:"delivery"`
	Subject  Subject          `json:"subject"`
}
