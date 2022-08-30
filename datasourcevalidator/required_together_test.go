package datasourcevalidator_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequiredTogether(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		pathExpressions path.Expressions
		req             datasource.ValidateConfigRequest
		expected        *datasource.ValidateConfigResponse
	}{
		"no-diagnostics": {
			pathExpressions: path.Expressions{
				path.MatchRoot("test"),
			},
			req: datasource.ValidateConfigRequest{
				Config: tfsdk.Config{
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test": {
								Optional: true,
								Type:     types.StringType,
							},
							"other": {
								Optional: true,
								Type:     types.StringType,
							},
						},
					},
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test":  tftypes.String,
								"other": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"test":  tftypes.NewValue(tftypes.String, "test-value"),
							"other": tftypes.NewValue(tftypes.String, "test-value"),
						},
					),
				},
			},
			expected: &datasource.ValidateConfigResponse{},
		},
		"diagnostics": {
			pathExpressions: path.Expressions{
				path.MatchRoot("test1"),
				path.MatchRoot("test2"),
			},
			req: datasource.ValidateConfigRequest{
				Config: tfsdk.Config{
					Schema: tfsdk.Schema{
						Attributes: map[string]tfsdk.Attribute{
							"test1": {
								Optional: true,
								Type:     types.StringType,
							},
							"test2": {
								Optional: true,
								Type:     types.StringType,
							},
							"other": {
								Optional: true,
								Type:     types.StringType,
							},
						},
					},
					Raw: tftypes.NewValue(
						tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"test1": tftypes.String,
								"test2": tftypes.String,
								"other": tftypes.String,
							},
						},
						map[string]tftypes.Value{
							"test1": tftypes.NewValue(tftypes.String, "test-value"),
							"test2": tftypes.NewValue(tftypes.String, nil),
							"other": tftypes.NewValue(tftypes.String, "test-value"),
						},
					),
				},
			},
			expected: &datasource.ValidateConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewAttributeErrorDiagnostic(
						path.Root("test1"),
						"Invalid Attribute Combination",
						"These attributes must be configured together: [test1,test2]",
					),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			validator := datasourcevalidator.RequiredTogether(testCase.pathExpressions...)
			got := &datasource.ValidateConfigResponse{}

			validator.ValidateDataSource(context.Background(), testCase.req, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}