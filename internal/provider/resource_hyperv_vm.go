package provider

import (
	"context"
	"fmt"
	hypervapi "terraform-provider-hyperv/internal/hyper-v-api"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &VMResource{}
	_ resource.ResourceWithConfigure   = &VMResource{}
	_ resource.ResourceWithImportState = &VMResource{}
)

type VMResourceModel struct {
	VMId               types.String `tfsdk:"vmid"`
	Name               types.String `tfsdk:"name"` // Adjusted from "vmname" to "name"
	Generation         types.Int64  `tfsdk:"generation"`
	MemoryStartupBytes types.Int64  `tfsdk:"memory_startup"` // Adjusted from "memorystartup"
	Path               types.String `tfsdk:"path"`
	SwitchName         types.String `tfsdk:"switch_name"` // Adjusted from "switchname"
	BootDevice         types.String `tfsdk:"boot_device"` // Adjusted from "bootdevice"
	Prerelease         types.Bool   `tfsdk:"prerelease"`
}

func NewVMResource() resource.Resource {
	return &VMResource{}
}

type VMResource struct {
	client *hypervapi.Client
}

func (r *VMResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *VMResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Debug(ctx, "Hit the schema setup")
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"vmid": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the VM",
				Required:            true,
			},
			"generation": schema.Int64Attribute{
				MarkdownDescription: "VM generation type (e.g., 1 or 2)",
				Required:            true,
			},
			"memory_startup": schema.Int64Attribute{
				MarkdownDescription: "Startup memory allocation in bytes",
				Required:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Path for storing VM files",
				Optional:            true,
			},
			"switch_name": schema.StringAttribute{
				MarkdownDescription: "Switch name for networking",
				Required:            true,
			},
			"boot_device": schema.StringAttribute{
				MarkdownDescription: "Boot device for the VM",
				Optional:            true,
			},
			"prerelease": schema.BoolAttribute{
				MarkdownDescription: "Enable prerelease ",
				Optional:            true,
			},
		},
	}
}

func (r *VMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Custom Error: Hit the Create setup")
	var plan VMResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Custom Error: Got plan data")

	vm := hypervapi.VMModel{
		Name:               plan.Name.String(),
		Generation:         int64(plan.Generation.ValueInt64()),
		MemoryStartupBytes: plan.MemoryStartupBytes.ValueInt64(),
		Path:               plan.Path.ValueString(),
		SwitchName:         plan.SwitchName.ValueString(),
		BootDevice:         plan.BootDevice.ValueString(),
		Prerelease:         plan.Prerelease.ValueBool(),
	}

	// Create command
	vmResults, err := r.client.CreateVM(ctx, vm)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create VM, unexpected error: "+err.Error(),
		)
	}
	tflog.Debug(ctx, "Custom Error: Created VM")

	plan.VMId = types.StringValue(vmResults.VMId)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// tflog.Debug(ctx, "Hit the Read setup")
	// var state VMResourceModel

	// diags := req.State.Get(ctx, &state)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// diags = resp.State.Set(ctx, vm)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }
}

func (r *VMResource) Update(_ context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *VMResource) Delete(_ context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

func (r *VMResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hypervapi.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hypervapi.Client, got %T. Please report this to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *VMResource) ImportState(_ context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

}
