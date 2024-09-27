package contract

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/wundergraph/cosmo/terraform-provider-cosmo/internal/api"
	"github.com/wundergraph/cosmo/terraform-provider-cosmo/internal/utils"
)

func NewContractDataSource() datasource.DataSource {
	return &contractDataSource{}
}

type contractDataSource struct {
	client *api.PlatformClient
}

type contractDataSourceModel struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Namespace              types.String `tfsdk:"namespace"`
	Readme                 types.String `tfsdk:"readme"`
	RoutingURL             types.String `tfsdk:"routing_url"`
	AdmissionWebhookUrl    types.String `tfsdk:"admission_webhook_url"`
	AdmissionWebhookSecret types.String `tfsdk:"admission_webhook_secret"`
	LabelMatchers          types.Map    `tfsdk:"label_matchers"`
}

func (d *contractDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contract"
}

func (d *contractDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the federated graph resource, automatically generated by the system.",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the federated graph.",
				Required:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The namespace in which the federated graph is located.",
				Required:            true,
			},
			"readme": schema.StringAttribute{
				MarkdownDescription: "Readme content for the federated graph.",
				Computed:            true,
			},
			"admission_webhook_url": schema.StringAttribute{
				MarkdownDescription: "The URL for the admission webhook that will be triggered during graph operations.",
				Computed:            true,
			},
			"admission_webhook_secret": schema.StringAttribute{
				MarkdownDescription: "The secret token used to authenticate the admission webhook requests.",
				Computed:            true,
				Sensitive:           true,
			},
			"label_matchers": schema.MapAttribute{
				MarkdownDescription: "A list of label matchers used to select the services that will form the federated graph.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"routing_url": schema.StringAttribute{
				MarkdownDescription: "The URL for the federated graph.",
				Computed:            true,
			},
		},
	}
}

func (d *contractDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.PlatformClient)
	if !ok {
		utils.AddDiagnosticError(resp,
			ErrUnexpectedDataSourceType,
			"Expected *api.PlatformClient, got: %T. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = client
}

func (d *contractDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data contractDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Name.IsNull() || data.Name.ValueString() == "" {
		utils.AddDiagnosticError(resp,
			ErrReadingContract,
			"The 'name' attribute is required.",
		)
		return
	}

	namespace := data.Namespace.ValueString()
	if namespace == "" {
		namespace = "default"
	}

	apiResponse, apiError := d.client.GetContract(ctx, data.Name.ValueString(), namespace)
	if apiError != nil {
		utils.AddDiagnosticError(resp,
			ErrReadingContract,
			"Could not read contract: "+apiError.Error(),
		)
		return
	}

	graph := apiResponse.Graph

	data.Id = types.StringValue(graph.GetId())
	data.Name = types.StringValue(graph.GetName())
	data.Namespace = types.StringValue(graph.GetNamespace())
	data.RoutingURL = types.StringValue(graph.GetRoutingURL())

	labelMatchers := make(map[string]attr.Value)
	for _, labelMatcher := range graph.GetLabelMatchers() {
		labelMatchers[labelMatcher] = types.StringValue(labelMatcher)
	}
	data.LabelMatchers = types.MapValueMust(types.StringType, labelMatchers)

	if graph.Readme != nil {
		data.Readme = types.StringValue(*graph.Readme)
	}

	tflog.Trace(ctx, "Read contract data source", map[string]interface{}{
		"id": data.Id.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}