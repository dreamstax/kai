package kai

const (
	// GroupName is the group name for Kai labels and annotations
	GroupName = "core.kai.io"
	// StepLabelKey is the label key attached to k8s resources to indicate which  step triggered their creation.
	StepLabelKey = GroupName + "/step"

	// StepUIDLabelKey is the label key attached to k8s resources to indicate which step triggered their creation
	StepUIDLabelKey = GroupName + "/stepUID"

	// PipelineLabelKey is the label key attached to k8s resources to indicate which pipeline triggered their creation.
	PipelineLabelKey = GroupName + "/pipeline"

	// PipelineUIDLabelKey is the label key attached to k8s resources to indicate which pipeline triggerd their creation
	PipelineUIDLabelKey = GroupName + "/pipelineUID"
)
