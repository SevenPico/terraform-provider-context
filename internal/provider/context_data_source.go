package provider

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Model struct {
	Enabled       types.Bool              `tfsdk:"enabled"`
	Attributes    []types.String          `tfsdk:"attributes"`
	AttributesMap map[string]types.String `tfsdk:"attributes_map"`
	Descriptors   map[string]Descriptor   `tfsdk:"descriptors"`
}

type Descriptor struct {
	Order     []types.String `tfsdk:"order"`
	Delimiter types.String   `tfsdk:"delimiter"`
	Upper     types.Bool     `tfsdk:"upper"`
	Lower     types.Bool     `tfsdk:"lower"`
	Title     types.Bool     `tfsdk:"title"`
	Reverse   types.Bool     `tfsdk:"reverse"`
	Limit     types.Int64    `tfsdk:"limit"`
	Value     types.String   `tfsdk:"value"`
}

type ContextDataSource struct{}

var _ datasource.DataSource = &ContextDataSource{}

func NewContextDataSource() datasource.DataSource {
	return &ContextDataSource{}
}

func (*ContextDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "context"
}

func (*ContextDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	descriptorAttributes := map[string]schema.Attribute{
		"order": schema.ListAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			MarkdownDescription: "Ordered list of keys in `tags` for which the corresponding value should be included in the output.\nDefault: `[]`",
		},
		"delimiter": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "String to separate `tags` values specified in `order`.\nDefault: \"\"",
		},
		"lower": schema.BoolAttribute{
			Optional:            true,
			MarkdownDescription: "Set `true` to force output to lower-case",
		},
		"upper": schema.BoolAttribute{
			MarkdownDescription: "Set `true` to force output to upper-case",
			Optional:            true,
		},
		"title": schema.BoolAttribute{
			MarkdownDescription: "Set `true` to force output to title-case",
			Optional:            true,
		},
		"reverse": schema.BoolAttribute{
			MarkdownDescription: "Set `true` to reverse the order of `tags` values specified in `order`",
			Optional:            true,
		},
		"limit": schema.Int64Attribute{
			MarkdownDescription: "Character limit of output. Tail characters are trimmed.",
			Optional:            true,
		},
		"value": schema.StringAttribute{
			MarkdownDescription: "Character limit of output. Tail characters are trimmed.",
			Optional:            true,
			Computed:            true,
		},
	}

	// id, prefix, dns_name, title
	partialAttributes := map[string]schema.Attribute{
		"enabled": schema.BoolAttribute{
			MarkdownDescription: "Set `true` if resources that use this context should be created.",
			Optional:            true,
		},
		"attributes": schema.ListAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "TODO",
			Optional:            true,
		},
		"attributes_map": schema.MapAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "TODO",
			Optional:            true,
		},
		"descriptors": schema.MapNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: descriptorAttributes,
			},
			Optional:            true,
			MarkdownDescription: "TODO",
		},

		// Outputs
		"id": schema.StringAttribute{
			MarkdownDescription: "TODO",
			Computed:            true,
			Optional:            true,
		},
		"namespace": schema.StringAttribute{
			MarkdownDescription: "TODO",
			Computed:            true,
			Optional:            true,
		},
	}

	attributes := map[string]schema.Attribute{}

	for k, v := range partialAttributes {
		attributes[k] = v
	}

	attributes["context"] = schema.SingleNestedAttribute{
		MarkdownDescription: "TODO",
		Attributes:          partialAttributes,
		Optional:            true,
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Context Data Source",
		Attributes:          attributes,
	}
}

