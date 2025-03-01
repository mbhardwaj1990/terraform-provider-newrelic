package newrelic

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/newrelic/newrelic-client-go/pkg/errors"
	"github.com/newrelic/newrelic-client-go/pkg/synthetics"
)

func resourceNewRelicSyntheticsMonitor() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNewRelicSyntheticsMonitorCreate,
		ReadContext:   resourceNewRelicSyntheticsMonitorRead,
		UpdateContext: resourceNewRelicSyntheticsMonitorUpdate,
		DeleteContext: resourceNewRelicSyntheticsMonitorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The monitor type. Valid values are SIMPLE, BROWSER, SCRIPT_BROWSER, and SCRIPT_API.",
				ValidateFunc: validation.StringInSlice([]string{
					"SIMPLE",
					"BROWSER",
					"SCRIPT_API",
					"SCRIPT_BROWSER",
					"CERT_CHECK",
				}, false),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The title of this monitor.",
			},
			"frequency": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: intInSlice([]int{1, 5, 10, 15, 30, 60, 360, 720, 1440}),
				Description:  "The interval (in minutes) at which this monitor should run. Valid values are 1, 5, 10, 15, 30, 60, 360, 720, or 1440.",
			},
			"uri": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The URI for the monitor to hit.",
				// TODO: ValidateFunc (required if SIMPLE or BROWSER)
			},
			"locations": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
				Required:    true,
				Description: "The locations in which this monitor should be run.",
			},
			"status": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The monitor status (i.e. ENABLED, MUTED, DISABLED).",
				ValidateFunc: validation.StringInSlice([]string{
					"ENABLED",
					"MUTED",
					"DISABLED",
				}, false),
			},
			"sla_threshold": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Default:     7,
				Description: "The base threshold (in seconds) to calculate the apdex score for use in the SLA report. (Default 7 seconds)",
			},
			// TODO: ValidationFunc (options only valid if SIMPLE or BROWSER)
			"validation_string": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The string to validate against in the response.",
			},
			"verify_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Verify SSL.",
			},
			"bypass_head_request": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Bypass HEAD request.",
			},
			"treat_redirect_as_failure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Fail the monitor check if redirected.",
			},
		},
	}
}

func buildSyntheticsMonitorStruct(d *schema.ResourceData) synthetics.Monitor {
	monitor := synthetics.Monitor{
		Name:         d.Get("name").(string),
		Type:         synthetics.MonitorType(d.Get("type").(string)),
		Frequency:    uint(d.Get("frequency").(int)),
		Status:       synthetics.MonitorStatusType(d.Get("status").(string)),
		SLAThreshold: d.Get("sla_threshold").(float64),
	}

	if uri, ok := d.GetOk("uri"); ok {
		monitor.URI = uri.(string)
	}

	locationsRaw := d.Get("locations").(*schema.Set)
	locations := make([]string, locationsRaw.Len())
	for i, v := range locationsRaw.List() {
		locations[i] = fmt.Sprint(v)
	}

	if validationString, ok := d.GetOk("validation_string"); ok {
		monitor.Options.ValidationString = validationString.(string)
	}

	if verifySSL, ok := d.GetOkExists("verify_ssl"); ok {
		monitor.Options.VerifySSL = verifySSL.(bool)
	}

	if bypassHeadRequest, ok := d.GetOkExists("bypass_head_request"); ok {
		monitor.Options.BypassHEADRequest = bypassHeadRequest.(bool)
	}

	if treatRedirectAsFailure, ok := d.GetOkExists("treat_redirect_as_failure"); ok {
		monitor.Options.TreatRedirectAsFailure = treatRedirectAsFailure.(bool)
	}

	monitor.Locations = locations
	return monitor
}

func buildSyntheticsUpdateMonitorArgs(d *schema.ResourceData) *synthetics.Monitor {
	monitor := synthetics.Monitor{
		ID:           d.Id(),
		Name:         d.Get("name").(string),
		Type:         synthetics.MonitorType(d.Get("type").(string)),
		Frequency:    uint(d.Get("frequency").(int)),
		Status:       synthetics.MonitorStatusType(d.Get("status").(string)),
		SLAThreshold: d.Get("sla_threshold").(float64),
	}

	if uri, ok := d.GetOk("uri"); ok {
		monitor.URI = uri.(string)
	}

	locationsRaw := d.Get("locations").(*schema.Set)
	locations := make([]string, locationsRaw.Len())
	for i, v := range locationsRaw.List() {
		locations[i] = fmt.Sprint(v)
	}

	if validationString, ok := d.GetOk("validation_string"); ok {
		monitor.Options.ValidationString = validationString.(string)
	}

	if verifySSL, ok := d.GetOkExists("verify_ssl"); ok {
		monitor.Options.VerifySSL = verifySSL.(bool)
	}

	if bypassHeadRequest, ok := d.GetOkExists("bypass_head_request"); ok {
		monitor.Options.BypassHEADRequest = bypassHeadRequest.(bool)
	}

	if treatRedirectAsFailure, ok := d.GetOkExists("treat_redirect_as_failure"); ok {
		monitor.Options.TreatRedirectAsFailure = treatRedirectAsFailure.(bool)
	}

	monitor.Locations = locations
	return &monitor
}

func readSyntheticsMonitorStruct(monitor *synthetics.Monitor, d *schema.ResourceData) {
	_ = d.Set("name", monitor.Name)
	_ = d.Set("type", monitor.Type)
	_ = d.Set("frequency", monitor.Frequency)
	_ = d.Set("uri", monitor.URI)
	_ = d.Set("locations", monitor.Locations)
	_ = d.Set("status", monitor.Status)
	_ = d.Set("sla_threshold", monitor.SLAThreshold)
	_ = d.Set("verify_ssl", monitor.Options.VerifySSL)
	_ = d.Set("validation_string", monitor.Options.ValidationString)
	_ = d.Set("bypass_head_request", monitor.Options.BypassHEADRequest)
	_ = d.Set("treat_redirect_as_failure", monitor.Options.TreatRedirectAsFailure)
}

func resourceNewRelicSyntheticsMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfig).NewClient
	monitorStruct := buildSyntheticsMonitorStruct(d)

	log.Printf("[INFO] Creating New Relic Synthetics monitor %s", monitorStruct.Name)

	monitor, err := client.Synthetics.CreateMonitorWithContext(ctx, monitorStruct)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(monitor.ID)
	return resourceNewRelicSyntheticsMonitorRead(ctx, d, meta)
}

func resourceNewRelicSyntheticsMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfig).NewClient

	log.Printf("[INFO] Reading New Relic Synthetics monitor %s", d.Id())

	monitor, err := client.Synthetics.GetMonitorWithContext(ctx, d.Id())
	if err != nil {
		if _, ok := err.(*errors.NotFound); ok {
			d.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	readSyntheticsMonitorStruct(monitor, d)

	return nil
}

func resourceNewRelicSyntheticsMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfig).NewClient
	log.Printf("[INFO] Updating New Relic Synthetics monitor %s", d.Id())

	_, err := client.Synthetics.UpdateMonitorWithContext(ctx, *buildSyntheticsUpdateMonitorArgs(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceNewRelicSyntheticsMonitorRead(ctx, d, meta)
}

func resourceNewRelicSyntheticsMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ProviderConfig).NewClient

	log.Printf("[INFO] Deleting New Relic Synthetics monitor %s", d.Id())

	if err := client.Synthetics.DeleteMonitorWithContext(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
