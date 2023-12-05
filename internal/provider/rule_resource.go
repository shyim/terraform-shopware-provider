package provider

import (
	"context"
	"encoding/json"
	"fmt"
	shopware_sdk "github.com/friendsofshopware/go-shopware-admin-api-sdk"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"strings"
	"terraform-provider-shopware/internal"
)

var _ resource.Resource = &RuleResource{}
var _ resource.ResourceWithImportState = &RuleResource{}

func NewRuleResource() resource.Resource {
	return &RuleResource{}
}

// RuleResource defines the resource implementation.
type RuleResource struct {
	client *shopware_sdk.Client
}

// RuleModel describes the resource data model.
type RuleModel struct {
	Id         types.String       `tfsdk:"id"`
	Name       types.String       `tfsdk:"name"`
	Type       types.List         `tfsdk:"type"`
	Priority   types.Float64      `tfsdk:"priority"`
	Conditions basetypes.SetValue `tfsdk:"conditions"`
}

func (r *RuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rule"
}

func (r *RuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Rule",

		Blocks: map[string]schema.Block{
			"conditions": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Type",
						},
						"value": schema.ObjectAttribute{
							Required: true,
							AttributeTypes: map[string]attr.Type{
								"operator":      types.StringType,
								"affiliateCode": types.StringType,
							},
						},
					},
				},
			},
		},

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name",
			},
			"type": schema.ListAttribute{
				Required:            true,
				MarkdownDescription: "Type",
				CustomType: types.ListType{
					ElemType: types.StringType,
				},
			},
			"priority": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "Priority",
			},
		},
	}
}

func (r *RuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*shopware_sdk.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *RuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue(internal.NewUuid())

	if err := r.upsertData(ctx, data); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	criteria := shopware_sdk.Criteria{IDs: []string{data.Id.ValueString()}}
	entities, _, err := r.client.Repository.Rule.Search(shopware_sdk.NewApiContext(ctx), criteria)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	if entities.Total == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	entity := entities.Data[0]

	data.Name = types.StringValue(entity.Name)
	data.Priority = types.Float64Value(entity.Priority)

	modules := entity.ModuleTypes.(map[string]interface{})["types"].([]interface{})
	tfModules := make([]attr.Value, 0)

	for _, module := range modules {
		tfModules = append(tfModules, types.StringValue(module.(string)))
	}

	moduleTypes, diag := types.ListValue(types.StringType, tfModules)

	if diag.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	data.Type = moduleTypes
	conditions, diag := basetypes.NewSetValue(types.ObjectValue(), []attr.Value{})

	if diag.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	data.Conditions = conditions

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.upsertData(ctx, data); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update delivery time, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RuleModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Repository.ShippingMethod.Delete(
		shopware_sdk.NewApiContext(ctx),
		[]string{data.Id.ValueString()},
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
		return
	}
}

func (r *RuleResource) upsertData(ctx context.Context, data RuleModel) error {
	moduleTypes := make([]string, 0)

	for _, ruleType := range data.Type.Elements() {
		moduleTypes = append(moduleTypes, strings.ReplaceAll(ruleType.String(), "\"", ""))
	}

	var conditions []shopware_sdk.RuleCondition

	if err := json.Unmarshal([]byte(data.Conditions.String()), &conditions); err != nil {
		return err
	}

	_, err := r.client.Repository.Rule.Upsert(
		shopware_sdk.NewApiContext(ctx),
		[]shopware_sdk.Rule{
			{
				Id:          data.Id.ValueString(),
				Name:        data.Name.ValueString(),
				ModuleTypes: map[string]interface{}{"types": moduleTypes},
				Priority:    data.Priority.ValueFloat64(),
				Conditions:  conditions,
			},
		},
	)

	return err
}

func (r *RuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
