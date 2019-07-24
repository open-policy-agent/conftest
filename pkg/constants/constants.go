package constants

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

const (
	OpenPolicyAgentConfigMediaType        = "application/vnd.cncf.openpolicyagent.config.v1+json"
	OpenPolicyAgentManifestLayerMediaType = "application/vnd.cncf.openpolicyagent.manifest.layer.v1+json"
	OpenPolicyAgentPolicyLayerMediaType   = "application/vnd.cncf.openpolicyagent.policy.layer.v1+rego"
	OpenPolicyAgentDataLayerMediaType     = "application/vnd.cncf.openpolicyagent.data.layer.v1+json"
)
