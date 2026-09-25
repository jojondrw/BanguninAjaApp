package regulation

type LookupRequest struct {
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
}

type Response struct {
	ZoneName       string  `json:"zone_name"`
	ZoneCode       string  `json:"zone_code"`
	SubZoneName    string  `json:"sub_zone_name"`
	SubZoneCode    string  `json:"sub_zone_code"`
	District       string  `json:"district,omitempty"`
	Village        string  `json:"village,omitempty"`
	KDB            float64 `json:"kdb"`
	KLB            float64 `json:"klb"`
	KDH            float64 `json:"kdh"`
	GSB            string  `json:"gsb,omitempty"`
	MaxHeight      string  `json:"max_height,omitempty"`
	RegulationNote string  `json:"regulation_note,omitempty"`
	IsSimulated    bool    `json:"is_simulated"`
	Source         string  `json:"source,omitempty"`
}
