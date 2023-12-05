package provider

import (
	"context"
	"fmt"
	shopware_sdk "github.com/friendsofshopware/go-shopware-admin-api-sdk"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-shopware/internal"
)

var _ resource.Resource = &ShippingMethodResource{}
var _ resource.ResourceWithImportState = &ShippingMethodResource{}

func NewShippingMethodResource() resource.Resource {
	return &ShippingMethodResource{}
}

// ShippingMethodResource defines the resource implementation.
type ShippingMethodResource struct {
	client *shopware_sdk.Client
}

// ShippingMethodModel describes the resource data model.
type ShippingMethodModel struct {
	Id                 types.String `tfsdk:"id"`
	TechnicalName      types.String `tfsdk:"technical_name"`
	Name               types.String `tfsdk:"name"`
	Active             types.Bool   `tfsdk:"active"`
	DeliveryTimeId     types.String `tfsdk:"delivery_time_id"`
	AvailabilityRuleId types.String `tfsdk:"availability_rule_id"`
}

func (r *ShippingMethodResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shipping_method"
}

func (r *ShippingMethodResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Shipping Method",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Active flag",
			},
			"technical_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name",
			},
			"delivery_time_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Delivery Time ID",
			},
			"availability_rule_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Availability Rule ID",
			},
		},
	}
}

func (r *ShippingMethodResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ShippingMethodResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ShippingMethodModel

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

func (r *ShippingMethodResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ShippingMethodModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	criteria := shopware_sdk.Criteria{IDs: []string{data.Id.ValueString()}}
	entities, _, err := r.client.Repository.ShippingMethod.Search(shopware_sdk.NewApiContext(ctx), criteria)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	if entities.Total == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	entity := entities.Data[0]

	data.TechnicalName = types.StringValue(entity.TechnicalName)
	data.Name = types.StringValue(entity.Name)
	data.Active = types.BoolValue(entity.Active)
	data.DeliveryTimeId = types.StringValue(entity.DeliveryTimeId)
	data.AvailabilityRuleId = types.StringValue(entity.AvailabilityRuleId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ShippingMethodResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ShippingMethodModel

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

func (r *ShippingMethodResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ShippingMethodModel

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

func (r *ShippingMethodResource) upsertData(ctx context.Context, data ShippingMethodModel) error {
	_, err := r.client.Repository.ShippingMethod.Upsert(
		shopware_sdk.NewApiContext(ctx),
		[]shopware_sdk.ShippingMethod{
			{
				Id:                 data.Id.ValueString(),
				Active:             data.Active.ValueBool(),
				Name:               data.Name.ValueString(),
				TechnicalName:      data.TechnicalName.ValueString(),
				DeliveryTimeId:     data.DeliveryTimeId.ValueString(),
				AvailabilityRuleId: data.AvailabilityRuleId.ValueString(),
			},
		},
	)

	return err
}

func (r *ShippingMethodResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
