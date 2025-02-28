package provider

import (
	"context"
	"fmt"
	hypervapi "terraform-provider-hyperv/internal/hyper-v-api"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
	BootDevice         types.String `tfsdk:"boot_device"` // Adjusted from "bootdevice"
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
			"boot_device": schema.StringAttribute{
				MarkdownDescription: "Boot device for the VM",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
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
		BootDevice:         plan.BootDevice.ValueString(),
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
	// plan.Path = types.StringValue(vmResults.Path)

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Hit the Read setup")
	var state VMResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	fmt.Println("Printing state path")
	fmt.Println(state.Path.ValueString())

	vm := hypervapi.VMModel{
		VMId: state.VMId.ValueString(), // Convert types.String to string
	}

	vmResults, err := r.client.GetVM(ctx, vm)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to fetch VM",
			"An unexpected error occurred while fetching the VM state. "+
				"Please report this issue to the provider developers.\n\n"+
				"API Error: "+err.Error(),
		)
		return
	}

	// Convert the API response (VMModel) back to the Terraform state model (VMResourceModel)
	newState := VMResourceModel{
		VMId:               types.StringValue(vmResults.VMId),
		Name:               types.StringValue(vmResults.Name),
		Generation:         types.Int64Value(vmResults.Generation),
		MemoryStartupBytes: types.Int64Value(vmResults.MemoryStartupBytes),
		// TODO: Resolve path issues
		Path:       types.StringValue(state.Path.ValueString()),
		BootDevice: types.StringValue(vmResults.BootDevice),
	}

	fmt.Println("Printing State")
	fmt.Println(state)

	fmt.Println("Printing New State")
	fmt.Println(newState)

	// Save the updated state back to Terraform
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)

}

func (r *VMResource) Update(_ context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Custom Error: Hit the Delete setup")

	var state VMResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm := hypervapi.VMModel{
		VMId: state.VMId.String(),
	}

	err := r.client.DeleteVM(ctx, vm)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not delete VM, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Custom Error: Deleted VM")

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