func (*ContextDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var child Model
	var parent Model

	{ // Merge Enabled
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("enabled"), &child.Enabled)...)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("context").AtName("enabled"), &parent.Enabled)...)

		if child.Enabled.IsUnknown() || child.Enabled.IsNull() {
			child.Enabled = types.BoolValue(true)
		}

		if parent.Enabled.IsUnknown() || parent.Enabled.IsNull() {
			parent.Enabled = types.BoolValue(true)
		}

		if child.Enabled.ValueBool() && parent.Enabled.ValueBool() {
			child.Enabled = types.BoolValue(true)
		} else {
			child.Enabled = types.BoolValue(false)
		}
	}

	{ // Merge Attributes
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("attributes"), &child.Attributes)...)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("context").AtName("attributes"), &parent.Attributes)...)

		child.Attributes = append(parent.Attributes, child.Attributes...)
	}

	{ // Merge Attributes Map
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("attributes_map"), &child.AttributesMap)...)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("context").AtName("attributes_map"), &parent.AttributesMap)...)

		if child.AttributesMap == nil {
			child.AttributesMap = make(map[string]types.String)
		}
		if parent.AttributesMap == nil {
			parent.AttributesMap = make(map[string]types.String)
		}

		for k, v := range child.AttributesMap {
			parent.AttributesMap[k] = v
		}
		child.AttributesMap = parent.AttributesMap
	}

	{ // Descriptors
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("descriptors"), &child.Descriptors)...)
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("context").AtName("descriptors"), &parent.Descriptors)...)

		if child.Descriptors == nil {
			child.Descriptors = make(map[string]Descriptor)
		}
		if parent.Descriptors == nil {
			parent.Descriptors = make(map[string]Descriptor)
		}

		// Merge
		for k, v := range child.Descriptors {
			parent.Descriptors[k] = v
		}
		child.Descriptors = parent.Descriptors

		// Defaults for first-class descriptors
		if _, ok := child.Descriptors["id"]; !ok {
			child.Descriptors["id"] = Descriptor{
				Delimiter: types.StringValue("-"),
				Upper:     types.BoolValue(false),
				Lower:     types.BoolValue(false),
				Title:     types.BoolValue(false),
				Reverse:   types.BoolValue(false),
				Limit:     types.Int64Value(64),
				Order:     []types.String{types.StringValue("attributes")},
			}
		}
		if _, ok := child.Descriptors["namespace"]; !ok {
			child.Descriptors["namespace"] = Descriptor{
				Delimiter: types.StringValue("-"),
				Upper:     types.BoolValue(false),
				Lower:     types.BoolValue(false),
				Title:     types.BoolValue(false),
				Reverse:   types.BoolValue(false),
				Limit:     types.Int64Value(math.MaxInt64),
				Order:     []types.String{},
			}
		}

		// Compute
		attributesMap := map[string]string{}
		attributes := []string{}

		for k, v := range child.AttributesMap {
			attributesMap[k] = v.ValueString()
		}

		for _, v := range child.Attributes {
			attributes = append(attributes, v.ValueString())
		}

		for k, d := range child.Descriptors {
			resp.Diagnostics.Append(d.Compute(attributesMap, attributes)...)
			child.Descriptors[k] = d
		}

		// Expand descriptors values
		for pk, pd := range child.Descriptors {
			newValue := pd.Value.ValueString()
			for sk, sd := range child.Descriptors {
				newValue = strings.ReplaceAll(newValue, fmt.Sprintf("${%s}", sk), sd.Value.ValueString())
				newValue = strings.Trim(newValue, pd.Delimiter.ValueString())
			}
			pd.Value = types.StringValue(newValue)
			child.Descriptors[pk] = pd
		}
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("enabled"), child.Enabled)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("attributes"), child.Attributes)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("attributes_map"), child.AttributesMap)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("descriptors"), child.Descriptors)...)

	// First-class descriptors
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), child.Descriptors["id"].Value)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace"), child.Descriptors["namespace"].Value)...)
}

func (d *Descriptor) Compute(attributesMap map[string]string, attributes []string) diag.Diagnostics {
	var diags diag.Diagnostics
	var parts []string

	for _, k := range d.Order {
		key := k.ValueString()

		if val, ok := attributesMap[key]; ok {
			parts = append(parts, val)
		} else if key == "attributes" {
			parts = append(parts, attributes...)
		} else if strings.HasPrefix(key, "$") {
			parts = append(parts, fmt.Sprintf("${%s}", strings.TrimPrefix(key, "$")))
		} else {
			fmt.Println("warn - no key:", key) // FIXME - diags
		}
	}

	var partsClean []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "$") {
			partsClean = append(partsClean, p)
			continue
		}

		if d.Lower.ValueBool() {
			p = strings.ToLower(p)
		}

		if d.Upper.ValueBool() {
			p = strings.ToUpper(p)
		}

		if d.Title.ValueBool() {
			p = strings.Title(p)
		}

		partsClean = append(partsClean, p)
	}

	if d.Reverse.ValueBool() {
		for i, j := 0, len(partsClean)-1; i < j; i, j = i+1, j-1 {
			partsClean[i], partsClean[j] = partsClean[j], partsClean[i]
		}
	}

	value := strings.Join(partsClean, d.Delimiter.ValueString())

	limit := d.Limit.ValueInt64()
	if !d.Limit.IsNull() && !d.Limit.IsUnknown() && int64(len(value)) > d.Limit.ValueInt64() {
		diags.AddWarning("Descriptor length exceeds limit. Output will be trimmed.", fmt.Sprintf("Limit = %d, Length = %d, Value = %s", limit, len(value), value))
		value = value[:limit]
	}

	d.Value = types.StringValue(value)
	return diags
}
