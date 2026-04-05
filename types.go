package main

// SendSmsRequest is the request body for POST /api/sms/send.
type SendSmsRequest struct {
	Recipients []string `json:"recipients"`
	Body       string   `json:"body"`
	DeviceID   string   `json:"device_id,omitempty"`
}

// SendSmsTemplateRequest is the request body for POST /api/sms/send-template.
type SendSmsTemplateRequest struct {
	Recipients []string          `json:"recipients"`
	TemplateID string            `json:"template_id"`
	Variables  map[string]string `json:"variables,omitempty"`
	DeviceID   string            `json:"device_id,omitempty"`
}

// SendSmsResponse is the response from POST /api/sms/send.
type SendSmsResponse struct {
	BatchID         string   `json:"batch_id,omitempty"`
	MessageIDs      []string `json:"message_ids"`
	RecipientsCount int      `json:"recipients_count"`
	Status          string   `json:"status"`
}

// QuotaResponse is the response from GET /api/plans/quota.
type QuotaResponse struct {
	Plan                string `json:"plan"`
	SmsSentThisMonth    int    `json:"sms_sent_this_month"`
	MaxSmsPerMonth      int    `json:"max_sms_per_month"`
	DevicesRegistered   int    `json:"devices_registered"`
	MaxDevices          int    `json:"max_devices"`
	ResetDate           string `json:"reset_date"`
	ScheduledSmsActive  int    `json:"scheduled_sms_active"`
	MaxScheduledSms     int    `json:"max_scheduled_sms"`
	IntegrationsCreated int    `json:"integrations_created"`
	MaxIntegrations     int    `json:"max_integrations"`
}

// PaginatedResponse is the standard PocketBase paginated list response.
type PaginatedResponse[T any] struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalPages int `json:"totalPages"`
	TotalItems int `json:"totalItems"`
	Items      []T `json:"items"`
}

// SmsMessage represents a record in the sms_messages collection.
type SmsMessage struct {
	ID           string `json:"id"`
	To           string `json:"to"`
	FromNumber   string `json:"from_number"`
	Body         string `json:"body"`
	Status       string `json:"status"`
	MessageType  string `json:"message_type"`
	BatchID      string `json:"batch_id"`
	Device       string `json:"device"`
	ErrorMessage string `json:"error_message"`
	SentAt       string `json:"sent_at"`
	DeliveredAt  string `json:"delivered_at"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
}

// SmsDevice represents a record in the sms_devices collection.
type SmsDevice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	DeviceType  string `json:"device_type"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

// SmsTemplate represents a record in the sms_templates collection.
type SmsTemplate struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// ScheduledSms represents a record in the scheduled_sms collection.
type ScheduledSms struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Recipients     []string `json:"recipients"`
	Body           string   `json:"body"`
	DeviceID       string   `json:"device_id"`
	ScheduleType   string   `json:"schedule_type"`
	ScheduledAt    string   `json:"scheduled_at"`
	CronExpression string   `json:"cron_expression"`
	Timezone       string   `json:"timezone"`
	NextRunAt      string   `json:"next_run_at"`
	LastRunAt      string   `json:"last_run_at"`
	Status         string   `json:"status"`
	Created        string   `json:"created"`
	Updated        string   `json:"updated"`
}
