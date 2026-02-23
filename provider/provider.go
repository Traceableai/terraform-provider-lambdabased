package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return createProvider(providerConfigure)
}

func createProvider(configureContextFunc schema.ConfigureContextFunc) *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"profile": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"account": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "AWS account ID to assume role in. If not specified, uses default credentials.",
			},
			"assume_role_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "OrganizationAccountAccessRole",
				Description: "Name of the IAM role to assume in the target account. Defaults to OrganizationAccountAccessRole.",
			},
			"assume_role": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"role_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"lambdabased_resource": LambdaBasedResource(),
		},

		ConfigureContextFunc: configureContextFunc,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(d.Get("profile").(string)),
		config.WithRegion(d.Get("region").(string)),
	)

	if err != nil {
		return nil, diag.FromErr(err)
	}

	// If account is specified, assume role in that account
	if account := d.Get("account").(string); account != "" {
		roleName := d.Get("assume_role_name").(string)
		roleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", account, roleName)
		stsSvc := sts.NewFromConfig(cfg)
		creds := stscreds.NewAssumeRoleProvider(stsSvc, roleArn)
		cfg.Credentials = aws.NewCredentialsCache(creds)
	} else if assumeRoleRaw, ok := d.GetOk("assume_role"); ok {
		// Fallback to explicit assume_role if provided
		assumeRole := assumeRoleRaw.([]interface{})[0]
		role := assumeRole.(map[string]interface{})["role_arn"].(string)
		stsSvc := sts.NewFromConfig(cfg)
		creds := stscreds.NewAssumeRoleProvider(stsSvc, role)
		cfg.Credentials = aws.NewCredentialsCache(creds)
	}
	// If neither account nor assume_role is specified, use default credentials

	return lambda.NewFromConfig(cfg), nil
}
