package uasset

// UAssetResult contains the outcome of a UAsset export or import operation.
// The FilesProcessed count is extracted from the bridge output if available.
// This type is shared between IPC and Native implementations.
type UAssetResult struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	Output         string `json:"output"`
	Error          string `json:"error"`
	Duration       string `json:"duration"`
	FilesProcessed int    `json:"files_processed"`
}
