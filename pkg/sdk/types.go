package sdk

// DNSRecord represents a DNS record in LuciRPC
type DNSRecord struct {
	Type   string `json:".type" validate:"required"`
	IP     string `json:"ip,omitempty"`
	Name   string `json:"name,omitempty"`
	CName  string `json:"cname,omitempty"`
	Target string `json:"target,omitempty"`
}

// PBR represents a Policy Based Routering in LuciRPC
type PBR struct {
	Type      string `json:".type" validate:"required"`
	Name      string `json:"name,omitempty"`
	SrcAddr   string `json:"src_addr,omitempty"`
	Enabled   string `json:"enabled,omitempty"`
	DsrAddr   string `json:"dest_addr,omitempty"`
	Interface string `json:"interface,omitempty"`
}
