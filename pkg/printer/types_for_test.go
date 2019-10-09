package printer_test

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterProvisioner describes provisioner options
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterProvisioner struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata" yaml:"metadata"`

	Spec ClusterProvisionerSpec `json:"spec" yaml:"spec"`
}

// ClusterProvisionerSpec is the spec that contains the provisioner options
type ClusterProvisionerSpec struct {
	Provider              string                 `json:"provider" yaml:"provider"`
	AWSProviderOptions    *AWSProviderOptions    `json:"aws,omitempty" yaml:"aws,omitempty"`
	DockerProviderOptions *DockerProviderOptions `json:"docker,omitempty" yaml:"docker,omitempty"`
	//AdminCIDRBlocks       []string               `json:"adminCIDRBlocks,omitempty" yaml:"adminCIDRBlocks,omitempty"`
	NodePools      []*MachinePool  `json:"nodePools,omitempty" yaml:"nodePools,omitempty"`
	SSHCredentials *SSHCredentials `json:"sshCredentials,omitempty" yaml:"sshCredentials,omitempty"`
	// not being used right now
	//InventoryPath   string           `json:"inventoryPath,omitempty" yaml:"inventoryPath,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

// MachinePool used by the provisioner to configure a machine
type MachinePool struct {
	Name         string   `json:"name,omitempty" yaml:"name,omitempty"`
	ControlPlane bool     `json:"controlPlane,omitempty" yaml:"controlPlane,omitempty"`
	Bastion      bool     `json:"bastion,omitempty" yaml:"bastion,omitempty"`
	Count        int      `json:"count" yaml:"count"`
	Machine      *Machine `json:"machine,omitempty" yaml:"machine,omitempty"`
}

// Machine used by the provisioner to configure a machine
type Machine struct {
	ImageID              string          `json:"imageID,omitempty" yaml:"imageID,omitempty"`
	ImageName            string          `json:"imageName,omitempty" yaml:"imageName,omitempty"`
	RootVolumeSize       int             `json:"rootVolumeSize,omitempty" yaml:"rootVolumeSize,omitempty"`
	RootVolumeType       string          `json:"rootVolumeType,omitempty" yaml:"rootVolumeType,omitempty"`
	ImagefsVolumeEnabled bool            `json:"imagefsVolumeEnabled,omitempty" yaml:"imagefsVolumeEnabled,omitempty"`
	ImagefsVolumeSize    int             `json:"imagefsVolumeSize,omitempty" yaml:"imagefsVolumeSize,omitempty"`
	ImagefsVolumeType    string          `json:"imagefsVolumeType,omitempty" yaml:"imagefsVolumeType,omitempty"`
	Type                 string          `json:"type,omitempty" yaml:"type,omitempty"`
	AWSMachineOpts       *AWSMachineOpts `json:"aws,omitempty" yaml:"aws,omitempty"`
}

// SSHCredentials describes the options passed to the provisioner regarding the ssh keys
type SSHCredentials struct {
	User           string `json:"user,omitempty" yaml:"user,omitempty"`
	PublicKeyFile  string `json:"publicKeyFile,omitempty" yaml:"publicKeyFile,omitempty"`
	PrivateKeyFile string `json:"privateKeyFile,omitempty" yaml:"privateKeyFile,omitempty"`
}

// DockerProviderOptions describes Docker provider related options
type DockerProviderOptions struct {
	DisablePortMapping            bool `json:"disablePortMapping,omitempty" yaml:"disablePortMapping,omitempty"`
	ControlPlaneMappedPortBase    int  `json:"controlPlaneMappedPortBase,omitempty" yaml:"controlPlaneMappedPortBase"`
	SSHControlPlaneMappedPortBase int  `json:"sshControlPlaneMappedPortBase,omitempty" yaml:"sshControlPlaneMappedPortBase,omitempty"`
	SSHWorkerMappedPortBase       int  `json:"sshWorkerMappedPortBase,omitempty" yaml:"sshWorkerMappedPortBase"`
	DedicatedNetwork              bool `json:"dedicatedNetwork,omitempty" yaml:"dedicatedNetwork,omitempty"`
}

const (
	// PrefixARNInstanceProfile is the prefix in ARNs for instanceProfile
	PrefixARNInstanceProfile = "instance-profile/"
	// ARNSeperator is a specific separator for ARNs
	ARNSeperator = ":"
)

// AWSProviderOptions describes AWS provider related options
type AWSProviderOptions struct {
	Region            string            `json:"region,omitempty" yaml:"region,omitempty"`
	VPC               *VPC              `json:"vpc,omitempty" yaml:"vpc,omitempty"`
	AvailabilityZones []string          `json:"availabilityZones,omitempty" yaml:"availabilityZones,omitempty"`
	ELB               *ELB              `json:"elb,omitempty" yaml:"elb,omitempty"`
	Tags              map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// VPC contains the vpc information required if using an existing vpc
type VPC struct {
	ID                      string `json:"ID,omitempty" yaml:"ID,omitempty"`
	RouteTableID            string `json:"routeTableID,omitempty" yaml:"routeTableID,omitempty"`
	InternetGatewayID       string `json:"internetGatewayID,omitempty" yaml:"internetGatewayID,omitempty"`
	InternetGatewayDisabled bool   `json:"internetGatewayDisabled,omitempty" yaml:"internetGatewayDisabled,omitempty"`
}

// ELB contains details for the kube-apiserver ELB
type ELB struct {
	Internal  bool     `json:"internal,omitempty" yaml:"internal,omitempty"`
	SubnetIDs []string `json:"subnetIDs,omitempty" yaml:"subnetIDs,omitempty"`
}

// AWSMachineOpts is aws specific options for machine
type AWSMachineOpts struct {
	IAM       *IAM     `json:"iam,omitempty" yaml:"iam,omitempty"`
	SubnetIDs []string `json:"subnetIDs,omitempty" yaml:"subnetIDs,omitempty"`
}

// IAM contains role information to use instead of creating one
type IAM struct {
	InstanceProfile *InstanceProfile `json:"instanceProfile,omitempty" yaml:"instanceProfile,omitempty"`
}

type InstanceProfile struct {
	ARN  string `json:"arn,omitempty" yaml:"arn,omitempty"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}
