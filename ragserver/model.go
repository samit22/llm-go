package ragserver

type GraphQLResponse struct {
	Get struct {
		Document []struct {
			Text string `json:"text"`
		} `json:"Document"`
	} `json:"Get"`
}
