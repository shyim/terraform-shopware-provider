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

var _ resource.Resource = &DeliveryTimeResource{}
var _ resource.ResourceWithImportState = &DeliveryTimeResource{}

func NewSystemConfigResource() resource.Resource {
	return &DeliveryTimeResource{}
}

// DeliveryTimeResource defines the resource implementation.
type DeliveryTimeResource struct {
	client *shopware_sdk.Client
}

// DeliveryTimeModel describes the resource data model.
type DeliveryTimeModel struct {
	Id      types.String  `tfsdk:"id"`
	Name    types.String  `tfsdk:"name"`
	Unit    types.String  `tfsdk:"unit"`
	Minimum types.Float64 `tfsdk:"minimum"`
	Maximum types.Float64 `tfsdk:"maximum"`
}

func (r *DeliveryTimeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_delivery_time"
}

func (r *DeliveryTimeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Delivery Time",

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
			"unit": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Unit of delivery time",
			},
			"minimum": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "Minimum",
			},
			"maximum": schema.Float64Attribute{
				Required:            true,
				MarkdownDescription: "Maximum",
			},
		},
	}
}

func (r *DeliveryTimeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeliveryTimeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeliveryTimeModel

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

func (r *DeliveryTimeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeliveryTimeModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	criteria := shopware_sdk.Criteria{IDs: []string{data.Id.ValueString()}}
	entities, _, err := r.client.Repository.DeliveryTime.Search(shopware_sdk.NewApiContext(ctx), criteria)

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
	data.Unit = types.StringValue(entity.Unit)
	data.Minimum = types.Float64Value(entity.Min)
	data.Maximum = types.Float64Value(entity.Max)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeliveryTimeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeliveryTimeModel

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

func (r *DeliveryTimeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeliveryTimeModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Repository.DeliveryTime.Delete(
		shopware_sdk.NewApiContext(ctx),
		[]string{data.Id.ValueString()},
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
		return
	}
}

func (r *DeliveryTimeResource) upsertData(ctx context.Context, data DeliveryTimeModel) error {
	_, err := r.client.Repository.DeliveryTime.Upsert(
		shopware_sdk.NewApiContext(ctx),
		[]shopware_sdk.DeliveryTime{
			{
				Id:   data.Id.ValueString(),
				Name: data.Name.ValueString(),
				Min:  data.Minimum.ValueFloat64(),
				Max:  data.Maximum.ValueFloat64(),
				Unit: data.Unit.ValueString(),
			},
		},
	)

	return err
}

func (r *DeliveryTimeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
