package schema

type ClientStatus struct {
	Checksum          string `json:"Checksum"`
	ChecksumAlgorithm string `json:"ChecksumAlgorithm"`
	ConfigurationName string `json:"ConfigurationName"`
}
