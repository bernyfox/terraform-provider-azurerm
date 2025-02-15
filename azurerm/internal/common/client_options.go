package common

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/sender"
	"github.com/hashicorp/terraform/httpclient"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/features"
	"github.com/terraform-providers/terraform-provider-azurerm/version"
)

type ClientOptions struct {
	SubscriptionId   string
	TenantID         string
	PartnerId        string
	TerraformVersion string

	GraphAuthorizer           autorest.Authorizer
	GraphEndpoint             string
	KeyVaultAuthorizer        autorest.Authorizer
	ResourceManagerAuthorizer autorest.Authorizer
	ResourceManagerEndpoint   string
	StorageAuthorizer         autorest.Authorizer

	SkipProviderReg             bool
	DisableCorrelationRequestID bool
	Environment                 azure.Environment

	// TODO: remove me in 2.0
	PollingDuration time.Duration
}

func (o ClientOptions) ConfigureClient(c *autorest.Client, authorizer autorest.Authorizer) {
	setUserAgent(c, o.TerraformVersion, o.PartnerId)

	c.Authorizer = authorizer
	c.Sender = sender.BuildSender("AzureRM")
	c.SkipResourceProviderRegistration = o.SkipProviderReg
	if !o.DisableCorrelationRequestID {
		c.RequestInspector = WithCorrelationRequestID(CorrelationRequestID())
	}

	// TODO: remove in 2.0
	if !features.SupportsCustomTimeouts() {
		c.PollingDuration = o.PollingDuration
	}
}

func setUserAgent(client *autorest.Client, tfVersion, partnerID string) {
	tfUserAgent := httpclient.TerraformUserAgent(tfVersion)

	providerUserAgent := fmt.Sprintf("%s terraform-provider-azurerm/%s", tfUserAgent, version.ProviderVersion)
	client.UserAgent = strings.TrimSpace(fmt.Sprintf("%s %s", client.UserAgent, providerUserAgent))

	// append the CloudShell version to the user agent if it exists
	if azureAgent := os.Getenv("AZURE_HTTP_USER_AGENT"); azureAgent != "" {
		client.UserAgent = fmt.Sprintf("%s %s", client.UserAgent, azureAgent)
	}

	if partnerID != "" {
		client.UserAgent = fmt.Sprintf("%s pid-%s", client.UserAgent, partnerID)
	}

	log.Printf("[DEBUG] AzureRM Client User Agent: %s\n", client.UserAgent)
}
